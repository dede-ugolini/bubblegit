# bubblegit

A terminal UI for git, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). It shells out to the real `git` CLI for every operation — no git library required.

## Requirements

- **Go 1.26+**
- **git** on `PATH`
- **[delta](https://github.com/dandavison/delta)** (optional, for enhanced diff rendering)

## Install

```sh
git clone https://github.com/yourname/bubblegit.git
cd bubblegit
go build -o bubblegit .
```

## Usage

Run `bubblegit` from inside any git repository:

```sh
bubblegit
```

or

```sh
go run .
```

The TUI operates on the git repo in the current working directory.

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `Tab` | Cycle panels |
| `1`-`4` | Jump to panel (1=Files, 2=Branch, 3=Log, 4=Stash) |
| `0` | Jump to diff panel |
| `+` / `-` | Toggle fullscreen / restore |
| `t` | Cycle theme (system → nord) |
| `V` | Toggle diff renderer (delta ↔ plain git) |
| `q` | Quit |

### Files / Staging

| Key | Action |
|-----|--------|
| `j`/`k` or `↑`/`↓` | Move |
| `Space` | Stage / unstage file |
| `a` | Stage / unstage all |
| `d` | Restore file (with confirm) |
| `c` | Commit |
| `A` | Amend HEAD (with confirm) |

### Branches

| Key | Action |
|-----|--------|
| `j`/`k` or `↑`/`↓` | Move |
| `Enter` | Checkout branch |
| `n` | Create new branch |
| `r` | Rename branch |
| `d` | Delete branch (with confirm) |
| `M` | Merge (choose: fast-forward / merge commit / squash) |
| `P` | Push to origin (sets upstream) |
| `R` | Set / repoint "origin" remote URL |

### Log

| Key | Action |
|-----|--------|
| `j`/`k` or `↑`/`↓` | Move |
| `PgUp`/`PgDown` | Scroll |
| `r` | Reword commit |
| `S` | Squash marked commit range |
| `d` | Drop last commit (with confirm) |
| `T` | Create annotated tag at commit |
| `Esc` | Cancel range selection |

### Stash

| Key | Action |
|-----|--------|
| `j`/`k` or `↑`/`↓` | Move |
| `Enter` | Apply stash |
| `n` | Push new stash (with optional message) |
| `p` | Pop stash (with confirm) |
| `b` | Create branch from stash |
| `d` | Drop stash (with confirm) |
| `D` | Clear all stashes (with confirm) |

## Themes

Press `t` to cycle between themes:

- **system** — uses your terminal's ANSI palette
- **nord** — [Nord](https://www.nordtheme.com/) color palette

## Diff Rendering

Press `V` to toggle between:

- **delta** (default) — uses [delta](https://github.com/dandavison/delta) with line numbers and syntax highlighting
- **git** — plain `git diff --color=always`

In fullscreen mode (`+`), delta also enables side-by-side view.

## Architecture

```
bubblegit/
├── main.go                 # entry point
└── internal/
    ├── git/                 # git CLI wrapper (only place that shells out)
    │   ├── git.go           # status, merge
    │   ├── files.go         # stage/unstage/restore
    │   ├── diff.go          # diff (+ delta variants)
    │   ├── commit.go        # commit, amend, reword, squash, drop
    │   ├── branch.go        # list, checkout, create, delete, rename, push
    │   ├── log.go           # log parsing
    │   ├── stash.go         # stash operations
    │   ├── tag.go           # annotated tags
    │   ├── remote.go        # remote management
    │   ├── show.go          # commit display (+ delta variant)
    │   └── git_test.go      # integration tests (real git, not mocks)
    └── ui/                  # Bubble Tea front-end
        ├── model.go         # Model struct, messages, Refresh()
        ├── update.go        # Update() + action handlers
        ├── view.go          # View() + renderers
        └── theme.go         # theme definitions
```

All git operations use structured output parsing (`\x00`, `\x1f`, `\x1e` separators) for unambiguous results. Tests create temporary repos and run real git commands — no mocks.

## Testing

```sh
go test ./...                     # all tests
go test ./internal/git -run Test  # run a single test
go vet ./...                      # static analysis
```

## License

MIT
