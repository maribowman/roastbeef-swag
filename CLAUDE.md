# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A Discord bot written in Go that tracks grocery and freezer inventories. Users interact by posting messages directly in Discord channels; the bot parses them, updates SQLite, and rewrites the channel with a formatted markdown table.

## Commands

Run all tests (uses in-memory SQLite via `configs/test.yaml`):
```sh
go test ./...
```

Run a single test by name:
```sh
go test ./app/service/... -run TestUpdateItems
go test ./app/model/... -run TestToMarkdownTable
```

Run locally (requires `.env` with `BOT_TOKEN` and `BOT_ID`):
```sh
go run . local
```

Build + run as Docker container (reads `BOT_ID`/`BOT_TOKEN` from `.env`):
```sh
task build
task run
```

## Architecture

```
main.go                  → init logger + DB, start HTTP server + Discord bot, graceful shutdown
app/server.go            → wires Gin router (Prometheus /metrics) and Discord bot together
app/config/              → loads configs/{local|int|prod|test}.yaml by CLI arg; .env overrides Discord credentials
app/model/               → PantryItem struct, rendering (ToMarkdownTable/ToList), and interfaces
app/service/             → Discord event handling and message parsing
app/repository/          → SQLite implementation of DatabaseClient and PantryClient
app/controller/          → Gin router setup; currently only exposes /metrics
```

### Discord event flow

`DiscordBot` (in `service/discord_bot.go`) routes all events to channel-specific `BotHandler` implementations. The active channel is a `groceries` or `freezer` channel, each backed by its own SQLite table.

On every message event, the handler calls `PreProcessMessageEvent` to read the last 100 channel messages, identifies the single authoritative bot message (the markdown table), collects all user messages as raw input, bulk-deletes everything, then calls `UpdateItems` to parse and persist changes, and finally `PublishItems` to rewrite the channel.

### Message parsing (service/pantry_handling.go)

Three regex-routed operations on each line of user input:
- **Add**: `<qty> <name>` or `<name> <qty>` — quantity defaults to 1
- **Remove**: digit list / ranges (`2`, `2 4`, `2-5`, `* 2 4`) — `*` means remove all except listed
- **Edit quantity**: `<index>++` / `<index>--` with optional delta (`3--2`)

### Config loading

The CLI first argument selects the config file: `local`, `int`, `prod`, or any string containing `"test"` maps to `test`. Default is `local`. `.env` file (loaded via godotenv) always overrides `BOT_TOKEN` and `BOT_ID`.

### Tests

Tests in `service` and `model` packages are table-driven. Service tests use a real in-memory SQLite database (`:memory:` from `configs/test.yaml`) — there are no mocks for the database layer. CGO is required to build (`go-sqlite3`).

### Deployment

Built as a multi-stage Docker image (builder runs tests, final image is Alpine). Published to GHCR (`ghcr.io/maribowman/roastbeef-swag`) and deployed to a Synology NAS. The `prod` config has `BOT_TOKEN`/`BOT_ID` injected via `sed` during `task build` from `.env`.
