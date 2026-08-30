build-bot:
  go build -o bin/bot ./cmd/cscsbot

build-commands:
  go build -o bin/commands ./cmd/commands

build: build-bot build-commands

test:
  go test ./...

vet:
  go vet ./...
