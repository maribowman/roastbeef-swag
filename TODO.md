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

## Task 15 — Run CI on every push, not only tag pushes

**Goal:** `.github/workflows/build.yml` runs tests (at minimum) on push to `main` and on PRs.

**Why:** Today CI only runs on tag push. Test failures are only surfaced at release time. The Dockerfile builder runs `go test`, but that's expensive feedback compared to a regular CI run.

**Files:**
- `.github/workflows/build.yml:3-6`

**Change:** Add a separate lightweight workflow (or extend the existing one) with `on: { push: { branches: [main] }, pull_request: {} }` that runs `go test ./...`. Leave the image build on tag-only.

**Acceptance:** PRs and pushes to `main` show a green/red check based on `go test`.

**Out of scope:** Adding lint, coverage upload, matrix builds, or release automation changes.

---

## Notes for future-Claude

- These tasks are intentionally **small and independent**. Do not bundle. If a task tempts you into a broader refactor, stop and ask the user.
- File paths use the layout as of branch `claude-review`. If something has moved, re-grep before editing.
- The codebase uses `zerolog`, table-driven tests, real in-memory SQLite (no mocks). Match the existing style.
- `CLAUDE.md` in the repo root has the canonical architectural overview — re-read it before any task that touches multiple packages.
