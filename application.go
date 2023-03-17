package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/yeom-c/matchmaker-grpc-go/config"
	"github.com/yeom-c/matchmaker-grpc-go/server"
	"github.com/yeom-c/matchmaker-grpc-go/service"
	"os"
	"time"
)

func main() {
	loc, _ := time.LoadLocation("UTC")
	time.Local = loc

	// zerolog Stack() 사용 위함.
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	if config.Config().LogFormat != "json" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	server := server.Server()
	service.NewService(server.Grpc)

	err := server.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}
