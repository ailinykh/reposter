-- name: GetSettings :one
SELECT * FROM chat_settings WHERE
  chat_id = $1 AND key = $2
;

-- name: SetSettings :one
INSERT INTO chat_settings
  (chat_id, key, value)
VALUES
  ($1, $2, $3)
ON CONFLICT (chat_id, key)
DO UPDATE SET
  value = EXCLUDED.value,
  updated_at = EXCLUDED.updated_at
RETURNING *;