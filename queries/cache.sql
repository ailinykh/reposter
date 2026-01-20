-- name: Get :one
SELECT * FROM cache WHERE
  key = $1
;

-- name: Set :one
INSERT INTO cache
  (key, value)
VALUES
  ($1, $2)
ON CONFLICT (key)
DO UPDATE SET
  value = EXCLUDED.value,
  updated_at = EXCLUDED.updated_at
RETURNING *;