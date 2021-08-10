package main

import (
	"context"
	"net"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	conf "github.com/amazingchow/photon-dance-consistent-hashing/internal/config"
	pb_api "github.com/amazingchow/photon-dance-consistent-hashing/pb/api"
)

func serveGPRC(ctx context.Context, srv *consistentHashingServiceServer, cfg *conf.Node) {
	l, err := net.Listen("tcp", cfg.GRPCEndpoint)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start grpc service")
	}

	opts := []grpc.ServerOption{
		grpc.MaxSendMsgSize(64 * 1024 * 1024),
		grpc.MaxRecvMsgSize(64 * 1024 * 1024),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             1 * time.Minute,
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    10 * time.Minute,
			Timeout: 20 * time.Second,
		}),
	}
	grpcServer := grpc.NewServer(opts...)
	pb_api.RegisterConsistentHashingServiceServer(grpcServer, srv)
	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	log.Info().Msgf("grpc service is listening at \x1b[1;31m%s\x1b[0m", cfg.GRPCEndpoint)
	go func() {
		if err := grpcServer.Serve(l); err != nil {
			log.Warn().Err(err)
		}
	}()

GRPC_LOOP:
	for { // nolint
		select {
		case <-ctx.Done():
			break GRPC_LOOP
		}
	}

	grpcServer.GracefulStop()
	log.Info().Msg("stop grpc service")
}
