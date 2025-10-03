package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-iptables/iptables"
	"github.com/zpnst/tabserv/internal/database"
	pb "github.com/zpnst/tabserv/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	table = "filter"
	chain = "INPUT"
)

type Sub struct {
	id        int
	eventChan chan *pb.ListenEvent
}

type IPTServer struct {
	pb.UnimplementedIPTablesServer
	db database.Databse

	ipt *iptables.IPTables

	subs     map[string]map[int]*Sub
	subsLock sync.RWMutex

	nextID int
}

func NewIPTServer(db database.Databse) *IPTServer {
	ipt, err := iptables.New()
	if err != nil {
		panic(err)
	}
	return &IPTServer{
		ipt:  ipt,
		db:   db,
		subs: make(map[string]map[int]*Sub),
	}
}

func (kvs *IPTServer) AddEvent(t pb.ListenEvent_Type, iptt string, rf string) {
	event := &pb.ListenEvent{
		Type:    t,
		Rulefmt: rf,
	}

	kvs.subsLock.RLock()
	defer kvs.subsLock.RUnlock()

	m := kvs.subs[iptt]
	if m == nil {
		return
	}
	for _, sub := range m {
		select {
		case sub.eventChan <- event:
		default:
			// log.Println("slow client")
		}
	}
}

func rulefmt(sourceIP string, dport uint32) []string {
	return []string{"-p", "tcp", "-s", sourceIP, "--dport", fmt.Sprint(dport), "-j", "DROP"}
}

func (s *IPTServer) AddDropTcp(ctx context.Context, req *pb.AddDropTcpRequest) (*pb.AddDropTcpResponse, error) {
	rf := rulefmt(req.SourceIp, req.Dport)

	exists, err := s.ipt.Exists(table, chain, rf...)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to check existence")
	}

	if exists {
		return nil, status.Errorf(codes.AlreadyExists, "rule already exists")
	}

	if err := s.ipt.Append(table, chain, rf...); err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to append ruke: %v", err)
	}

	if err := s.db.Put(ctx, strings.Join(rf, " "), []byte(time.Now().String())); err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to append rule to history: %v", err)
	}

	resp := &pb.AddDropTcpResponse{
		Added:   !exists,
		Rulefmt: fmt.Sprintf("-A %s %v", chain, rf),
	}

	s.AddEvent(pb.ListenEvent_ADD, "tcp", strings.Join(rf, " "))
	return resp, nil
}

func (s *IPTServer) DeleteDropTcp(ctx context.Context, req *pb.DeleteDropTcpRequest) (*pb.DeleteDropTcpResponse, error) {
	rf := rulefmt(req.SourceIp, req.Dport)

	exists, err := s.ipt.Exists(table, chain, rf...)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to check existence")
	}

	if !exists {
		return nil, status.Errorf(codes.NotFound, "rule not found")
	}

	deleted := false
	if exists {
		if err := s.ipt.Delete(table, chain, rf...); err != nil {
			return nil, status.Errorf(codes.Unknown, "failed to delete rule: %v", err)
		}
		deleted = true
	}

	resp := &pb.DeleteDropTcpResponse{
		Deleted: deleted,
		Rulefmt: fmt.Sprintf("-D %s %v", chain, rf),
	}

	s.AddEvent(pb.ListenEvent_DELETE, "tcp", strings.Join(rf, " "))
	return resp, nil
}

func (s *IPTServer) ListInput(ctx context.Context, _ *pb.ListInputRequest) (*pb.ListInputResponse, error) {
	currRules, err := s.db.All(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to read db: %v", err)
	}

	rules := make([]string, 0, len(currRules))

	for ruleKey := range currRules {
		parts := strings.Fields(ruleKey)
		if len(parts) == 0 {
			continue
		}
		exists, err := s.ipt.Exists(table, chain, parts...)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "exists check failed: %v", err)
		}
		if !exists {
			continue
		}

		rules = append(rules, fmt.Sprintf("-A %s %s", chain, ruleKey))
	}

	resp := &pb.ListInputResponse{
		Lines: rules,
	}

	return resp, nil
}

func (s *IPTServer) GetHistory(req *pb.ListInputRequest, stream pb.IPTables_GetHistoryServer) error {
	m, err := s.db.All(stream.Context())
	if err != nil {
		return status.Errorf(codes.Internal, "failed to load history: %v", err)
	}
	for rule := range m {
		resp := &pb.ListInputResponse{
			Lines: []string{rule},
		}
		if err := stream.Send(resp); err != nil {
			return status.Errorf(codes.Unavailable, "failed stream send: %v", err)
		}
	}
	return nil
}

func (kvs *IPTServer) Listen(req *pb.ListenRequest, stream pb.IPTables_ListenServer) error {
	key := req.Type

	kvs.subsLock.Lock()
	kvs.nextID += 1
	id := kvs.nextID

	sub := Sub{
		id:        id,
		eventChan: make(chan *pb.ListenEvent, 32),
	}
	if kvs.subs[key] == nil {
		kvs.subs[key] = make(map[int]*Sub)
	}
	kvs.subs[key][id] = &sub
	kvs.subsLock.Unlock()

	defer func() {
		kvs.subsLock.Lock()
		if keys := kvs.subs[key]; keys != nil {
			if _, ok := keys[id]; ok {
				delete(keys, id)
				close(sub.eventChan)
				if len(keys) == 0 {
					delete(kvs.subs, key)
				}
			}
		}
		kvs.subsLock.Unlock()
	}()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case e := <-sub.eventChan:
			if err := stream.Send(e); err != nil {
				return status.Errorf(codes.Unavailable, "failed stream send: %v", err)
			}

		}
	}
}
