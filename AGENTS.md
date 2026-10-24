# AGENTS.md

This file provides guidance to AI coding agents (Claude Code and others) when working with code in this repository. `CLAUDE.md` is a symlink to this file.

## What this is

A Discord bot written in Go that tracks grocery and freezer inventories. Users interact by posting messages directly in Discord channels; the bot parses them, updates SQLite, and rewrites the channel with a formatted markdown table.

## Commands

Verification gate (CGO is required: `go-sqlite3`):

```sh
go vet ./... && go test ./... -count=1 && go test -race ./... -count=1
```

Run a single test by name:

```sh
go test ./app/service/... -run TestUpdateItems
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
main.go          → init logger + DB, start HTTP server + Discord bot; shutdown closes Discord session, then HTTP server, then DB
app/server.go    → wires Gin router and Discord bot together
app/config/      → loads configs/{local|prod|test}.yaml by CLI arg; BOT_TOKEN/BOT_ID env vars override Discord credentials
app/model/       → PantryItem, rendering (ToMarkdownTable/ToList), interfaces; mocks/ holds a testify PantryClient mock (currently unused)
app/service/     → Discord event handling (discord_bot.go, pantry_handler.go) and message parsing/publishing (pantry_handling.go, suggestion helpers in suggestions.go)
app/repository/  → SQLite implementation of DatabaseClient and PantryClient, incl. undo snapshots
app/controller/  → Gin router setup; only exposes Prometheus /metrics
```

### Discord event flow

`DiscordBot` (in `service/discord_bot.go`) creates one `PantryHandler` per channel listed in the config's `discord.channels`, keyed by channel ID. The channel `name` must be `groceries` or `freezer` (other names are logged and skipped); it selects the SQLite table name, date format and modal title. Messages from the bot itself are ignored.

Every message event, Ready event, modal submit, Undo click and suggestion 👍 click goes through `PantryHandler.syncChannel(session, action)` (in `service/pantry_handler.go`):

1. A per-channel `sync.Mutex` is held for the whole sync, so concurrent events on the same channel are serialized and never apply the same input twice.
2. `PreProcessMessageEvent` reads the last 100 channel messages. Suggestion messages (bot messages with the 👍 button, see `isSuggestionMessage`) are skipped first: they are neither the table nor removable. Of the remaining messages it keeps the single oldest bot message (the markdown table) for reuse, marks every other message as removable, and joins user message contents **oldest first** (by timestamp, then snowflake ID) as raw input. In groceries, `?` lines are split off per message (keeping the author ID) as suggestions instead of being joined into the input.
3. `deleteMessages` deletes the removable messages **before** any input is applied: bulk delete first, then a per-message fallback (which also covers messages older than 14 days). If any delete still fails, the sync aborts without mutating anything, discarding the suggestions too; input from messages already deleted in the fallback is logged and dropped rather than re-applied on every later sync.
4. The optional `action` runs (modal submit, Undo or suggestion accept), then any non-empty user input is applied via `ApplyUserInput` (snapshot, then `UpdateItems`). Suggestion-only input takes no snapshot.
5. `PublishItems` rewrites the channel, splitting tables over 2000 characters into multiple messages; the Edit/Undo buttons always sit on the last message.
6. One suggestion message per extracted `?` line is sent below the table (`💡 Suggestion: <qty> <name> (by @author)`, 👍/👎 buttons, no ping). A failed send is logged and the suggestion dropped.

Items are listed ordered by `date, id` (dates are truncated to the minute, so `id` keeps same-minute items stable), which keeps the index a user sees in sync with the index used for remove/edit.

Interaction handlers (modal submit, Undo, suggestion 👍/👎) send the deferred interaction response first (Discord's 3 s window), then run their action inside `syncChannel` (👎 only deletes the suggestion message under the channel lock, without a sync).

**Suggestions** are stateless: an open suggestion exists only as its Discord message, and the item text is read back from the content via `parseSuggestion`, so they survive restarts without a DB table. The first click decides. 👍 deletes the suggestion message, then adds the item via `ApplySuggestion`; if the delete fails (e.g. a concurrent click already took it), the add is skipped, so an item is added exactly once. Known edge: when the table grows into several chunks, new chunks may appear below open suggestions.

**Undo** restores a single-step snapshot stored in the shared `snapshots` table (one row per pantry table, items JSON-encoded). `ApplyUserInput`, `ApplyModalInput` and `ApplySuggestion` take the snapshot right before mutating; every mutation path must go through one of them. An accepted suggestion is undone like any add; the suggestion message is not restored. Undo reverts exactly that last action, once: restoring deletes the snapshot and Undo never takes one, so a second Undo is a no-op. Syncs with no input and no action (startup, empty channel) never touch the snapshot. Known edge: if a modal submit and pending channel messages land in the same sync, the channel input snapshots after the modal, so Undo reverts only the channel input.

`DiscordBot.Ready` (also fired on reconnects) initializes all handlers concurrently, logs each handler's error, and never terminates the process.

### Message parsing (service/pantry_handling.go)

In groceries, `?` lines (`? <qty> <item>`, `?milk`) are split off by `extractSuggestions` (in `service/suggestions.go`) before `UpdateItems` sees the input; a bare `?` is dropped. In freezer they reach `UpdateItems` and are added literally.

`UpdateItems` routes each line of channel input by regex:

- **Edit quantity**: `<index>++` / `<index>--` with optional delta (`3--2`); a quantity of 0 or less removes the item
- **Remove**: digit list / ranges (`2`, `2 4`, `2-5`, `* 2 4`); `*` means remove all except listed
- **Add**: anything else, as `<qty> <name>` or `<name> <qty>`; quantity defaults to 1

`UpdateItemsFromModal` treats the modal text as the whole new list. The modal is prefilled by `ToList` as `[<index>] <qty> <name>` lines: a line with a valid `[n]` prefix updates item n (keeping its date), a line without one (or with an out-of-range index) adds a new item, and every existing item with no matching line is removed.

### Config loading

The CLI first argument selects the config file: `local`, `prod`, or any string containing `"test"` maps to `test`. Anything else (and no argument) falls back to `local`. `go test` picks `test` automatically because the test binary's first argument is a `-test.*` flag. Config files are found by walking up from the working directory to `go.mod` (falling back to `/`, which is why the image ships `/configs`).

`BOT_TOKEN` and `BOT_ID` are read from the process environment and override the YAML values when set; a `.env` file (loaded via godotenv) is optional and only a local-dev convenience. Startup fails fast (`log.Fatal`) if the token or bot ID is empty or still the `BOT_TOKEN`/`BOT_ID` placeholder, except under the `test` config.

### Tests

- Table-driven with `// given` / `// when` / `// then` comments.
- Service and repository tests run against a real SQLite database (`:memory:` from `configs/test.yaml`), not mocks. `RemoveAllItems` resets items and IDs but leaves the undo snapshot, so snapshot-dependent cases use their own table name (e.g. `undo_tests_<n>`).
- `:memory:` is per connection: a multi-statement operation must run on a single transaction or connection, otherwise part of it hits a different, empty database (see `RestoreSnapshot`).
- Item timestamps come from the package-level `now` in `pantry_handling.go`; tests override it to freeze time.
- Discord-facing code (`syncChannel`, `deleteMessages`, `PublishItems`, `DiscordBot.Ready`) has no unit tests because there is no fake `discordgo.Session`. Verify changes there with a manual run against a test channel.

### Deployment

The Dockerfile is a multi-stage build: the builder runs `go test ./...` (failing tests fail the image), and the final image is Alpine running `/main` with the DB at `/data/roastbeef-swag.db`. The GitHub workflow builds and pushes to GHCR (`ghcr.io/maribowman/roastbeef-swag`) on tag pushes only. Deploying to the Synology NAS is manual (`task deploy` is a stub).

Credentials are not baked into the image: `configs/prod.yaml` ships only the `BOT_TOKEN`/`BOT_ID` placeholders. Start the container with `-e BOT_TOKEN -e BOT_ID` (as `task run` does), otherwise it exits at startup.
