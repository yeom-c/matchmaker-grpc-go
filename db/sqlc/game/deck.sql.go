// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: deck.sql

package db_game

import (
	"context"
)

const getDeck = `-- name: GetDeck :one
SELECT id, account_user_id, ` + "`" + `index` + "`" + `, name, character_id_0, character_id_1, character_id_2, character_id_3, character_id_4, created_at FROM deck WHERE id = ?
`

func (q *Queries) GetDeck(ctx context.Context, id int32) (Deck, error) {
	row := q.db.QueryRowContext(ctx, getDeck, id)
	var i Deck
	err := row.Scan(
		&i.ID,
		&i.AccountUserID,
		&i.Index,
		&i.Name,
		&i.CharacterID0,
		&i.CharacterID1,
		&i.CharacterID2,
		&i.CharacterID3,
		&i.CharacterID4,
		&i.CreatedAt,
	)
	return i, err
}
