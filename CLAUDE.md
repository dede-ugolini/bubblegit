# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`bubblegit` is a terminal UI for git, built with Charm's Bubble Tea (`charm.land/bubbletea/v2`). It shells out to the real `git` CLI (and optionally `delta`) for every operation rather than using a Git library.

## Commands

```sh
go build -o bubblegit .   # build the binary (also gitignored as /bubblegit)
go run .                  # run against the git repo in the current directory
go test ./...             # run all tests
go test ./internal/git -run TestMerge   # run a single test
go vet ./...
```

`internal/git` tests are integration tests: they call `t.TempDir()`, run real `git init`/`git commit` via `exec.Command` (see `initRepo`/`run` helpers in `git_test.go`), and assert against actual repo state — not mocks. A working `git` binary must be on `PATH`. When adding a git operation, add a matching integration test using the same `initRepo`/`run` pattern.

## Architecture

Three packages:

- **`internal/git`** (`git.go`) — the only place that shells out to `git`/`delta`. Every function takes a `dir string` (the repo path) as its first argument and runs `exec.Command` with `cmd.Dir = dir`; errors from failed commands are wrapped as `fmt.Errorf("%s", strings.TrimSpace(string(out)))` so CLI stderr reaches the UI. No package-level state. Structured output (status, branches, log, stashes) is parsed from `git` invoked with NUL/unit/record separators (`\x00`, `\x1f`, `\x1e`) or custom `--pretty=format:` strings rather than plain-text output, to make parsing unambiguous.
- **`internal/ui`** — the Bubble Tea `Model`, split by Elm-architecture convention:
  - `model.go`: `Model` struct, message types (`filesMsg`, `branchesMsg`, `logMsg`, `stashesMsg`, `diffMsg`, `errMsg`, `tickMsg`), popup state structs, and `Refresh()`, which issues the git commands as batched `tea.Cmd`s.
  - `update.go`: `Update()` and one `handleXxx` method per git action (e.g. `handleCommit`, `handleMerge`, `handleRewordCommit`). Each wraps a git package call and returns a `tea.Msg`/`tea.Cmd` that triggers a `Refresh()`.
  - `view.go`: `View()` and one `renderXxx` method per panel/popup.
  - `theme.go`: theme definitions and switching (`nextTheme`).
- **`main.go`** — thin entry point: gets the working directory, constructs `ui.NewModel(dir)`, runs the `tea.Program`.

### UI structure

The app has five focusable panels cycled via `focus` (`focusStag`, `focusBranch`, `focusLog`, `focusStash`, `focusDiff`), each with its own `idxXxx`/`xxxHeight`/`xxxWidth` fields on `Model`, and a diff viewport that updates on every focus/selection change. Multiple modal popups exist for actions needing input (commit message, tag, branch rename/create, merge mode, stash branch, stash-clear confirm) — each is its own struct (`commitPopup`, `tagPopup`, `mergePopup`, `inputPopup`, `stashBranchPopup`) with an `active` bool, and `inputPopup`/`commitPopup` are reused across multiple related actions distinguished by an `action`/`reword` field rather than having a separate popup type per action.

Diffs can render via plain `git diff --color=always` or through `delta` (external command) — see `useDelta` on `Model` and the `XxxDelta` variants of diff functions in `internal/git`.

## Feature tracking

`TODO.md` tracks git features as a checklist grouped by area (Files, Branch, Log, Commit, Stash, Remote, Tags). Check it before adding a new git action, and check an item off when you implement it.
