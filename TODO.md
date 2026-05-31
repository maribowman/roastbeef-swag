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

## Task 9 — Fix `UpdateItemsFromModal` `updatedItems` slice initialization

**Goal:** `updatedItems` should be a zero-length, pre-allocated slice — not a slice of `len(items)` zeros that we then `append` to.

**Why:** `make([]int, len(items))` allocates N zeros; the subsequent `append` adds beyond those. `slices.Contains` still works only because `0` happens not to collide with valid 1-based indices. Confusing; will bite when someone adds 0-based logic.

**Files:**
- `app/service/pantry_handling.go:75`

**Change:** `updatedItems := make([]int, 0, len(items))`.

**Acceptance:** `go test ./...` passes (existing tests cover this path).

**Out of scope:** Anything else in this function (see Task 8 for the bounds check).

---

## ~~Task 10 — Add a missing `int.yaml` config (or remove `int` from the allowlist)~~ ✓ DONE

Dropped `"int"` from the `slices.Contains` allowlist in `app/config/config.go:27`. No integration environment exists, so the option was removed.

---

## ~~Task 11 — Collapse `GroceryHandler` and `FreezerHandler` into one type~~ ✓ DONE

Replaced both handlers with a single `PantryHandler` parameterized by `tableName`, `dateFormat`, and `modalTitle`. `app/service/grocery_handler.go` and `app/service/freezer_handler.go` were deleted; `app/service/pantry_handler.go` contains the unified type. `NewDiscordBot` now constructs one type with per-channel params.

---

## Task 12 — Allowlist the SQLite table name in `SqlitePantryClient`

**Goal:** Reject table names that aren't in a known-good set, since they're interpolated into SQL via `fmt.Sprintf`.

**Why:** Today only `groceries`, `freezer`, and `unit_tests` are passed in, so this is a defense-in-depth task, not an exploitable vuln. But interpolating arbitrary strings into DDL is a footgun.

**Files:**
- `app/repository/sqlite_pantry_client.go:17-24` (`NewSqlitePantryClient`)

**Change:** Add a constant `var allowedTableNames = map[string]bool{"groceries": true, "freezer": true, "unit_tests": true}` and `log.Fatal` (or return an error) if `tableName` isn't in the set.

**Acceptance:** Constructor rejects unknown table names; existing tests still pass.

**Out of scope:** Switching to a query builder, parameterized DDL (not supported by SQLite for identifiers anyway), or schema migrations.

---

## Task 13 — Make `time.Now()` in tests stable

**Goal:** Service tests should not depend on the wall clock.

**Why:** `app/service/pantry_handling_test.go` compares structs containing `Date: time.Now().Truncate(time.Minute)` between the expected and actual values. If the test crosses a minute boundary, `time.Now()` in expected and `time.Now()` in `generateNewPantryItem` will differ.

**Files:**
- `app/service/pantry_handling.go:223-242` (`generateNewPantryItem`) and `:113-167` (`UpdateItems` adds via this).
- `app/service/pantry_handling_test.go`

**Change:** Introduce a package-level `var now = time.Now` (lowercase — internal). Use `now()` in `generateNewPantryItem`. In tests, override with `now = func() time.Time { return fixed }` in a `t.Cleanup` that restores it.

**Acceptance:** Tests pass when run repeatedly; no `time.Now()` calls in the production add path.

**Out of scope:** Introducing a `Clock` interface, dependency injection, or anything beyond the package-private var.

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
