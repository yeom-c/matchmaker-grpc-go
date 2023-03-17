package server

import (
	"context"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/rs/zerolog/log"
	"github.com/yeom-c/matchmaker-grpc-go/config"
	"github.com/yeom-c/protobuf-grpc-go/gen/golang/protos/error_res"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"runtime/debug"

	"net"
	"sync"

	"google.golang.org/grpc"
)

type server struct {
	Grpc *grpc.Server
}

var once sync.Once
var instance *server

func Server() *server {
	once.Do(func() {
		if instance == nil {
			instance = &server{}
			serverOptions := []grpc.ServerOption{
				grpc_middleware.WithStreamServerChain(
					grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(
						func(ctx context.Context, p interface{}) error {
							log.Error().Msgf("%s", p)
							if config.Config().DebugPrintStack {
								debug.PrintStack()
							}
							return status.Error(codes.Internal, error_res.Code_ServerErr.String())
						},
					)),
					logStreamServerInterceptor(),
					sessionStreamServerInterceptor(),
				),
			}

			if config.Config().UseSSL == true {
				creds, _ := credentials.NewServerTLSFromFile("./cert/server-cert.pem", "./cert/server-key.pem")
				serverOptions = append(serverOptions, grpc.Creds(creds))
			}

			instance.Grpc = grpc.NewServer(serverOptions...)
		}
	})

	return instance
}

func (server *server) Run() error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Config().GrpcServerPort))
	if err != nil {
		return err
	}

	log.Info().Int("port", config.Config().GrpcServerPort).Msg("started game server")
	err = server.Grpc.Serve(listen)
	if err != nil {
		return err
	}

	return nil
}
