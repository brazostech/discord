# CSCS Discord Bot

[![Latest release](https://img.shields.io/github/v/release/brazostech/discord)](https://github.com/brazostech/discord/releases)
[![Go version](https://img.shields.io/github/go-mod/go-version/brazostech/discord)](go.mod)
[![Lint](https://github.com/brazostech/discord/actions/workflows/lint.yml/badge.svg)](https://github.com/brazostech/discord/actions/workflows/lint.yml)
[![Release](https://github.com/brazostech/discord/actions/workflows/release.yml/badge.svg)](https://github.com/brazostech/discord/actions/workflows/release.yml)

A Discord bot for a book club that reads one book at a time. Every server tracks a single **Current Book** and the **Chapter** the club is on. The domain model lives in [CONTEXT.md](CONTEXT.md).

## Commands

| Command | Description |
| --- | --- |
| `/book register <name> [url]` | Set or replace the server's Current Book; starts at Chapter 0. |
| `/book update-chapter <chapter>` | Move the Current Book to any non-negative Chapter. |
| `/test` | Connectivity check. |

## Development

Requires Go 1.27+ and [just](https://github.com/casey/just).

```sh
just test   # go test ./...
just vet    # go vet ./...
just build  # build the bot into bin/bot
```

Install the application commands, then run the bot (gateway client):

```sh
go run ./cmd/commands -token <bot-token> -appid <application-id>
just start-bot <bot-token>
```

Storage is in-memory, so books and chapters reset when the bot restarts.

## Releases

[Conventional Commits](https://www.conventionalcommits.org) merged to `main` are analyzed by [semantic-release](https://semantic-release.org), which cuts `vX.Y.Z` tags and GitHub Releases. `feat` → minor, `fix`/`perf` → patch, breaking changes → major; all other types produce no release.
