---
title: ronit.sh (this site)
tagline: A portfolio that lives in the terminal
status: in-progress
stack:
  - Go
  - Wish v2
  - Bubble Tea v2
  - Lip Gloss v2
  - SQLite
links:
  github: https://github.com/ArkXero/ronitSH
---

# ronit.sh

You are looking at this project right now.

An SSH-only portfolio site. Visitors run `ssh -p 422 ronit.sh` and drop into a full
Bubble Tea TUI -- no browser, no JavaScript.

## Why

Most portfolios are websites. I wanted mine to be a little different. The terminal is
where I spend most of my time, so it felt right to build something that lives there.

Also, building an SSH server from scratch in Go is a genuinely interesting systems
programming challenge.

## How it works

A Wish v2 SSH server accepts connections on port 422. For each connection, it spawns
a new Bubble Tea program and pipes the SSH session's PTY into it. The TUI is a single
root model that routes to sub-models for each section.

Content is written in Markdown and embedded directly into the binary at build time
using Go's `//go:embed`. Editing content requires a rebuild.

The guestbook entries and connection logs are stored in SQLite (pure Go, no CGo).

## Stack

- **Go 1.25** -- language
- **Wish v2** (`charm.land/wish/v2`) -- SSH server
- **Bubble Tea v2** (`charm.land/bubbletea/v2`) -- TUI framework
- **Lip Gloss v2** (`charm.land/lipgloss/v2`) -- terminal styling
- **Glamour v2** (`charm.land/glamour/v2`) -- Markdown rendering
- **Huh v2** (`charm.land/huh/v2`) -- guestbook form
- **modernc.org/sqlite** -- pure-Go SQLite
- systemd + scp for deployment
