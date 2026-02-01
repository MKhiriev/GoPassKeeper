package server

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

type grpcServer struct {
	server          *grpc.Server
	gRPCNetListener net.Listener
}

func (g *grpcServer) RunServer() {
	if err := g.server.Serve(g.gRPCNetListener); err != nil {
		fmt.Printf("gRPC server Serve: %v\n", err)
	}
}

func (g *grpcServer) Shutdown() {
	g.server.GracefulStop()
}
