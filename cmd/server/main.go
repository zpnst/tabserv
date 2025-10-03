package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/zpnst/tabserv/internal/database"
	"github.com/zpnst/tabserv/internal/ioc"
	pb "github.com/zpnst/tabserv/proto/gen"
	"google.golang.org/grpc"
)

var port = flag.Int("port", 4400, "server port")

func main() {
	flag.Parse()
	log.SetPrefix("[iptables server] :: ")

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Server is running on :%d\n", *port)

	grpcServer := grpc.NewServer()

	c := buildCont()
	s := ioc.Resolve[*IPTServer](c)

	pb.RegisterIPTablesServer(grpcServer, s)
	grpcServer.Serve(lis)
}

func buildCont() *ioc.Container {
	c := ioc.NewContainer()

	ioc.RegisterSingleton[database.Databse](c, func(_ *ioc.Container) database.Databse {
		return database.NewDatabseInMemory()
	})

	ioc.Register[*IPTServer](c, func(c *ioc.Container) *IPTServer {
		return NewIPTServer(ioc.Resolve[database.Databse](c))
	})
	return c
}
