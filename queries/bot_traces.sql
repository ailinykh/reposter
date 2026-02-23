-- name: CreateBotTrace :exec
INSERT INTO bot_traces
  (bot_id, method, request, response)
VALUES
  ($1, $2, sqlc.narg(request), @response)
;