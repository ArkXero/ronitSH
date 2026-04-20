// Package storage manages the SQLite database for guestbook, connection logs,
// and command usage counters.
//
// All writes are serialized via SetMaxOpenConns(1) on the sql.DB plus an
// explicit sync.Mutex on DB. Reads are unrestricted.
package storage

import (
	"database/sql"
	"fmt"
	"sync"

	_ "modernc.org/sqlite" // registers the "sqlite" driver via init()
)

// DB wraps a *sql.DB with a mutex to serialize writes.
// The underlying SQLite file is opened with WAL mode for better
// read-concurrency across sessions.
type DB struct {
	db *sql.DB
	mu sync.Mutex
}

// Open opens (or creates) the SQLite database at path and runs any pending
// migrations. Call Close when done.
func Open(path string) (*DB, error) {
	// WAL mode + 5s busy timeout prevent "database is locked" under load.
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(wal)&_pragma=busy_timeout(5000)", path)
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %q: %w", path, err)
	}

	// Serialise all writes through one connection.
	sqlDB.SetMaxOpenConns(1)

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	store := &DB{db: sqlDB}

	if err := migrate(sqlDB); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return store, nil
}

// Close closes the underlying database.
func (d *DB) Close() error {
	return d.db.Close()
}

// IncrementCommandUsage records that a TUI command was used.
func (d *DB) IncrementCommandUsage(command string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec(`
		INSERT INTO command_usage (command, count, last_used)
		VALUES (?, 1, datetime('now'))
		ON CONFLICT(command) DO UPDATE SET
			count    = count + 1,
			last_used = datetime('now')
	`, command)
	return err
}

// GetStats returns aggregate statistics.
func (d *DB) GetStats() (Stats, error) {
	var s Stats
	err := d.db.QueryRow(`SELECT COUNT(*) FROM connections`).Scan(&s.TotalConnections)
	if err != nil {
		return s, err
	}
	err = d.db.QueryRow(`SELECT COUNT(DISTINCT ip_prefix) FROM connections`).Scan(&s.UniqueVisitors)
	if err != nil {
		return s, err
	}
	var cmd sql.NullString
	err = d.db.QueryRow(
		`SELECT command FROM command_usage ORDER BY count DESC LIMIT 1`,
	).Scan(&cmd)
	if err != nil && err != sql.ErrNoRows {
		return s, err
	}
	if cmd.Valid {
		s.MostPopularSection = cmd.String
	}
	return s, nil
}

// Stats holds aggregate stats returned by GetStats.
type Stats struct {
	TotalConnections   int
	UniqueVisitors     int
	MostPopularSection string
}
