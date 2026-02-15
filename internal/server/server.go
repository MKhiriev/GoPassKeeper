package server

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/handler"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
)

type server struct {
	httpServer *httpServer
	gRPCServer *grpcServer
	logger     *logger.Logger
}

func NewServer(handlers *handler.Handlers, cfg config.Server, logger *logger.Logger) (Server, error) {
	logger.Info().Msg("creating new server...")
	servers := new(server)

	if cfg.HTTPAddress != "" {
		servers.httpServer = newHTTPServer(handlers.HTTP.Init(), cfg, logger)
	}
	if cfg.GRPCAddress != "" {
		servers.gRPCServer = newGRPCServer(handlers.GRPC, cfg, logger)
	}

	if servers.httpServer == nil && servers.gRPCServer == nil {
		return nil, errNoServersAreCreated
	}

	servers.logger = logger

	return servers, nil
}

func (s *server) RunServer() {
	if err := s.run(); err != nil {
		s.logger.Info().Msgf("Error running server: %v \n", err)
	}
}

func (s *server) Shutdown() {
	// finish HTTP server
	if s.httpServer != nil {
		s.httpServer.Shutdown()
	}

	// finish gRPC server
	if s.gRPCServer != nil {
		s.gRPCServer.Shutdown()
	}
}

func (s *server) run() error {
	// check if any server was created
	if s.httpServer == nil && s.gRPCServer == nil {
		return errors.New("no servers to run")
	}

	idleConnectionsClosed := make(chan struct{})
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer stop()

	// listen for stop signals
	go func() {
		<-ctx.Done()

		// finish started servers
		s.Shutdown()

		close(idleConnectionsClosed)
	}()

	// launch all created servers
	if s.httpServer != nil {
		s.logger.Info().Msg("Launching HTTP server")
		go s.httpServer.RunServer()
	}
	if s.gRPCServer != nil {
		s.logger.Info().Msg("Launching GRPC server")
		go s.gRPCServer.RunServer()
	}

	<-idleConnectionsClosed
	s.logger.Info().Msg("server Shutdown gracefully")

	return nil
}
