package server

import (
	"fmt"
	"net"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	myGRPC "github.com/MKhiriev/go-pass-keeper/internal/handler/grpc"

	"google.golang.org/grpc"
)

type grpcServer struct {
	handler *myGRPC.Handler

	server          *grpc.Server
	gRPCNetListener net.Listener
}

func newGRPCServer(handler *myGRPC.Handler, cfg *config.Server) *grpcServer {
	return &grpcServer{}
}

func (g *grpcServer) RunServer() {
	if err := g.server.Serve(g.gRPCNetListener); err != nil {
		fmt.Printf("gRPC server Serve: %v\n", err)
	}
}

func (g *grpcServer) Shutdown() {
	g.server.GracefulStop()
}
