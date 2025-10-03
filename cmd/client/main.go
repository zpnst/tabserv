package main

import (
	"flag"
	"log"
	"sync"

	pb "github.com/zpnst/tabserv/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:4400", "Address to connect to")
)

func main() {
	flag.Parse()
	log.SetPrefix("[iptables client] :: ")

	// Соединение
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewIPTablesClient(conn)

	wg := sync.WaitGroup{}

	// Listening events
	//
	// //
	wg.Add(1)
	go func() {
		defer wg.Done()
		ListenEvents(client, &pb.ListenRequest{
			Type: "tcp",
		})
	}()

	// Add and Delete
	//
	// //
	AddDropTcp(client, &pb.AddDropTcpRequest{
		SourceIp: "192.168.1.100",
		Dport:    22,
	})

	DeleteDropTcp(client, &pb.DeleteDropTcpRequest{
		SourceIp: "192.168.1.100",
		Dport:    22,
	})

	// Add two rules
	//
	// //
	AddDropTcp(client, &pb.AddDropTcpRequest{
		SourceIp: "10.0.0.1",
		Dport:    8080,
	})

	AddDropTcp(client, &pb.AddDropTcpRequest{
		SourceIp: "127.0.0.1",
		Dport:    8080,
	})

	// List Input
	//
	// //
	ListInput(client, &pb.ListInputRequest{})

	// Delte one rule
	//
	// //
	DeleteDropTcp(client, &pb.DeleteDropTcpRequest{
		SourceIp: "10.0.0.1",
		Dport:    8080,
	})

	// Getting History
	//
	// //
	GetHistory(client, &pb.ListInputRequest{})

	wg.Wait()
}
