# TODO — Review-Driven Tasks

Tasks discovered during a codebase review on the `claude-review` branch (2026-05-22).
Sorted by impact (highest first). Each task is scoped to stay small and isolated — do **not** combine tasks or expand into a broader refactor unless the task explicitly says so.

When picking up a task in a new session, say: *"work on task N from TODO.md"*.

---

## ~~Task 1 — Fix `UpdateItems` fall-through after `RemoveItem`~~ ✓ DONE

Restructured the `editRegex` branch in `UpdateItems` to `if/else` so `UpdateItem` is no longer called after `RemoveItem` when `updatedQuantity ≤ 0`. Added a `"quantity update to zero removes item"` test case (`1--1` on a qty-1 item) that asserts the row is gone — no resurrected row with `Quantity: 0`.

---

## ~~Task 2 — Fix `PublishItems` dropping the tail of long tables~~ ✓ DONE

Extracted the chunking into a pure `splitMarkdownTable(table string, max int) []string` helper that splits the fenced `` ```md `` table on line boundaries, re-wrapping every chunk as its own code block. Reworked the split branch in `PublishItems` to iterate **all** chunks (edit the first into `messageID`, send the rest as new messages) instead of sending one chunk and `return`ing — fixing the silent tail-drop. Dropped the bogus `...``` ` truncation marker; the action buttons now ride only on the final chunk (and are cleared from the edited first chunk when it isn't last). Header repeats only on the first chunk (per decision). Added `TestSplitMarkdownTable` asserting no rows are dropped, every chunk is ≤ max and properly fenced, and no truncation marker remains.

---

## ~~Task 3 — Unify the channel-ID source in `FreezerHandler.MessageEvent`~~ ✓ DONE

Switched `FreezerHandler.MessageEvent` to use `handler.channelID` for the bulk delete, matching the grocery handler.

---

## ~~Task 4 — Move the Discord bot token out of the image~~ ✓ DONE

Stopped baking secrets into the image. Removed the `sed` injection + `git restore` from the `build` task (`Taskfile.yaml`) and deleted the "Inject bot token" step from `.github/workflows/build.yml`, so the published image now ships `configs/prod.yaml` with only the literal `BOT_TOKEN`/`BOT_ID` placeholders. Reworked `loadAndReplaceFromDotEnv` (`app/config/config.go`) to treat `.env` as an optional local-dev convenience and always read `BOT_TOKEN`/`BOT_ID` from the process environment, overriding the YAML only when a value is set (non-empty guards keep `go test` unaffected). The `run` task passes `-e BOT_TOKEN -e BOT_ID` through to `docker run`.

**Deployment follow-ups (user action):** the Synology NAS run must now supply `-e BOT_TOKEN`/`-e BOT_ID`, and the existing bot token should be **rotated** (it shipped inside published GHCR layers) along with purging old images.

**Out of scope (unchanged):** Docker secrets, vault, or any external secret manager. Just env vars.

---

## ~~Task 5 — Stop double-processing in `ReadyEvent`~~ ✓ DONE

Removed the redundant second `PreProcessMessageEvent` + `UpdateItems` block from both handlers' `ReadyEvent`. The single `MessageEvent` call now runs the full pre-process / update / publish cycle once on startup. The optional synthetic-`MessageCreate` cleanup was skipped — replacing the `Author.ID == "init"` placeholder requires extracting `MessageEvent`'s body, which became unnecessary once Task 11 collapsed both handlers.

---

## ~~Task 6 — Guard `defer stmt.Close()` against nil statements~~ ✓ DONE

Added a `return` to the `Prepare`-error branch of `UpdateItem`, `RemoveItem`, and both `Prepare` calls in `RemoveAllItems` (`app/repository/sqlite_pantry_client.go`), so a failed prepare (which returns a nil `*sql.Stmt`) never reaches `defer stmt.Close()` and panics on a nil receiver. `AddItem` and `GetItems` already returned on prepare error — left unchanged. The `RemoveAllItems` variable-reuse cleanup remains for Task 7 (not bundled); the added early return is safe because `defer stmt.Close()` binds its receiver value when the defer executes.

---

## ~~Task 7 — Fix `RemoveAllItems` double-defer / leaked statement~~ ✗ INVALID (no bug)

**Verified during the Task 8 planning pass (2026-05-31):** the premise is wrong. `defer
stmt.Close()` is a *method value* whose receiver is evaluated **when the `defer` statement
executes** (Go spec: "the expression x is evaluated and saved during the evaluation of the
method value"). So the first `defer` captures the first statement's value and the second
`defer` captures the second — reassigning the `stmt` variable does **not** redirect the
earlier defer. There is no double-close and no leak; both statements are closed exactly
once. The only residual is a stylistic variable-reuse footgun, which is not the described
bug — closing this task as a non-issue rather than implementing a fix.

---

## ~~Task 8 — Bounds-check the modal index in `UpdateItemsFromModal`~~ ✓ DONE

Guarded the index in the `len(matches) == 2` branch of `UpdateItemsFromModal` (`app/service/pantry_handling.go`): when the parsed `[N]` is `< 1` or `> len(items)`, the line is logged at debug and added as a new item (`AddItem` + `continue`) instead of indexing `items[index-1]` out of range and panicking the interaction goroutine. Added the `"out-of-range modal index adds a new item"` test case (`modalInput: "[1] milk\n[5] foo"` → original `milk` preserved + `foo` added). Used `[1] milk\n[5] foo` rather than the TODO's literal `[5] foo` because a lone `[5] foo` omits the valid index and would (correctly) remove the unreferenced original — the two-line input matches the TODO's stated intent of "original plus foo".

---

## ~~Task 9 — Fix `UpdateItemsFromModal` `updatedItems` slice initialization~~ ✓ DONE

Changed `updatedItems := make([]int, len(items))` to `make([]int, 0, len(items))` in `UpdateItemsFromModal` (`app/service/pantry_handling.go`). The old form pre-filled the slice with `len(items)` zeros and then appended the real 1-based indices on top, so the slice held `[0, …, 0, realIndex₁, …]`; the removal pass only worked because `slices.Contains(updatedItems, index+1)` always looks up `index+1 ≥ 1`, which never collides with the spurious zeros. The new form is zero-length with the capacity reserved, so it holds only genuinely-updated indices and no longer relies on "0 never matches." Behavior is identical — the existing table-driven `TestUpdateItemsFromModal` cases cover the path; `go vet` and `go test ./...` clean.

---

## ~~Task 10 — Add a missing `int.yaml` config (or remove `int` from the allowlist)~~ ✓ DONE

Dropped `"int"` from the `slices.Contains` allowlist in `app/config/config.go:27`. No integration environment exists, so the option was removed.

---

## ~~Task 11 — Collapse `GroceryHandler` and `FreezerHandler` into one type~~ ✓ DONE

Replaced both handlers with a single `PantryHandler` parameterized by `tableName`, `dateFormat`, and `modalTitle`. `app/service/grocery_handler.go` and `app/service/freezer_handler.go` were deleted; `app/service/pantry_handler.go` contains the unified type. `NewDiscordBot` now constructs one type with per-channel params.

---

## ~~Task 12 — Allowlist the SQLite table name in `SqlitePantryClient`~~ ✗ WON'T DO

Dismissed (2026-05-31). The table name is always internally controlled — it comes only
from the static `configs/*.yaml` channel list (`groceries`, `freezer`) and the hardcoded
`unit_tests` in tests; it is never derived from user input. The interpolated-DDL path is
therefore not reachable by an attacker, so the allowlist adds no practical protection.
Closing as won't-do rather than adding defense-in-depth code with no reachable threat.

---

## ~~Task 13 — Make `time.Now()` in tests stable~~ ✓ DONE

Introduced a package-private clock indirection `var now = time.Now` in `app/service/pantry_handling.go` and switched `generateNewPantryItem` to `now().Truncate(time.Minute)` (the only `time.Now()` in the service add path — confirmed via grep). In `app/service/pantry_handling_test.go`, replaced every `time.Now().Truncate(time.Minute)` date literal with `now().Truncate(time.Minute)`, and froze the clock at the top of `TestUpdateItemsFromModal` and `TestUpdateItems` (`fixed := time.Now(); now = func() time.Time { return fixed }`) with a `t.Cleanup(func() { now = time.Now })` restore. `fixed` is left un-truncated on purpose — both production and the literals call `now().Truncate(...)`, so they truncate the same instant to the same minute regardless of evaluation order. Verified with `go test ./app/service/... -count=20` (no flakes), full suite, and `go vet` all clean.

---

## ~~Task 14 — Add unit tests for `PreProcessMessageEvent` and the chunking logic in `PublishItems`~~ ✓ DONE

Both Discord-coupled functions now have pure, tested cores. The chunking logic was extracted into `splitMarkdownTable(table string, max int) []string` with `TestSplitMarkdownTable` (landed with Task 2). The message-classification logic was extracted out of `PreProcessMessageEvent` into a pure `partitionChannelMessages(messages []*discordgo.Message, botID string) (lastBotMessageID, userInput string, removableMessageIDs []string)` helper, covered by `TestPartitionChannelMessages` (keeps the oldest bot message, marks every other message removable, accumulates user input). The Discord-sending/fetching parts remain untested by design (no session fake).

---

## ~~Task 15 — Run CI on every push, not only tag pushes~~ ✗ WON'T DO

Dismissed (2026-05-31). The value of a separate push/PR test workflow is catching failures
before release and on others' PRs. This is a single-maintainer repo with no PR flow, and the
Dockerfile builder already runs `go test ./... -cover -v` on every tag push (release),
gating the image. The extra workflow adds little for this setup, so closing as won't-do
rather than maintaining a CI file with no practical payoff.

---

# Follow-up review (2026-10-24)

Findings from a second review of `claude-review` vs `main`, fixed in two waves of parallel work packages. Task numbers continue from above; the original finding number is given in each title.

---

## ~~Task 16 — Serialize concurrent `syncChannel` runs (finding #1)~~ ✓ DONE

Added `mu sync.Mutex` to `PantryHandler` (`app/service/pantry_handler.go`) and held it for the whole `syncChannel` call. discordgo runs every handler in its own goroutine, so two quick messages (or a message plus an Edit/Undo click) could both read the same user input before either deleted it and apply it twice. `syncChannel` now takes `(session, action func()) error`; `MessageEvent`, `ReadyEvent`, the modal submit and the Undo button all go through it, so every mutation runs under the per-channel lock. The modal submit sends its deferred response first, inside Discord's 3 s window, and only then waits on the lock.

---

## ~~Task 17 — Abort the sync when deleting messages fails (finding #2)~~ ✓ DONE

Added `deleteMessages(session, channelID, ids) error` (`app/service/pantry_handler.go`): it tries `ChannelMessagesBulkDelete` and, on error, falls back to `ChannelMessageDelete` per message (which also covers messages older than 14 days, which bulk delete rejects). `syncChannel` deletes **before** applying any input and returns the error, without mutating anything, if any message still can't be deleted. Previously a failed bulk delete was only logged, so the same user messages were re-applied on every sync. Trade-off (documented in the code): input from messages already deleted in the fallback is logged and dropped instead of being re-applied forever.

---

## ~~Task 18 — Keep the Edit/Undo buttons when the table shrinks to one message (finding #3)~~ ✓ DONE

The single-message **edit** path in `PublishItems` (`app/service/pantry_handling.go`) now always sets `Components` to `CreateMessageButtons()`. Before, a message reused from the split path kept its cleared components, and the last chunk holding the buttons was deleted, so the buttons were gone for good. One-line fix, verified by review (no session fake).

---

## ~~Task 19 — Fix the slice panic and UTF-8 splitting in `processItemMarkdownLayout` (finding #4)~~ ✓ DONE

Reworked the name wrapping in `app/model/pantry.go` to count in runes (`[]rune`, `utf8.RuneCountInString`) instead of bytes, so umlauts are never cut mid-character. When an over-long word meets a line with no room left (`charsLeft <= 0`), the current line is flushed and the word starts on a fresh line, instead of panicking on `split[:charsLeft]` (a crash loop, since the item was already persisted and every restart re-rendered it). The fixed "split a second time" step became a loop that keeps splitting until the remainder fits. Existing `TestToMarkdownTable` expectations are unchanged; added `TestProcessItemMarkdownLayout` covering the review's panic input, a first word that exactly fills the line, a word longer than 3× the line break, and umlaut names (asserting `utf8.ValidString` per row).

---

## ~~Task 20 — Apply batched user messages oldest first (finding #5)~~ ✓ DONE

`partitionChannelMessages` (`app/service/pantry_handling.go`) now collects the user messages, sorts them with `compareMessagesChronologically` (timestamp, then snowflake ID compared numerically by length then lexicographically) and joins them oldest first. `ChannelMessages` returns newest first, so e.g. `butter` followed by `1--` used to run as `1--\nbutter`. Bot-message selection is unchanged. Extended `TestPartitionChannelMessages` with newest-first, three-message and equal-timestamp cases.

---

## ~~Task 21 — Stable order for items added in the same minute (finding #6)~~ ✓ DONE

`GetItems` (`app/repository/sqlite_pantry_client.go`) now orders by `date asc, id asc`. Dates are truncated to the minute, so items added together had no defined order and a later `2` or `2--` could hit a different item than the one shown. Covered by `TestGetItemsOrdersTiedDatesByID` in the new `app/repository/sqlite_pantry_client_test.go`.

---

## ~~Task 22 — Fail fast on missing `BOT_TOKEN`/`BOT_ID` (finding #7)~~ ✓ DONE

Added a pure `validateDiscordConfig(DiscordConfig) error` in `app/config/config.go` that rejects an empty or still-placeholder (`"BOT_TOKEN"` / `"BOT_ID"`) token or bot ID, with a message telling the operator to set the env var. `loadConfig` calls it after the env override and `log.Fatal`s on error, except under the `test` config (`configs/test.yaml` has no discord section). Before, a container started without `BOT_ID` kept the literal placeholder, treated the bot's own table as user input and filled the DB with garbage. Covered by table-driven `TestValidateDiscordConfig` in the new `app/config/config_test.go`.

---

## ~~Task 23 — Stop swallowing startup errors and drop the fatal Ready timeout (finding #8)~~ ✓ DONE

`syncChannel` now returns its error, and `ReadyEvent` returns it directly instead of faking a `MessageCreate`. `DiscordBot.Ready` (`app/service/discord_bot.go`) was rewritten without the 5 s timeout, `log.Fatal` and closable `errorChannel`: it starts one goroutine per handler that logs its own error with the channel ID, waits on a `sync.WaitGroup`, and logs either "All handlers initialized" or "N of M handlers failed to initialize". Ready also fires on reconnects, so it no longer kills the process on a slow Discord API and can't panic sending on a closed channel.

---

## ~~Task 24 — Make the Undo button actually undo (finding #9)~~ ✓ DONE

Storage: `model.PantryClient` gained `CreateSnapshot()` and `RestoreSnapshot() bool`, backed by a shared `snapshots(table_name text primary key, items text not null)` table with the items JSON-encoded per pantry table (`app/repository/sqlite_pantry_client.go`). A present row means a snapshot exists, so a snapshot of an empty table (after `*`) is still restorable. `RestoreSnapshot` runs in a single transaction (read snapshot, delete all rows, re-insert with original ids and dates, delete the snapshot) and returns `false` without changes if no snapshot exists. `MockPantryClient` gained the two methods plus the missing `RemoveAllItems`. Covered by `TestSnapshotRestore`, `TestRestoreSnapshotWithoutSnapshot` and `TestSnapshotsAreIsolatedPerTable`.

Behavior: a snapshot is taken right before each mutating action (a sync that applies user message input, or a modal submit). The Undo button sends the deferred ack and then restores the snapshot inside `syncChannel`, logging "Nothing to undo" if there is none. Undo reverts exactly the last action, once: it never takes a snapshot itself, so a second click is a no-op. Syncs with no input and no action (startup, empty channel) never touch the snapshot.

---

## ~~Task 25 — Bring `AGENTS.md` / `CLAUDE.md` back in line with the code (finding #10)~~ ✓ DONE

Updated `AGENTS.md` (`CLAUDE.md` stays a symlink to it): config choices are `local`, `prod` and anything containing `test` (`int` was removed); `BOT_TOKEN`/`BOT_ID` come from the environment with `.env` optional, and startup fails fast on missing or placeholder values outside `test`; deployment no longer mentions `sed` injection, and the container needs `-e BOT_TOKEN -e BOT_ID`. The Discord event flow section now documents the per-channel lock, delete-before-apply with the bulk/per-message fallback and abort, oldest-first input, `date, id` ordering, the non-fatal `Ready`, and single-step Undo via the `snapshots` table.

---

## ~~Task 26 — Lower-priority review notes~~ ✓ DONE

- **Duplicate rendering:** the single-message send path in `PublishItems` reuses the already rendered `markdownTable` instead of calling `ToMarkdownTable` a second time.
- **Shutdown order:** `main.go` now closes the Discord session first, then shuts down the HTTP server, then closes the DB, so in-flight handlers never hit a closed database.

---

## Notes for future-Claude

- These tasks are intentionally **small and independent**. Do not bundle. If a task tempts you into a broader refactor, stop and ask the user.
- File paths use the layout as of branch `claude-review`. If something has moved, re-grep before editing.
- The codebase uses `zerolog`, table-driven tests, real in-memory SQLite (no mocks). Match the existing style.
- `CLAUDE.md` in the repo root has the canonical architectural overview — re-read it before any task that touches multiple packages.
