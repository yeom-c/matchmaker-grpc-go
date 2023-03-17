package db

import (
	"database/sql"
	"sync"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
	db_battle "github.com/yeom-c/matchmaker-grpc-go/db/sqlc/battle"
	db_game "github.com/yeom-c/matchmaker-grpc-go/db/sqlc/game"

	"github.com/yeom-c/matchmaker-grpc-go/config"
	redis_battle "github.com/yeom-c/matchmaker-grpc-go/db/redis/battle"
	redis_session "github.com/yeom-c/matchmaker-grpc-go/db/redis/session"
)

var once sync.Once
var instance *store

type store struct {
	SessionRedis        *redis.Client
	SessionRedisQueries *redis_session.Queries
	BattleRedis         *redis.Client
	BattleRedisQueries  *redis_battle.Queries
	GameDb              map[int32]*sql.DB
	GameQueries         map[int32]*db_game.Queries
	BattleDb            *sql.DB
	BattleQueries       *db_battle.Queries
}

func Store() *store {
	once.Do(func() {
		sessionRedis := redis.NewClient(&redis.Options{
			Addr:     config.Config().RedisSessionAddr,
			Username: config.Config().RedisSessionUsername,
			Password: config.Config().RedisSessionPassword,
			DB:       config.Config().RedisSessionDb,
		})

		battleRedis := redis.NewClient(&redis.Options{
			Addr:     config.Config().RedisBattleAddr,
			Username: config.Config().RedisBattleUsername,
			Password: config.Config().RedisBattlePassword,
			DB:       config.Config().RedisBattleDb,
		})

		game0Db, err := sql.Open(config.Config().DbGame0Driver, config.Config().DbGame0Source)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to game_0 database")
		}

		game1Db, err := sql.Open(config.Config().DbGame1Driver, config.Config().DbGame1Source)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to game_1 database")
		}

		battleDb, err := sql.Open(config.Config().DbBattleDriver, config.Config().DbBattleSource)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to game_1 database")
		}

		instance = &store{
			SessionRedis:        sessionRedis,
			SessionRedisQueries: redis_session.New(sessionRedis),
			BattleRedis:         battleRedis,
			BattleRedisQueries:  redis_battle.New(battleRedis),
			GameDb:              map[int32]*sql.DB{0: game0Db, 1: game1Db},
			GameQueries: map[int32]*db_game.Queries{
				0: db_game.New(game0Db),
				1: db_game.New(game1Db),
			},
			BattleDb:      battleDb,
			BattleQueries: db_battle.New(battleDb),
		}
	})

	return instance
}
