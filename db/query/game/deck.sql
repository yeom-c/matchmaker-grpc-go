-- name: GetDeck :one
SELECT * FROM deck WHERE id = ?;
