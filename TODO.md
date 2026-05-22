# TODO — Review-Driven Tasks

Tasks discovered during a codebase review on the `claude-review` branch (2026-05-22).
Sorted by impact (highest first). Each task is scoped to stay small and isolated — do **not** combine tasks or expand into a broader refactor unless the task explicitly says so.

When picking up a task in a new session, say: *"work on task N from TODO.md"*.

---

## Task 1 — Fix `UpdateItems` fall-through after `RemoveItem`

**Goal:** When edit-quantity drops an item to ≤0, only remove it. Do not also update it with the non-positive quantity.

**Why:** Silent correctness bug. The branch calls `RemoveItem` then falls straight into `UpdateItem` with `updatedQuantity ≤ 0`. The `UPDATE` no-ops because the row is gone, but the intent is clearly wrong and the next person reading this will mis-trust the code.

**Files:**
- `app/service/pantry_handling.go:124-135` (inside `UpdateItems`, the `editRegex.MatchString(line)` branch)

**Change:** After `pantryClient.RemoveItem(item.ID)`, `continue` the inner loop (or restructure to `if/else`) so `UpdateItem` is not called for the removed row.

**Acceptance:**
- `go test ./...` still passes (the existing `"advanced negative quantity update exception"` case in `pantry_handling_test.go:122` covers this and currently passes by accident — confirm it still passes for the right reason).
- Add one table-driven case where an edit reduces qty to 0 (e.g., `1--1` on a qty-1 item) and asserts the item is gone, no resurrected row.

**Out of scope:** Refactoring the wider `UpdateItems` function, touching the regex, restructuring the loop.

---

## Task 2 — Fix `PublishItems` dropping the tail of long tables

**Goal:** When the markdown table exceeds 2000 chars, send all chunks instead of only the first.

**Why:** Silent data loss. The "split" branch sends/edits one chunk then `return`s inside the loop, so any rows beyond ~1980 chars are never published. Also the chunk currently appends `...```` to the body even when the table fits a single chunk on the split path.

**Files:**
- `app/service/pantry_handling.go:247-292` (`PublishItems`)

**Change:**
- Rework the split branch to iterate all chunks: edit the first chunk into `messageID` (if non-empty), then send the remaining chunks as new messages. Only the final chunk should carry the action buttons (`CreateMessageButtons()`).
- Remove the `...```` truncation marker (or only append it if a chunk was actually truncated mid-line — which shouldn't happen if you split on line boundaries).
- Make sure each chunk re-opens with ` ```md ` and closes with ` ``` ` so Discord renders the code block.

**Acceptance:**
- Add a unit test for a helper that splits a markdown table string into ≤2000-char chunks on line boundaries (extract the chunking logic so it can be tested without a Discord session).
- Manual smoke (or note in the PR) that a list with 60+ items still renders fully.

**Out of scope:** Adding pagination buttons, persisting message IDs across multi-message tables, supporting >10 chunks.

---

## Task 3 — Unify the channel-ID source in `FreezerHandler.MessageEvent`

**Goal:** Use `handler.channelID` (not `message.ChannelID`) in the freezer's bulk delete, matching the grocery handler.

**Why:** `ReadyEvent` calls `MessageEvent` with a synthetic `discordgo.MessageCreate{Message: &discordgo.Message{Author: ...}}` that has an empty `ChannelID`. The freezer handler then calls `ChannelMessagesBulkDelete("", ...)` on startup, which errors. The grocery handler already uses `handler.channelID` and works.

**Files:**
- `app/service/freezer_handler.go:46`

**Change:** Replace `message.ChannelID` with `handler.channelID`.

**Acceptance:** Repro check (read-only): with the change applied, both handlers use `handler.channelID` everywhere. `grep "message.ChannelID" app/service/` returns nothing.

**Out of scope:** The bigger handler-unification refactor (see Task 11).

---

## Task 4 — Move the Discord bot token out of the image

**Goal:** Stop baking `BOT_TOKEN` / `BOT_ID` into `configs/prod.yaml` at build time. Read them from runtime env vars instead.

**Why:** Anyone with pull access to `ghcr.io/maribowman/roastbeef-swag` currently has the production bot token. The `sed` substitution in `Taskfile.yaml` and `.github/workflows/build.yml` writes secrets into a file that the Dockerfile `COPY /configs /configs` ships in the final image layer.

**Files:**
- `Taskfile.yaml:23-24, 27` — remove the `sed` lines and the `git restore`.
- `.github/workflows/build.yml:29-32` — remove the "Inject bot token" step.
- `app/config/config.go:73-82` (`loadAndReplaceFromDotEnv`) — read `BOT_TOKEN` / `BOT_ID` from `os.Getenv` unconditionally (not only when `.env` exists). Keep `.env` loading for local dev convenience.
- `configs/prod.yaml:10-11` — leave the placeholder values; they will be overwritten by env at runtime.
- `Taskfile.yaml: run` task — pass `-e BOT_TOKEN -e BOT_ID` to `docker run` so the local run still works.

**Acceptance:**
- `docker build` produces an image where `grep -r '<your real token>' /configs` returns nothing.
- Running the image with `docker run -e BOT_TOKEN=... -e BOT_ID=... <image> prod` still connects.
- Old images on GHCR should be rotated (note this in the PR description — the user does that, not Claude).

**Out of scope:** Switching to Docker secrets, vault, or any external secret manager. Just env vars.

---

## Task 5 — Stop double-processing in `ReadyEvent`

**Goal:** `ReadyEvent` should run the channel pre-process + update + publish exactly once on startup.

**Why:** Currently `ReadyEvent` calls `handler.MessageEvent(...)` with a synthetic event (which already does the full pre-process / update / publish cycle), and then immediately calls `PreProcessMessageEvent` and `UpdateItems` again. The second pass is mostly seeing the bot's own freshly-published table and re-processing it. Wastes Discord API calls; racy if a user posts in the gap.

**Files:**
- `app/service/grocery_handler.go:27-37`
- `app/service/freezer_handler.go:27-37`

**Change:** Delete the second `PreProcessMessageEvent` + `UpdateItems` block. The single `MessageEvent` call (or a small dedicated startup function) is sufficient.

Optional follow-up *within the same task*: replace the synthetic `MessageCreate` with a direct call to the underlying logic (extract the body of `MessageEvent` into a helper that takes only the session), so the fake `Author.ID == "init"` hack goes away. Only do this if the diff stays under ~30 lines per handler.

**Acceptance:**
- `go test ./...` passes.
- `ReadyEvent` no longer constructs a `discordgo.MessageCreate` (if you took the optional cleanup).

**Out of scope:** Merging the two handlers (Task 11). Snapshot/undo work.

---

## Task 6 — Guard `defer stmt.Close()` against nil statements

**Goal:** Stop the SQLite repository from panicking when `db.Prepare` returns an error.

**Why:** Every `Prepare` call in `sqlite_pantry_client.go` logs on error but still proceeds to `defer stmt.Close()`. If `stmt` is nil, the deferred call panics (`*sql.Stmt.Close` on a nil receiver dereferences `s.cg`). Today this only fires if SQLite itself errors on prepare, which is rare but observable on schema drift or closed connections.

**Files:**
- `app/repository/sqlite_pantry_client.go:50-60` (`UpdateItem`)
- `app/repository/sqlite_pantry_client.go:62-72` (`RemoveItem`)
- `app/repository/sqlite_pantry_client.go:74-95` (`RemoveAllItems`)
- `app/repository/sqlite_pantry_client.go:97-124` (`GetItems`)
- `app/repository/sqlite_pantry_client.go:33-48` (`AddItem`) — already returns on prepare error; double-check.

**Change:** On `Prepare` error, `return` (or `return zero, err` for `AddItem`/`GetItems`) before the `defer`. Matches the existing `AddItem` pattern.

**Acceptance:** All `Prepare` call sites either return immediately on error or guard the defer.

**Out of scope:** Changing the SQL itself, switching to `sqlx`, or introducing transactions.

---

## Task 7 — Fix `RemoveAllItems` double-defer / leaked statement

**Goal:** Both prepared statements in `RemoveAllItems` are closed exactly once.

**Why:** The function prepares two statements but reuses the `stmt` variable. The first `defer stmt.Close()` captures the variable, not the value, so at function return both defers close the *second* statement. The first statement leaks until the connection closes.

**Files:**
- `app/repository/sqlite_pantry_client.go:74-95`

**Change:** Use two distinct variable names (e.g., `deleteStmt`, `resetStmt`), or scope each prepare/exec in its own function. Each statement gets its own `defer`.

**Acceptance:** `go vet ./...` clean; visual confirmation that each statement has its own `defer` bound to its own variable.

**Out of scope:** Switching to `Exec` directly without `Prepare` (functionally equivalent here and would be a tiny improvement, but keep this task minimal).

---

## Task 8 — Bounds-check the modal index in `UpdateItemsFromModal`

**Goal:** Submitting `[99]` on a 3-item list should not panic.

**Why:** `items[index-1]` at `pantry_handling.go:91` panics out of range. The Discord interaction goroutine then dies silently — no error returned to the user, no recovery.

**Files:**
- `app/service/pantry_handling.go:73-111` (`UpdateItemsFromModal`)

**Change:** When the parsed `index` is `< 1` or `> len(items)`, treat the line as a new item (i.e., fall through to the `else` add path). Log a debug line so the behavior is discoverable.

**Acceptance:**
- Add a test case: 1 existing item, modal input `[5] foo` → result has the original item plus `foo` as a new item.
- `go test ./...` passes.

**Out of scope:** Returning a user-visible error message via Discord, reworking the modal protocol, fixing the `slices.Contains` initialization (Task 9).

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

## Task 10 — Add a missing `int.yaml` config (or remove `int` from the allowlist)

**Goal:** Either `./roastbeef-swag int` works, or `int` is no longer a documented option.

**Why:** `app/config/config.go:27` accepts `int` as a config target but no `configs/int.yaml` exists. Running with `int` calls `log.Fatal` on the missing file.

**Files:**
- `app/config/config.go:27`
- `configs/` (new file or removal of `"int"` from the allowlist)

**Change:** Decide with the user first (ask in the session). Likely the right move is to drop `int` from the allowlist since there's no integration environment. If kept, copy `prod.yaml` and adjust.

**Acceptance:** Either `configs/int.yaml` exists and `./roastbeef-swag int` boots, or `int` is gone from `loadConfig`.

**Out of scope:** Reworking the config loader, switching off YAML.

---

## Task 11 — Collapse `GroceryHandler` and `FreezerHandler` into one type

**Goal:** A single `PantryHandler` parameterized by `{tableName, dateFormat, modalTitle}` replaces both files.

**Why:** The two files differ in ~5 string literals and one date format. The duplication is what caused Task 3 (drift between the two channel-ID usages). Removing it prevents future drift.

**Files:**
- `app/service/grocery_handler.go` — delete after migration.
- `app/service/freezer_handler.go` — delete after migration.
- New `app/service/pantry_handler.go` (or rename one of the above).
- `app/service/discord_bot.go:26-36` — `NewDiscordBot` switch now constructs one type with different params.
- `app/service/pantry_handling.go:16-18` — the `GroceriesChannel`/`FreezerChannel` constants stay; they're still used as routing keys.

**Change:**
- Define `PantryHandler` with fields `channelID`, `pantryClient`, `lineBreak`, `dateFormat`, `modalTitle`.
- `NewPantryHandler(channelID string, db model.DatabaseClient, lineBreak int, tableName, dateFormat, modalTitle string)`.
- Move both handlers' methods over verbatim (they're identical aside from the param-driven strings).

**Acceptance:**
- `go test ./...` passes.
- `grep -r "GroceryHandler\|FreezerHandler" app/` returns nothing.
- The Discord behavior is unchanged (manual smoke if possible).

**Out of scope:** Changing the bot interface, snapshot/undo, modal UX changes. Do not pick this up if you're also doing Tasks 3 or 5 in the same PR — land those first, then this.

**Pre-req:** Tasks 3 and 5 should be done first (so the diff is purely structural).

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

## Task 14 — Add unit tests for `PreProcessMessageEvent` and the chunking logic in `PublishItems`

**Goal:** Cover the two most Discord-coupled functions that currently have zero tests.

**Why:** These functions hold most of the bug surface (Tasks 2, 3, 5 all touch them). Tests would have caught the tail-drop bug.

**Files:**
- `app/service/pantry_handling.go` — extract the chunking inside `PublishItems` into a pure helper (e.g., `splitMarkdownTable(table string, max int) []string`) that takes a string and returns chunks. The Discord-sending part stays untestable without a fake session, but the splitting logic is pure.
- `app/service/pantry_handling_test.go` — add tests for `splitMarkdownTable`.
- For `PreProcessMessageEvent`: extract the message-classification logic (which messages are bot/user/removable) into a pure function over `[]*discordgo.Message` so it can be tested without a session.

**Acceptance:**
- New tests pass.
- `go test ./... -cover` shows non-zero coverage on the extracted helpers.

**Out of scope:** Mocking `discordgo.Session`. The point is to extract enough pure logic to test, not to build a full Discord fake.

**Pre-req:** Task 2 should land first so the chunking helper already exists in its fixed form.

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
