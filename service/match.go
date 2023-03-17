package service

import (
	"context"
	"fmt"
	err_pb "github.com/yeom-c/protobuf-grpc-go/gen/golang/protos/error_res"
	match_pb "github.com/yeom-c/protobuf-grpc-go/gen/golang/protos/match"
	model_pb "github.com/yeom-c/protobuf-grpc-go/gen/golang/protos/model"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/yeom-c/matchmaker-grpc-go/config"
	"github.com/yeom-c/matchmaker-grpc-go/db"
	db_session "github.com/yeom-c/matchmaker-grpc-go/db/redis/session"
	db_battle "github.com/yeom-c/matchmaker-grpc-go/db/sqlc/battle"
	"github.com/yeom-c/matchmaker-grpc-go/helper"
)

type matchService struct {
	match_pb.MatchServiceServer
}

func NewMatchService() *matchService {
	return &matchService{}
}

type leaguePool struct {
	sync.RWMutex
	leagues map[int32]*league
}

func (lgp *leaguePool) getLeague(tier int32) *league {
	lgp.Lock()
	defer lgp.Unlock()

	if _, has := lgp.leagues[tier]; !has {
		return nil
	}
	return lgp.leagues[tier]
}

func (lgp *leaguePool) setLeague(tier int32, league *league) error {
	lgp.Lock()
	defer lgp.Unlock()

	if _, has := lgp.leagues[tier]; !has {
		lgp.leagues[tier] = league
	}

	return nil
}

type league struct {
	sync.RWMutex
	tier        int32
	connections map[int32]*connection
}

func (lg *league) lenConnections() int {
	lg.Lock()
	defer lg.Unlock()

	return len(lg.connections)
}

func (lg *league) getConnection(key int32) *connection {
	lg.Lock()
	defer lg.Unlock()

	if _, has := lg.connections[key]; !has {
		return nil
	}
	return lg.connections[key]
}

func (lg *league) setConnection(key int32, val *connection) error {
	lg.Lock()
	defer lg.Unlock()

	if existConn, has := lg.connections[key]; has {
		existConn.errCh <- helper.ErrorWithStack(err_pb.Code_SessionErrInvalidSession.String())
	}
	lg.connections[key] = val

	return nil
}

func (lg *league) deleteConnection(key int32, conn *connection) error {
	lg.Lock()
	defer lg.Unlock()

	if existConn, has := lg.connections[key]; has {
		if existConn == conn {
			delete(lg.connections, key)
		}
	}

	return nil
}

func (lg *league) matcher() {
	ctx := context.Background()
	for {
		if lg.lenConnections() < 2 {
			time.Sleep(500 * time.Millisecond)
		}

		battleServerUrl := getBattleServerUrl()
		matchConnections := []*connection{}
		lg.Lock()
		for _, readyConn := range lg.connections {
			timeoutSec := time.Now().Sub(*readyConn.readyAt).Seconds()
			if readyConn.done {
				continue
			} else if timeoutSec >= limitTimeoutSec {
				// 봇 매칭.
				var matchAccountUserId int32
				var matchAccountUserGameDb int32
				channelId := fmt.Sprintf("%d_%d_%v", readyConn.session.AccountUserId, matchAccountUserId, time.Now().Unix())
				err := db.Store().SessionRedisQueries.UpdateBattleChannelId(ctx, readyConn.session.AccountId, battleServerUrl, channelId)
				if err != nil {
					readyConn.done = true
					readyConn.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
					continue
				}
				err = db.Store().BattleRedisQueries.EnterChannel(ctx, channelId, readyConn.session.AccountUserId)
				if err != nil {
					readyConn.done = true
					readyConn.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
					continue
				}
				_, err = db.Store().BattleQueries.CreateBattleResult(ctx, db_battle.CreateBattleResultParams{
					AccountUserID:      readyConn.session.AccountUserId,
					MatchAccountUserID: matchAccountUserId,
					ChannelID:          channelId,
					DeckID:             readyConn.deckId,
					BattleStartAt:      time.Now().UTC(),
				})
				if err != nil {
					readyConn.done = true
					readyConn.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
					continue
				}

				err = readyConn.stream.Send(&match_pb.StreamRes{
					BattleServerUrl: battleServerUrl,
					BattleChannelId: channelId,
					MatchUser: &model_pb.AccountUserData{
						Id:        matchAccountUserId,
						AccountId: 0,
						GameDb:    matchAccountUserGameDb,
						Nickname:  "퀘이사봇",
					},
					MatchUserPoint: rand.Int31n(150),
				})
				if err != nil {
					readyConn.done = true
					readyConn.errCh <- err
					continue
				}

				readyConn.done = true
				readyConn.errCh <- nil
				continue
			}

			if len(matchConnections) < 2 {
				matchConnections = append(matchConnections, readyConn)
			} else {
				break
			}
		}
		lg.Unlock()

		// 매칭 성공.
		if len(matchConnections) == 2 {
			now := time.Now().UTC()
			matchConn1 := matchConnections[0]
			matchConn2 := matchConnections[1]
			matchConnections = []*connection{}
			if matchConn1.done {
				continue
			}
			if matchConn2.done {
				continue
			}

			channelId := fmt.Sprintf("%d_%d_%v", matchConn1.session.AccountUserId, matchConn2.session.AccountUserId, time.Now().Unix())
			err := db.Store().SessionRedisQueries.UpdateBattleChannelId(ctx, matchConn1.session.AccountId, battleServerUrl, channelId)
			if err != nil {
				matchConn1.done = true
				matchConn1.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
				continue
			}
			err = db.Store().SessionRedisQueries.UpdateBattleChannelId(ctx, matchConn2.session.AccountId, battleServerUrl, channelId)
			if err != nil {
				matchConn2.done = true
				matchConn2.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
				continue
			}
			err = db.Store().BattleRedisQueries.EnterChannel(ctx, channelId, matchConn1.session.AccountUserId, matchConn2.session.AccountUserId)
			if err != nil {
				matchConn1.done = true
				matchConn2.done = true
				matchConn1.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
				matchConn2.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
				continue
			}

			deck1, err := db.Store().GameQueries[matchConn1.session.GameDb].GetDeck(ctx, matchConn1.deckId)
			if err != nil {
				matchConn1.done = true
				matchConn1.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
				continue
			}
			if deck1.AccountUserID != matchConn1.session.AccountUserId {
				matchConn1.done = true
				matchConn1.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
				continue
			}
			_, err = db.Store().BattleQueries.CreateBattleResult(ctx, db_battle.CreateBattleResultParams{
				AccountUserID:      matchConn1.session.AccountUserId,
				MatchAccountUserID: matchConn2.session.AccountUserId,
				ChannelID:          channelId,
				DeckID:             matchConn1.deckId,
				BattleStartAt:      now,
			})
			if err != nil {
				matchConn1.done = true
				matchConn1.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
				continue
			}

			deck2, err := db.Store().GameQueries[matchConn2.session.GameDb].GetDeck(ctx, matchConn2.deckId)
			if err != nil {
				matchConn2.done = true
				matchConn2.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
				continue
			}
			if deck2.AccountUserID != matchConn2.session.AccountUserId {
				matchConn2.done = true
				matchConn2.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
				continue
			}
			_, err = db.Store().BattleQueries.CreateBattleResult(ctx, db_battle.CreateBattleResultParams{
				AccountUserID:      matchConn2.session.AccountUserId,
				MatchAccountUserID: matchConn1.session.AccountUserId,
				ChannelID:          channelId,
				DeckID:             matchConn2.deckId,
				BattleStartAt:      now,
			})
			if err != nil {
				matchConn2.done = true
				matchConn2.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
				continue
			}

			matchBattleUser1, err := db.Store().BattleQueries.GetUserByAccountUserId(ctx, matchConn1.session.AccountUserId)
			if err != nil {
				matchConn1.done = true
				matchConn1.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
				continue
			}
			matchBattleUser2, err := db.Store().BattleQueries.GetUserByAccountUserId(ctx, matchConn2.session.AccountUserId)
			if err != nil {
				matchConn2.done = true
				matchConn2.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
				continue
			}

			err = matchConn1.stream.Send(&match_pb.StreamRes{
				BattleServerUrl: battleServerUrl,
				BattleChannelId: channelId,
				MatchUser: &model_pb.AccountUserData{
					Id:        matchConn2.session.AccountUserId,
					AccountId: matchConn2.session.AccountId,
					GameDb:    matchConn2.session.GameDb,
					Nickname:  matchConn2.session.Nickname,
				},
				MatchUserPoint: matchBattleUser2.MatchPoint,
				RoomMaker:      true,
			})
			if err != nil {
				matchConn1.done = true
				matchConn1.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
				continue
			}
			err = matchConn2.stream.Send(&match_pb.StreamRes{
				BattleServerUrl: battleServerUrl,
				BattleChannelId: channelId,
				MatchUser: &model_pb.AccountUserData{
					Id:        matchConn1.session.AccountUserId,
					AccountId: matchConn1.session.AccountId,
					GameDb:    matchConn1.session.GameDb,
					Nickname:  matchConn1.session.Nickname,
				},
				MatchUserPoint: matchBattleUser1.MatchPoint,
				RoomMaker:      false,
			})
			if err != nil {
				matchConn2.done = true
				matchConn2.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
				continue
			}

			matchConn1.done = true
			matchConn2.done = true
			matchConn1.errCh <- nil
			matchConn2.errCh <- nil
		}
	}
}

type connection struct {
	stream  match_pb.MatchService_ConnectStreamServer
	session db_session.Session
	league  *league
	errCh   chan error
	done    bool
	deckId  int32
	readyAt *time.Time
}

func (c *connection) receiver() {
	for {
		recvMsg, err := c.stream.Recv()
		if c.done {
			return
		} else if err != nil {
			c.done = true
			c.errCh <- helper.ErrorWithStack(err_pb.Code_MatchmakerServerErrStreamDisconnect.String())
			return
		} else {
			// 대기 리스트 저장.
			if recvMsg.Action == "READY" {
				now := time.Now()
				deckId, _ := strconv.Atoi(recvMsg.Data)
				c.deckId = int32(deckId)
				c.readyAt = &now
				c.league.setConnection(c.session.AccountUserId, c)
			}
		}
	}
}

var limitTimeoutSec = float64(15)
var lgPool = leaguePool{
	leagues: map[int32]*league{},
}

// TODO: battle server 가 여러대 일경우 적당한 server url 가져오는 방법 필요.
func getBattleServerUrl() string {
	url := "quasargamestudio.asuscomm.com:17000"
	if config.Config().Env == "test" {
		url = "quasargamestudio.asuscomm.com:27000"
	}

	return url
}

func (s *matchService) ConnectStream(stream match_pb.MatchService_ConnectStreamServer) error {
	log.Info().Int("goroutine", runtime.NumGoroutine()).Msg("connect stream")
	ctx := stream.Context()
	mySession := db_session.GetMySession(ctx)
	if mySession.AccountUserId == 0 {
		return helper.ErrorWithStack(err_pb.Code_SessionErrEmptyAccountUser.String())
	}

	// TODO: session tier 처리 필요.
	myTier := int32(0)
	myLeague := lgPool.getLeague(myTier)
	if myLeague == nil {
		myLeague = &league{
			tier:        myTier,
			connections: map[int32]*connection{},
		}
		// leaguePool 에 league 추가.
		lgPool.setLeague(myTier, myLeague)
		go myLeague.matcher()
	}

	// 새 connection 생성.
	conn := connection{
		stream:  stream,
		session: mySession,
		league:  myLeague,
		errCh:   make(chan error),
	}
	defer func() {
		conn.done = true
		conn.league.deleteConnection(conn.session.AccountUserId, &conn)
		//close(conn.errCh)
	}()

	// 메세지 수신.
	go conn.receiver()

	select {
	case err := <-conn.errCh:
		return err
	}
}

func (s *matchService) LogConnectStream(_ context.Context, _ *model_pb.Empty) (*model_pb.Result, error) {
	for tier, league := range lgPool.leagues {
		league.Lock()
		for accountUserId, conn := range league.connections {
			log.Debug().Int32("tier", tier).Int32("account_user_id", accountUserId).Interface("session", conn.session).Time("ready_at", *conn.readyAt).Bool("done", conn.done).Msg("")
		}
		league.Unlock()
	}

	return &model_pb.Result{
		Result: 1,
	}, nil
}
