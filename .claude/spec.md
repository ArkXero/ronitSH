# termfolio — Build Specification

> **To the coding agent reading this:** This is your primary reference document for building an SSH-accessible portfolio site. Treat it as the source of truth. When you're unsure, re-read the relevant section before asking the user. The user (Ronit) is a high school sophomore building this as a portfolio piece, working in Claude Code. Default to making decisions and explaining them, not asking for approval on every choice — but DO ask before any irreversible action (deploying, purchasing, deleting, force-pushing).

---

## 0. Critical constraints — read first, never violate

1. **Use Wish v2, not v1.** The import path is `charm.land/wish/v2` (NOT `github.com/charmbracelet/wish`). Wish v2 shipped in 2025 and runs on Bubble Tea v2 with the Cursed Renderer. Most tutorials online are still v1 and will mislead you. If you see `github.com/charmbracelet/wish` in an example, it is outdated — translate to the v2 import path.
2. **Use Bubble Tea v2.** Import as `charm.land/bubbletea/v2`. APIs differ from v1 (renderer, color profile handling, PTY). Same translation rule applies.
3. **Pure Go only.** Use `modernc.org/sqlite` (not `mattn/go-sqlite3`). The binary must build with `CGO_ENABLED=0` so cross-compilation for the Ubuntu server is trivial.
4. **Target platform: Ubuntu Server (amd64).** Final binary runs on `192.168.0.249`, Ronit's self-hosted box. Do NOT assume the dev machine matches — Ronit develops on both macOS (M2) and Arch Linux.
5. **SSH port is 422, NEVER 22.** Port 22 gets brute-forced by botnets on residential IPs. Never suggest binding to 22.
6. **Content is version-controlled markdown, embedded via `//go:embed`.** No database of content, no CMS. A deploy = rebuild binary.
7. **No CGo, no Docker, no Node, no Python.** Single Go binary + systemd + SQLite file. That's it.
8. **Never commit secrets.** Cloudflare API tokens, SSH host keys, and anything in `/var/lib/termfolio/` stay off git.
9. **Ronit's style preferences:** strict ASCII in generated text output (no em dashes, no smart quotes). Formal but direct tone. Complete file contents when editing, not snippets. These apply to code comments, commit messages, and any content files you generate.
10. **Ask before any of these actions:** `git push --force`, `rm -rf`, modifying `~/.ssh/` or the server's `/etc/`, opening firewall ports, running `sudo`, making DNS changes, pushing to the server.

## 1. What we're building

A personal portfolio site that is ONLY accessible via SSH. A visitor runs `ssh -p 422 <domain>` and drops directly into a Bubble Tea TUI. No browser, no JavaScript, no hybrid web interface. The novelty is the point.

One compromise: a minimal static HTML page at `https://<domain>` that just says "this portfolio lives in the terminal, run `ssh -p 422 <domain>`" with an asciinema demo. Served by Caddy. ~50 lines of HTML. Everything else is SSH-only.

## 2. About the user (Ronit)

Context that should shape decisions:

- **Level:** Sophomore at TJHSST. Strong at TypeScript/Next.js, comfortable with Linux sysadmin, new to Go. Assume Go unfamiliarity when explaining — show don't elide.
- **Existing stack familiarity:** Next.js 16, PostgreSQL, Supabase, Caddy, Docker Compose, systemd, self-hosted Ubuntu. He has shipped a real civic tech app (Civic Cycle / fairfaxcivic.com). Lean on his existing Caddy and systemd knowledge.
- **Tooling:** Ghostty terminal, Zed editor, Claude Code is his primary agentic tool. He uses the WISC framework (CLAUDE.md, slash commands, handoff docs).
- **Dev environments:** Arch Linux (dual-boot) and macOS M2. Cross-platform-aware builds matter.
- **Server:** Ubuntu at `192.168.0.249`, username `arkx`. Cox Panoramic residential internet, port forwarding available. Caddy already running for existing projects.
- **Other active projects:** Civic Cycle (primary), Botball robotics (C++), AP Seminar PT2. Don't let termfolio block those.
- **Output preferences:** ASCII only (no em dashes), complete file contents when editing, formal but direct.

## 3. Goals and non-goals

### In scope for v1
- Production-grade SSH portfolio reachable at `ssh -p 422 <domain>` from anywhere
- At least 3 project writeups with real depth (Civic Cycle is the headline)
- Guestbook (read + write, stored in SQLite)
- "Now" page (what Ronit is working on this month)
- Connection logging with privacy-safe truncation
- systemd-managed service with auto-restart
- GitHub Actions CI: build on push, manual deploy
- Static landing page at apex domain

### Out of scope for v1 (explicitly do NOT build these yet)
- Web UI / any HTTP surface beyond the static landing page
- User authentication (all content is public)
- Content management beyond markdown files in the repo
- Per-user personalization (recognized keys, custom greetings)
- Games, easter eggs, Matrix rain, fortune cookies
- Analytics dashboards
- Multi-language support
- Themes / theme switching
- A database for anything other than guestbook + logs + counters

If the user asks for something in "out of scope" during v1 build, push back: "that's planned for v1.1, let's get v1 shipped first." If they insist, do it, but note the scope expansion in a TODO.md.

## 4. Tech stack (locked)

| Layer | Choice | Import path |
|---|---|---|
| Language | Go 1.24+ | — |
| SSH server | Wish v2 | `charm.land/wish/v2` |
| TUI framework | Bubble Tea v2 | `charm.land/bubbletea/v2` |
| Styling | Lip Gloss | `charm.land/lipgloss` |
| Markdown rendering | Glamour | `charm.land/glamour` |
| Prebuilt components | Bubbles | `charm.land/bubbles` |
| Forms (guestbook input) | Huh | `charm.land/huh` |
| SQLite (pure Go) | modernc.org/sqlite | `modernc.org/sqlite` |
| Build CI | GitHub Actions | — |
| Deployment | systemd unit + scp | — |
| Dynamic DNS | Cloudflare API + cron | — |
| Static page server | Caddy (already running on host) | — |

Confirm exact latest versions with `go list -m -versions <module>` before pinning in `go.mod`.

## 5. Repository layout

```
termfolio/
├── cmd/
│   ├── termfolio/              # main SSH server binary
│   │   └── main.go
│   └── termfolioctl/           # admin CLI (moderate guestbook, view stats)
│       └── main.go
├── internal/
│   ├── server/                 # Wish server setup, middleware wiring
│   │   ├── server.go
│   │   └── middleware.go       # connection logging, rate limit
│   ├── tui/                    # Bubble Tea program
│   │   ├── root.go             # root model, view routing
│   │   ├── keys.go             # keymap
│   │   ├── banner.go           # neofetch-style intro
│   │   ├── menu.go             # main menu state
│   │   ├── about.go
│   │   ├── projects.go         # project list + detail
│   │   ├── now.go
│   │   ├── posts.go
│   │   ├── guestbook.go        # read + compose
│   │   └── contact.go
│   ├── theme/                  # Lip Gloss styles, color palette
│   │   └── theme.go
│   ├── content/                # embedded markdown loader
│   │   └── content.go
│   ├── storage/                # SQLite access layer
│   │   ├── db.go
│   │   ├── guestbook.go
│   │   ├── connections.go
│   │   └── migrations.go
│   └── sanitize/               # ANSI/control char stripping for user input
│       └── sanitize.go
├── content/                    # markdown sources, //go:embed'd
│   ├── about.md
│   ├── now.md
│   ├── projects/
│   │   ├── civic-cycle.md
│   │   ├── termfolio.md
│   │   └── botball-2026.md
│   └── posts/
│       └── 2026-04-building-termfolio.md
├── deploy/
│   ├── termfolio.service       # systemd unit
│   ├── landing.html            # static apex-domain page
│   ├── Caddyfile.snippet       # block to add to existing Caddy config
│   └── ddns-update.sh          # Cloudflare DDNS cron script
├── .github/
│   └── workflows/
│       └── ci.yml              # test + cross-compile on push
├── scripts/
│   └── deploy.sh               # local operator deploy script
├── CLAUDE.md                   # project-specific agent context (see section 14)
├── README.md
├── TODO.md                     # backlog / scope expansions
├── Makefile
├── go.mod
├── go.sum
└── LICENSE                     # MIT
```

## 6. Functional requirements

### 6.1 Connection flow
- **F1.** On connect, render neofetch-style banner: ASCII logo (letter "R" or "ArkXero"), quick facts (school, location, current focus), connection count ("visitor #N"), prompt "press any key to enter".
- **F2.** Main menu exposes: `about`, `projects`, `now`, `posts`, `guestbook`, `contact`, `help`, `quit`.
- **F3.** Keyboard nav: `j`/`k` or arrows for up/down, `h`/`l` for panels, `enter` to select, `esc`/`q` to go back, `?` for help overlay, `ctrl+c` for hard quit.
- **F4.** Handle terminal resize. Below 80 cols, collapse sidebar. Below 40 cols, show "terminal too narrow" notice.
- **F5.** On quit: short sign-off ("thanks for stopping by — Ronit") then close cleanly. No panics to stderr.

### 6.2 Projects
- **F6.** List of projects with taglines, selectable.
- **F7.** Detail page per project rendered from markdown via Glamour. Each markdown file has frontmatter: `title`, `tagline`, `status` (shipped / in-progress / archived), `stack` (list of tech chips), `links` (github, live, etc.).
- **F8.** v1 projects: Civic Cycle (headline), termfolio (self-referential), Botball 2026. Ronit will write the content; your job is to render it correctly.

### 6.3 Guestbook
- **F9.** Scrollable list of messages, newest first. Each shows name, message, relative timestamp ("3 hours ago").
- **F10.** Compose form with name field and message field. Max 280 chars. Use `huh` for the form.
- **F11.** Rate limit: 1 post per IP prefix per 10 minutes. In-memory map is fine for v1.
- **F12.** Sanitize input: strip ANSI escape codes, control characters, and null bytes BEFORE storing. Terminal injection through guestbook is a real risk — a malicious entry could mess up every future viewer's terminal.
- **F13.** Messages have a `hidden` boolean. Default false. `termfolioctl hide <id>` flips it. Hidden messages are not rendered.

### 6.4 Now / posts
- **F14.** `now` renders a single markdown file with a prominent "last updated" timestamp derived from file mtime at build time.
- **F15.** `posts` lists markdown files from `content/posts/`, newest first by filename date prefix.

### 6.5 Ops
- **F16.** Connection logging middleware writes to SQLite: `started_at`, `ended_at`, `ip_prefix` (last octet masked), `key_fingerprint` (truncated SHA256), `term`, `width`, `height`.
- **F17.** `termfolioctl` subcommands: `connections [--since 1h]`, `guestbook list`, `guestbook hide <id>`, `guestbook show <id>`, `stats`.
- **F18.** Hidden `stats` command in the TUI shows total connections, unique visitors, most popular section.

## 7. Non-functional requirements

### Security
- Run as unprivileged user `termfolio`.
- SSH port 422, never 22.
- Accept all public keys (this is a public portfolio) but protect the host key: `chmod 600`, owned by `termfolio` user.
- Sanitize all user input for terminal escape codes before storing in SQLite.
- `fail2ban` jail for the new port (config goes in `deploy/`).
- `ufw` rules: allow 80, 443, 422 inbound; 22 only from LAN (`192.168.0.0/24`).

### Performance
- Cold start < 500ms.
- Connect-to-menu < 1s.
- Handle 50 concurrent Bubble Tea sessions.
- SQLite access serialized via single goroutine or mutex. Do not race.

### Reliability
- systemd `Restart=on-failure`, 5s backoff.
- On SIGTERM, drain active sessions for 30s before forcing exit.
- Host key persisted on disk. Never regenerate on deploy (breaks users' `known_hosts`).

### Privacy
- Mask last IP octet before writing to DB.
- Store only truncated key fingerprints (first 16 chars of SHA256).
- `privacy` TUI command displays exactly what's logged.
- Do NOT log message bodies in the connection logs — separate tables.

### Accessibility
- Respect `NO_COLOR` env var.
- Use symbols + color, not color alone, for state indicators.
- No mouse required anywhere.

## 8. Architecture notes for the agent

### Bubble Tea model shape
Use a single root `Model` with a `View` enum (`ViewMenu`, `ViewAbout`, `ViewProjects`, etc.). Sub-models per section. Root routes key events to active sub-model and handles transitions via returned commands.

DON'T build a general-purpose view router or state machine abstraction. The Elm-architecture switch on `View` is fine. Over-engineering here is a known failure mode.

### Content embedding
Use `//go:embed content/*` in `internal/content/content.go`. Parse frontmatter with `gopkg.in/yaml.v3`. Don't pull in a heavy frontend-style content framework.

### Markdown rendering
Glamour has a `DarkStyle` built-in. Start with that. Custom style later if needed.

### Storage
One SQLite file at `/var/lib/termfolio/termfolio.db`. Migrations as plain `.sql` files embedded via `//go:embed`, applied in order. Don't pull in a migration framework for three tables.

### Error handling
Anywhere an error would be surfaced to the SSH client: log the real error server-side, show a friendly message to the client. Never leak stack traces.

### Testing
Unit tests for sanitizer, rate limiter, storage layer, frontmatter parser. Integration test for the full SSH flow using Wish's `testsession` package. Skip TUI render tests for v1 — too brittle, not worth the time budget.

## 9. Build and run

### Local dev
```bash
# First time
go mod download

# Run locally (binds to 127.0.0.1:23234 in dev)
make dev

# Then from another terminal
ssh -p 23234 localhost
```

### Production build
```bash
# Cross-compile from any platform
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/termfolio ./cmd/termfolio
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/termfolioctl ./cmd/termfolioctl
```

### Makefile targets (required)
- `make dev` — run locally on port 23234
- `make build` — cross-compile for Linux amd64
- `make test` — run all tests
- `make lint` — run `go vet` and `staticcheck`
- `make deploy` — scp binaries to server, restart service (asks for confirmation)

## 10. Deployment

### Server one-time setup (ask Ronit to run these himself — do not automate)
```bash
sudo adduser --system --group termfolio
sudo mkdir -p /opt/termfolio/bin /var/lib/termfolio
sudo chown -R termfolio:termfolio /var/lib/termfolio /opt/termfolio
sudo cp deploy/termfolio.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now termfolio
sudo ufw allow 422/tcp
```

### Network
- Cloudflare DNS A record, **DNS only (gray cloud), NOT proxied**. Cloudflare's proxy is HTTP/HTTPS only; SSH will not route through it.
- Router port forward: external 422 → `192.168.0.249:422`.
- DDNS cron: script in `deploy/ddns-update.sh` runs every 5 min, updates the A record if the public IP changed. Needs a Cloudflare API token with DNS:Edit scope on the specific zone.
- Test externally from cellular network before announcing.

### Caddy
Ronit already runs Caddy on the host. Add a block to his Caddyfile for the apex domain pointing to `deploy/landing.html`. Automatic HTTPS.

## 11. Design tokens

Palette (carry over from Civic Cycle for visual through-line):
- Primary: `#1A8A9A` (teal)
- Primary dark: `#0D5E6B`
- Primary light: `#2BBDD4`
- Accent: `#F5A623` (amber)
- Background: terminal default
- Foreground: terminal default

Use Lip Gloss `AdaptiveColor` so it reads on both light and dark terminals.

Typography:
- Section titles: uppercase, bold, primary color, letter-spaced via single-space padding
- Project titles: bold, primary color
- Body: terminal default
- Code: Glamour default code block styling

Layout:
- Default (≥80 cols): left sidebar 25 cols, right pane for content
- Wide (≥120 cols): add right preview pane
- Narrow (<80 cols): top nav, single pane
- Too narrow (<40 cols): error message

## 12. Database schema

```sql
-- migrations/001_init.sql
CREATE TABLE IF NOT EXISTS connections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    started_at DATETIME NOT NULL,
    ended_at DATETIME,
    ip_prefix TEXT,
    key_fingerprint TEXT,
    term TEXT,
    width INTEGER,
    height INTEGER
);

CREATE TABLE IF NOT EXISTS guestbook (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    message TEXT NOT NULL,
    ip_prefix TEXT,
    key_fingerprint TEXT,
    hidden BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS command_usage (
    command TEXT PRIMARY KEY,
    count INTEGER DEFAULT 0,
    last_used DATETIME
);

CREATE INDEX IF NOT EXISTS idx_connections_started ON connections(started_at);
CREATE INDEX IF NOT EXISTS idx_guestbook_created ON guestbook(created_at DESC);
```

## 13. Milestones

### Week 1 — Foundations
1. Scaffold repo layout (section 5), write `go.mod`, `Makefile`, `README.md` stub, `CLAUDE.md` (section 14)
2. Get Wish v2 Bubble Tea example running locally on `127.0.0.1:23234`
3. Port example into `cmd/termfolio/main.go` + `internal/server/`
4. Define root `Model`, `View` enum, stub each section model
5. Implement keymap and menu navigation (empty sections OK)

Done when: `make dev` + `ssh -p 23234 localhost` shows a working menu.

### Week 2 — Content and rendering
6. Implement `internal/content/content.go` with embed + frontmatter parse
7. Integrate Glamour, render a test markdown file
8. Build banner component
9. Implement `about`, `projects` (list + detail), `now`, `posts`, `contact` views
10. Ronit writes `about.md`, `now.md`, `projects/civic-cycle.md`, `projects/termfolio.md`, `projects/botball-2026.md`

Done when: all read-only sections render real content.

### Week 3 — Storage and guestbook
11. Wire up `modernc.org/sqlite`, migrations, connection pool
12. Build `internal/sanitize/` with tests for ANSI stripping
13. Implement guestbook read view
14. Implement guestbook compose view with `huh` form + rate limit
15. Build connection logging middleware
16. Build `cmd/termfolioctl/` with subcommands

Done when: can SSH in, leave a guestbook entry, see it, `termfolioctl guestbook hide` works.

### Week 4 — Deploy and launch
17. Write `termfolio.service` systemd unit
18. Write `deploy/landing.html` + asciinema recording
19. Write `deploy/Caddyfile.snippet` for apex domain
20. Write `deploy/ddns-update.sh` and test the Cloudflare token flow
21. GitHub Actions workflow for CI (test + cross-compile)
22. Ronit purchases domain, configures Cloudflare DNS (gray cloud), router port forward
23. First deploy to server, smoke test from cellular
24. `fail2ban` jail config
25. Soft launch to friends, collect guestbook entries
26. Public launch

Done when: `ssh -p 422 <domain>` works from outside the LAN, landing page live, service auto-restarts.

## 14. CLAUDE.md (to be written in repo root)

When you scaffold the repo, create a `CLAUDE.md` with this content so future Claude Code sessions have context:

```markdown
# termfolio

SSH-accessible portfolio site. Single Go binary, Wish v2 + Bubble Tea v2 + Lip Gloss + Glamour, SQLite for guestbook, systemd-deployed on Ubuntu.

## Critical facts
- Wish v2 import: `charm.land/wish/v2` (NOT github.com/charmbracelet/wish)
- Bubble Tea v2 import: `charm.land/bubbletea/v2`
- SQLite: `modernc.org/sqlite` only (pure Go, no CGo)
- SSH port: 422, not 22
- Deploy target: Ubuntu at 192.168.0.249 as user `termfolio`
- Content is //go:embed'd markdown — edits require rebuild

## Commands
- `make dev` — run locally on :23234
- `make build` — cross-compile for linux/amd64
- `make test` — run tests
- `make deploy` — scp + restart remote service (asks confirmation)

## Conventions
- ASCII only in generated content (no em dashes, no smart quotes)
- Complete file contents when editing, not snippets
- Elm-architecture for Bubble Tea: switch on View enum in root model
- Serialize SQLite access via mutex or single goroutine

## See also
- `BUILD_SPEC.md` — full project spec
- `TODO.md` — scope expansions and v1.1+ ideas
```

## 15. Open decisions — ask the user before starting

Before scaffolding, confirm with Ronit:

1. **Project name.** `termfolio` is a placeholder. Options: `termfolio`, `ronit.sh`, `arkxero`, something else. This becomes the binary name and module path.
2. **Domain.** Not yet purchased. Needs to be SSH-typeable (short, no hyphens ideally). Suggestions: `ronit.sh`, `arkxero.dev`, `arkxero.com`.
3. **Module path.** Likely `github.com/ArkXero/<project-name>` once repo created.
4. **License.** MIT assumed unless Ronit says otherwise.
5. **Has he tested that his ISP allows inbound SSH on a non-standard port?** If no, do this BEFORE domain purchase. Forward any service on port 422 and try from cellular. Cox residential is known to allow this but verify.

## 16. Interaction guidelines for the agent

- **Batch tool calls when possible.** Reading multiple files to understand the codebase? Do it in parallel.
- **Commit often, with focused diffs.** One feature per commit. Conventional commits format (`feat:`, `fix:`, `chore:`, `docs:`).
- **When stuck on Wish v2 / Bubble Tea v2 API differences from v1:** check `charm.land/wish/v2/examples/` and the repo's own source before guessing. Do not invent API surface.
- **Ronit prefers complete file contents over patches.** When asking him to apply a change, show the whole file.
- **Do not narrate every thought.** Make decisions, implement, then summarize what you did. Reserve questions for genuine blockers.
- **If a task would take more than 20 minutes of continuous work, break it into sub-tasks and check in after each.**
- **When a milestone completes, pause and summarize progress against section 13.**

## 17. Definition of done — v1

- [ ] `ssh -p 422 <domain>` works from outside LAN, on IPv4 and cellular
- [ ] Banner renders correctly on 80x24 and 120x40
- [ ] All 8 main menu sections navigable via keyboard only
- [ ] At least 3 project writeups published with frontmatter
- [ ] Guestbook read + write working, rate limit enforced, sanitization enforced
- [ ] `termfolioctl` admin CLI works for listing, hiding, unhiding
- [ ] systemd service survives reboot and auto-restarts on crash
- [ ] Static landing page live at `https://<domain>`
- [ ] Asciinema recording embedded on landing page
- [ ] GitHub Actions CI green on main branch
- [ ] README has screenshot / asciinema GIF and setup instructions
- [ ] At least 5 guestbook entries from real people other than Ronit

---

## Appendix A — Wish v2 minimum example

Reference when scaffolding. This is the shape. Note the v2 import paths.

```go
package main

import (
    "context"
    "errors"
    "net"
    "os"
    "os/signal"
    "syscall"
    "time"

    tea "charm.land/bubbletea/v2"
    "charm.land/wish/v2"
    "charm.land/wish/v2/activeterm"
    "charm.land/wish/v2/bubbletea"
    "charm.land/wish/v2/logging"
    "charm.land/log"
)

const (
    host = "127.0.0.1"
    port = "23234"
)

func main() {
    s, err := wish.NewServer(
        wish.WithAddress(net.JoinHostPort(host, port)),
        wish.WithHostKeyPath(".ssh/id_ed25519"),
        wish.WithMiddleware(
            bubbletea.Middleware(teaHandler),
            activeterm.Middleware(),
            logging.Middleware(),
        ),
    )
    if err != nil {
        log.Error("could not start server", "error", err)
        os.Exit(1)
    }

    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGTERM)
    log.Info("starting SSH server", "host", host, "port", port)

    go func() {
        if err = s.ListenAndServe(); err != nil && !errors.Is(err, net.ErrClosed) {
            log.Error("server error", "error", err)
            done <- syscall.SIGTERM
        }
    }()

    <-done
    log.Info("stopping server")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := s.Shutdown(ctx); err != nil && !errors.Is(err, net.ErrClosed) {
        log.Error("shutdown error", "error", err)
    }
}

// teaHandler constructs the initial model for each new SSH session.
// Verify the exact v2 signature against the Wish examples directory before finalizing.
func teaHandler( /* ssh.Session */ ) ( /* tea.Model, []tea.ProgramOption */ ) {
    // TODO: build the root model here with access to pty dimensions
    return nil, nil
}
```

Note: the `teaHandler` signature changed between v1 and v2. Do NOT hardcode from memory — read the actual v2 example source at `charm.land/wish/v2/examples/bubbletea/` before implementing.

## Appendix B — Reference documentation URLs

- Wish repo: https://github.com/charmbracelet/wish
- Wish v2 release notes: https://github.com/charmbracelet/wish/releases/tag/v2.0.0
- Wish Bubble Tea example: https://github.com/charmbracelet/wish/blob/main/examples/bubbletea/main.go
- Bubble Tea: https://github.com/charmbracelet/bubbletea
- Lip Gloss: https://github.com/charmbracelet/lipgloss
- Glamour: https://github.com/charmbracelet/glamour
- Huh: https://github.com/charmbracelet/huh
- modernc SQLite: https://pkg.go.dev/modernc.org/sqlite
- Cloudflare DNS API: https://developers.cloudflare.com/api/operations/dns-records-for-a-zone-update-dns-record

## Appendix C — First session kickoff sequence

When Ronit starts the first Claude Code session with this doc, execute in order:

1. Confirm open decisions in section 15 (project name, domain, module path, license).
2. Create repo directory and `git init`.
3. Scaffold the layout in section 5, empty files OK.
4. Write `go.mod` with correct module path.
5. Write `Makefile` with targets from section 9.
6. Write `CLAUDE.md` from section 14.
7. Write `README.md` stub.
8. Implement Appendix A example in `cmd/termfolio/main.go`, verified against current v2 source.
9. `go mod tidy`, `make dev`, verify SSH connection works.
10. First commit: `feat: initial scaffold and working Wish v2 server`.
11. Pause and summarize. Hand off.
