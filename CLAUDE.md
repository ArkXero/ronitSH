# ronit.sh

SSH-accessible portfolio site. Single Go binary, Wish v2 + Bubble Tea v2 + Lip Gloss v2 + Glamour v2, SQLite for guestbook + logs, systemd-deployed on Ubuntu.

Visitors run: `ssh -p 422 ronit.sh`

## Critical facts

- Wish v2 import: `charm.land/wish/v2` (NOT github.com/charmbracelet/wish)
- Bubble Tea v2 import: `charm.land/bubbletea/v2`
- Lip Gloss v2 import: `charm.land/lipgloss/v2`
- Glamour v2 import: `charm.land/glamour/v2`
- Bubbles v2 import: `charm.land/bubbles/v2`
- Huh v2 import: `charm.land/huh/v2`
- Log v2 import: `charm.land/log/v2`
- SSH import: `github.com/charmbracelet/ssh`
- SQLite: `modernc.org/sqlite` only (pure Go, no CGo) -- driver name is "sqlite"
- SSH port: 422, not 22 (dev port: 23234 on localhost)
- Deploy target: Ubuntu at 192.168.0.249 as user `termfolio`
- Content is //go:embed'd markdown -- editing content requires a rebuild
- Binary names: ronit-sh (server), ronit-sh-ctl (admin CLI)
- Module path: github.com/ArkXero/termfolio

## Bubble Tea v2 API changes from v1 (do NOT use v1 patterns)

- Model.Init() returns only tea.Cmd (not (Model, Cmd))
- Model.View() returns tea.View (not string) -- use tea.NewView("content")
- AltScreen is set on the View struct: v.AltScreen = true (no WithAltScreen option)
- Env vars arrive as tea.EnvMsg (not os.Getenv)
- Wish v2 teaHandler signature: func(ssh.Session) (tea.Model, []tea.ProgramOption)
- lipgloss.AdaptiveColor does not exist in v2 -- use lipgloss.Color(hex) or lipgloss.LightDark(isDark)

## Commands

- `make dev`    -- run locally on :23234
- `make build`  -- cross-compile for linux/amd64 into dist/
- `make test`   -- run tests
- `make lint`   -- go vet
- `make deploy` -- build + scp + restart remote service (asks confirmation)

## Environment variables

- `RONIT_ENV=production` -- binds 0.0.0.0:422, uses /var/lib/termfolio/id_ed25519 for host key
- (default dev) -- binds 127.0.0.1:23234, uses .ssh/id_ed25519

## Architecture

- Single root Model (internal/tui/root.go) with View enum controlling which sub-model is active
- Sub-models: BannerModel, MenuModel, AboutModel, ProjectsModel, NowModel, PostsModel, GuestbookModel, ContactModel
- Sub-models return navigateMsg via tea.Cmd to request view transitions -- root intercepts these
- Responsive layout: sidebar+content (>=80 cols), top-nav+content (<80), error (<40)
- AltScreen declared in View() via v.AltScreen = true

## Conventions

- ASCII only in generated content (no em dashes, no smart quotes)
- Complete file contents when editing, not snippets
- Serialize SQLite access via sync.Mutex + SetMaxOpenConns(1)
- Never commit .ssh/ keys, *.db files, or secrets
- All user input (guestbook) MUST pass through internal/sanitize before storage

## See also

- `.claude/spec.md` -- full project specification, source of truth
- `TODO.md` -- backlog and v1.1+ ideas
- `internal/tui/root.go` -- start here to understand the TUI
