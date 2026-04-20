-- 001_init.sql
-- Initial schema for ronit.sh

CREATE TABLE IF NOT EXISTS connections (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    started_at  DATETIME NOT NULL,
    ended_at    DATETIME,
    ip_prefix   TEXT,
    key_fp      TEXT,
    term        TEXT,
    width       INTEGER,
    height      INTEGER
);

CREATE TABLE IF NOT EXISTS guestbook (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at  DATETIME NOT NULL,
    name        TEXT NOT NULL,
    message     TEXT NOT NULL,
    ip_prefix   TEXT,
    key_fp      TEXT,
    hidden      BOOLEAN NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS command_usage (
    command     TEXT PRIMARY KEY,
    count       INTEGER NOT NULL DEFAULT 0,
    last_used   DATETIME
);

CREATE INDEX IF NOT EXISTS idx_connections_started ON connections(started_at);
CREATE INDEX IF NOT EXISTS idx_guestbook_created   ON guestbook(created_at DESC);
