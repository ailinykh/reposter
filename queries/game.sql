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
  first_name=$2, last_name=$3, username=$4, updated_at=NOW()
WHERE
  user_id = $1
RETURNING *;


-- name: GetRounds :many
SELECT
	game_rounds.*,
	game_players.username AS actual_username
FROM
	game_rounds
	LEFT JOIN game_players ON game_rounds.user_id = game_players.user_id
	AND game_rounds.chat_id = game_players.chat_id
WHERE
	game_rounds.chat_id = $1;


-- name: CreateRound :one
INSERT INTO game_rounds (
  chat_id, user_id, username
) VALUES (
  $1, $2, $3
) RETURNING *;
