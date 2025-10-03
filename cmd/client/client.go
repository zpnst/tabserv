package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/zpnst/tabserv/proto/gen"
)

var (
	BaseTimeout = 10 * time.Second
)

func AddDropTcp(client pb.IPTablesClient, req *pb.AddDropTcpRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), BaseTimeout)
	defer cancel()
	resp, err := client.AddDropTcp(ctx, req)
	if err != nil {
		log.Printf("[ADD DROP TCP]    :: error    :: %+v\n", err)
		return
	}
	log.Printf("[ADD DROP TCP]     :: response :: added=%v rule=%s\n", resp.Added, resp.Rulefmt)
}

func DeleteDropTcp(client pb.IPTablesClient, req *pb.DeleteDropTcpRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), BaseTimeout)
	defer cancel()
	resp, err := client.DeleteDropTcp(ctx, req)
	if err != nil {
		log.Printf("[DELETE DROP TCP] :: error    :: %+v\n", err)
		return
	}
	log.Printf("[DELETE DROP TCP] :: response :: deleted=%v rule=%s\n", resp.Deleted, resp.Rulefmt)
}

func ListInput(client pb.IPTablesClient, req *pb.ListInputRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), BaseTimeout)
	defer cancel()
	resp, err := client.ListInput(ctx, req)
	if err != nil {
		log.Printf("[LIST INPUT]      :: error    :: %+v\n", err)
		return
	}
	log.Printf("[LIST INPUTs]:\n")
	for i, line := range resp.Lines {
		log.Printf("	rule[%d]    :: %s\n", i, line)
	}
	log.Printf("[LIST INPUTs]:    :: ok\n")
}

func GetHistory(client pb.IPTablesClient, req *pb.ListInputRequest) {
	ctx := context.Background()
	stream, err := client.GetHistory(ctx, req)
	if err != nil {
		log.Printf("[GET HISTORY]     :: error    :: %+v\n", err)
		return
	}

	log.Printf("[GET HISTORY]:\n")
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			log.Printf("[GET HISTORY]:   :: stream   :: completed\n")
			break
		}
		if err != nil {
			log.Printf("[GET HISTORY]     :: error    :: %+v\n", err)
			return
		}
		for _, line := range msg.Lines {
			log.Printf("       entry     :: %s\n", line)
		}
	}
}

func ListenEvents(client pb.IPTablesClient, req *pb.ListenRequest) {
	ctx := context.Background()
	stream, err := client.Listen(ctx, req)
	if err != nil {
		log.Printf("[LISTEN]      :: error    :: %+v\n", err)
		return
	}
	for {
		event, err := stream.Recv()
		if err == io.EOF {
			log.Printf("[LISTEN]      :: stream completed\n")
			break
		}
		if err != nil {
			log.Printf("[LISTEN]      :: error    :: %+v\n", err)
			return
		}
		log.Printf("[LISTEN]          :: event    :: type: %s, rulefmt: %+v\n", event.Type.String(), event.Rulefmt)
	}
}
