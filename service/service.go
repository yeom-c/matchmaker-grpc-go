package service

import (
	match_pb "github.com/yeom-c/protobuf-grpc-go/gen/golang/protos/match"
	"google.golang.org/grpc"
)

func NewService(grpcServer *grpc.Server) {
	match_pb.RegisterMatchServiceServer(grpcServer, NewMatchService())
}
