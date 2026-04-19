# ronit.sh

A portfolio site that lives entirely in the terminal.

```
ssh -p 422 ronit.sh
```

No browser, no JavaScript. Just SSH into a full interactive TUI.

## What you will find

- About me and my background
- Project writeups (Civic Cycle, Botball 2026, this project)
- What I am working on right now
- A guestbook -- leave a message
- Contact info

## Tech

Go 1.25, Wish v2 (SSH server), Bubble Tea v2 (TUI), Lip Gloss v2 (styling),
Glamour v2 (Markdown), SQLite (guestbook + logs), systemd, Ubuntu.

Single binary. No Docker, no Node, no Python.

## Local dev

```bash
# Requires Go 1.24+
make dev
# then in another terminal:
ssh -p 23234 localhost
```

## Build

```bash
make build   # cross-compiles for linux/amd64 into dist/
```
