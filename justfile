build-bot:
  go build -o bin/bot ./cmd/cscsbot

start-bot token:
  go run ./cmd/cscsbot -token {{token}}

build: build-bot

test:
  go test ./...

vet:
  go vet ./...
