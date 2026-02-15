package server

import (
	"net"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	myGRPC "github.com/MKhiriev/go-pass-keeper/internal/handler/grpc"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"

	"google.golang.org/grpc"
)

type grpcServer struct {
	handler *myGRPC.Handler

	server          *grpc.Server
	gRPCNetListener net.Listener

	logger *logger.Logger
}

func newGRPCServer(handler *myGRPC.Handler, cfg config.Server, logger *logger.Logger) *grpcServer {
	return &grpcServer{
		logger: logger,
	}
}

func (g *grpcServer) RunServer() {
	if err := g.server.Serve(g.gRPCNetListener); err != nil {
		g.logger.Error().Msgf("gRPC server Serve: %v\n", err)
	}
}

func (g *grpcServer) Shutdown() {
	g.logger.Info().Msg("GRPC server Shutdown")
	g.server.GracefulStop()
}
