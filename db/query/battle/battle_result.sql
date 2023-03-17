-- name: GetUserByAccountUserId :one
SELECT * FROM `user` WHERE account_user_id = ? LIMIT 1;

-- name: CreateBattleResult :execresult
INSERT INTO battle_result (account_user_id, match_account_user_id, channel_id, deck_id, battle_start_at) VALUES (?, ?, ?, ?, ?);
