# AGENTS.md

This file provides guidance to AI coding agents (Claude Code and others) when working with code in this repository. `CLAUDE.md` is a symlink to this file.

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

Run locally (requires `BOT_TOKEN` and `BOT_ID` in the environment, or in an optional `.env`):

```sh
go run . local
```

Build + run as Docker container (Taskfile loads `.env` and passes `BOT_TOKEN`/`BOT_ID` via `docker run -e`):

```sh
task build
task run
```

## Architecture

```
main.go                  → init logger + DB, start HTTP server + Discord bot, graceful shutdown
app/server.go            → wires Gin router (Prometheus /metrics) and Discord bot together
app/config/              → loads configs/{local|prod|test}.yaml by CLI arg; BOT_TOKEN/BOT_ID env vars override Discord credentials
app/model/               → PantryItem struct, rendering (ToMarkdownTable/ToList), and interfaces
app/service/             → Discord event handling and message parsing
app/repository/          → SQLite implementation of DatabaseClient and PantryClient
app/controller/          → Gin router setup; currently only exposes /metrics
```

### Discord event flow

`DiscordBot` (in `service/discord_bot.go`) routes all events to channel-specific `BotHandler` implementations. The active channel is a `groceries` or `freezer` channel, each backed by its own SQLite table.

Every message event, Ready event, modal submit and Undo click goes through `PantryHandler.syncChannel(session, action)` (in `service/pantry_handler.go`):

1. A per-channel `sync.Mutex` is held for the whole sync, so concurrent events on the same channel are serialized and never apply the same input twice.
2. `PreProcessMessageEvent` reads the last 100 channel messages, keeps the single oldest bot message (the markdown table) for reuse, marks every other message as removable, and joins user message contents **oldest first** (by timestamp, then snowflake ID) as raw input.
3. `deleteMessages` deletes the removable messages **before** any input is applied: bulk delete first, then a per-message fallback (which also covers messages older than 14 days). If any delete still fails, the sync aborts without mutating anything; input from messages already deleted in the fallback is logged and dropped rather than re-applied on every later sync.
4. The optional `action` runs (modal submit or Undo), then any non-empty user input is parsed and persisted via `UpdateItems`.
5. `PublishItems` rewrites the channel, splitting tables over 2000 characters into multiple messages; the Edit/Undo buttons always sit on the last message.

Items are listed ordered by `date, id` (dates are truncated to the minute, so `id` keeps same-minute items stable), which keeps the index a user sees in sync with the index used for remove/edit.

Modal submits send the deferred interaction response first (Discord's 3 s window), then apply the modal input inside `syncChannel`.

**Undo** restores a single-step snapshot stored in the shared `snapshots` table (one row per pantry table, items JSON-encoded). A snapshot is taken right before each mutating action, i.e. a sync that applies user message input, or a modal submit. Undo reverts exactly that last action, once: restoring deletes the snapshot and Undo never takes one, so a second Undo is a no-op. Syncs with no input and no action (startup, empty channel) never touch the snapshot.

`DiscordBot.Ready` (also fired on reconnects) initializes all handlers concurrently, logs each handler's error, and never terminates the process.

### Message parsing (service/pantry_handling.go)

Three regex-routed operations on each line of user input:

- **Add**: `<qty> <name>` or `<name> <qty>` — quantity defaults to 1
- **Remove**: digit list / ranges (`2`, `2 4`, `2-5`, `* 2 4`) — `*` means remove all except listed
- **Edit quantity**: `<index>++` / `<index>--` with optional delta (`3--2`)

### Config loading

The CLI first argument selects the config file: `local`, `prod`, or any string containing `"test"` maps to `test`. Anything else (and no argument) falls back to `local`. `BOT_TOKEN` and `BOT_ID` are read from the process environment and override the YAML values when set; a `.env` file (loaded via godotenv) is optional and only a local-dev convenience. Startup fails fast (`log.Fatal`) if the token or bot ID is empty or still the `BOT_TOKEN`/`BOT_ID` placeholder, except under the `test` config.

### Tests

Tests in `service` and `model` packages are table-driven. Service tests use a real in-memory SQLite database (`:memory:` from `configs/test.yaml`) — there are no mocks for the database layer. CGO is required to build (`go-sqlite3`).

### Deployment

Built as a multi-stage Docker image (builder runs tests, final image is Alpine). Published to GHCR (`ghcr.io/maribowman/roastbeef-swag`) and deployed to a Synology NAS. Credentials are not baked into the image: there is no `sed` injection, and `configs/prod.yaml` ships only the `BOT_TOKEN`/`BOT_ID` placeholders. The container must be started with `-e BOT_TOKEN -e BOT_ID` (as `task run` does), otherwise it exits at startup.
