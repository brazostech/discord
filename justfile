build-bot:
  go build -o bin/bot ./cmd/cscsbot

start-local pubkey:
  go run ./cmd/cscsbot -pubkey {{pubkey}} &
  ngrok http 8080

build: build-bot

test:
  go test ./...

vet:
  go vet ./...
