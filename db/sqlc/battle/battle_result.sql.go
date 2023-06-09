// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: battle_result.sql

package db_battle

import (
	"context"
	"database/sql"
	"time"
)

const createBattleResult = `-- name: CreateBattleResult :execresult
INSERT INTO battle_result (account_user_id, match_account_user_id, channel_id, deck_id, battle_start_at) VALUES (?, ?, ?, ?, ?)
`

type CreateBattleResultParams struct {
	AccountUserID      int32     `json:"account_user_id"`
	MatchAccountUserID int32     `json:"match_account_user_id"`
	ChannelID          string    `json:"channel_id"`
	DeckID             int32     `json:"deck_id"`
	BattleStartAt      time.Time `json:"battle_start_at"`
}

func (q *Queries) CreateBattleResult(ctx context.Context, arg CreateBattleResultParams) (sql.Result, error) {
	return q.db.ExecContext(ctx, createBattleResult,
		arg.AccountUserID,
		arg.MatchAccountUserID,
		arg.ChannelID,
		arg.DeckID,
		arg.BattleStartAt,
	)
}

const getUserByAccountUserId = `-- name: GetUserByAccountUserId :one
SELECT id, account_user_id, match_point, match_win, match_lose, created_at FROM ` + "`" + `user` + "`" + ` WHERE account_user_id = ? LIMIT 1
`

func (q *Queries) GetUserByAccountUserId(ctx context.Context, accountUserID int32) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByAccountUserId, accountUserID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.AccountUserID,
		&i.MatchPoint,
		&i.MatchWin,
		&i.MatchLose,
		&i.CreatedAt,
	)
	return i, err
}
