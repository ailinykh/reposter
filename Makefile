ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: bot worker pkill test clean restart

APP1 ?= bin/bot
APP2 ?= bin/worker
GO_FILES = $(shell find . -name '*.go' -not -path './vendor/*')

bot: bin/bot
	@$^ &
	@fswatch -x -o --event Created --event Updated --event Renamed -r -e '.*' -i '\.go$$' . | xargs -n1 -I{} make restart APP=bot || make pkill

worker: bin/worker
	@$^ &
	@fswatch -x -o --event Created --event Updated --event Renamed -r -e '.*' -i '\.go$$' . | xargs -n1 -I{} make restart APP=worker || make pkill

pkill:
	@pkill -f "bin/(bot|worker)" || true

test:
	go test ./... -v -coverprofile=coverage.txt -race -covermode=atomic -timeout 5m

bin/%: cmd/% $(GO_FILES)
	@echo "🛠️ building $< ..."
	@go build -o $@ ./$<   
# $@ = bin/bot, $< = cmd/bot

clean:
	rm -f $(APP1) $(APP2)

restart: pkill
	@make $(APP)