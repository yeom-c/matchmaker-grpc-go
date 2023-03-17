package server

import (
	"context"
	err_pb "github.com/yeom-c/protobuf-grpc-go/gen/golang/protos/error_res"
	"strconv"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/rs/zerolog/log"
	"github.com/yeom-c/matchmaker-grpc-go/config"
	"github.com/yeom-c/matchmaker-grpc-go/db"
	db_session "github.com/yeom-c/matchmaker-grpc-go/db/redis/session"
	"github.com/yeom-c/matchmaker-grpc-go/helper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func logStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()

		err := handler(srv, stream)

		duration := time.Since(startTime)
		statusCode := codes.Unknown
		if st, ok := status.FromError(err); ok {
			statusCode = st.Code()
		}

		ctx := stream.Context()
		md, _ := metadata.FromIncomingContext(ctx)
		logger := log.Info()
		if err != nil {
			logger = log.Error().Err(err)
			if config.Config().DebugPrintStack {
				logger = log.Error().Stack().Err(err)
			}

			err = status.Error(statusCode, err.Error())
		}
		logger.
			Dur("duration", duration).
			Str("status", statusCode.String()).
			Int("status_code", int(statusCode)).
			Str("method", info.FullMethod).
			Interface("metadata", md).
			Msg("")

		return err
	}
}

func sessionStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return helper.ErrorWithStack(err_pb.Code_ServerErr.String())
		}

		if len(md["session_id"]) != 1 || len(md["account_id"]) != 1 {
			return helper.ErrorWithStack(err_pb.Code_SessionErrInvalidHeader.String())
		}
		sessionId := md["session_id"][0]
		accountId, _ := strconv.Atoi(md["account_id"][0])
		// session key 비교
		session, _ := db.Store().SessionRedisQueries.GetAccountSession(ctx, int32(accountId))
		if session.SessionId == "" {
			return helper.ErrorWithStack(err_pb.Code_SessionErrEmptySession.String())
		}

		if sessionId != session.SessionId {
			return helper.ErrorWithStack(err_pb.Code_SessionErrInvalidSession.String())
		}

		wrapStream := grpc_middleware.WrapServerStream(stream)
		wrapStream.WrappedContext = db_session.SetMySession(context.Background(), session)
		return handler(srv, wrapStream)
	}
}
