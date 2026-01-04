-- name: GetPlayers :many
SELECT * FROM game_players WHERE
  chat_id = $1
;

-- name: CreatePlayer :one
INSERT INTO game_players (
  chat_id, user_id, first_name, last_name, username
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: UpdatePlayer :many
UPDATE game_players SET
  first_name=$2, last_name=$3, username=$4
WHERE
  user_id = $1
RETURNING *;


-- name: GetRounds :many
SELECT * FROM game_rounds WHERE
  chat_id = $1
;

-- name: CreateRound :one
INSERT INTO game_rounds (
  chat_id, user_id, username
) VALUES (
  $1, $2, $3
) RETURNING *;
