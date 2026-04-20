package storage

import (
	"fmt"
	"time"
)

// GuestbookEntry represents one row in the guestbook table.
type GuestbookEntry struct {
	ID        int
	CreatedAt time.Time
	Name      string
	Message   string
	IPPrefix  string
	KeyFP     string
	Hidden    bool
}

// AddGuestbookEntry inserts a new entry. The entry is sanitized by the caller
// before reaching this function.
func (d *DB) AddGuestbookEntry(name, message, ipPrefix, keyFP string) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	res, err := d.db.Exec(`
		INSERT INTO guestbook (created_at, name, message, ip_prefix, key_fp)
		VALUES (datetime('now'), ?, ?, ?, ?)
	`, name, message, ipPrefix, keyFP)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListGuestbookEntries returns entries newest-first.
// If includeHidden is false, hidden entries are excluded.
func (d *DB) ListGuestbookEntries(includeHidden bool) ([]GuestbookEntry, error) {
	query := `SELECT id, created_at, name, message, ip_prefix, key_fp, hidden
	          FROM guestbook`
	if !includeHidden {
		query += ` WHERE hidden = 0`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []GuestbookEntry
	for rows.Next() {
		var e GuestbookEntry
		var createdStr string
		if err := rows.Scan(
			&e.ID, &createdStr, &e.Name, &e.Message,
			&e.IPPrefix, &e.KeyFP, &e.Hidden,
		); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdStr)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// SetGuestbookEntryHidden sets the hidden flag on an entry.
func (d *DB) SetGuestbookEntryHidden(id int, hidden bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	res, err := d.db.Exec(`UPDATE guestbook SET hidden = ? WHERE id = ?`, hidden, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("entry %d not found", id)
	}
	return nil
}

// GetGuestbookCount returns the number of visible entries.
func (d *DB) GetGuestbookCount() (int, error) {
	var n int
	err := d.db.QueryRow(`SELECT COUNT(*) FROM guestbook WHERE hidden = 0`).Scan(&n)
	return n, err
}

// RelativeTime returns a human-readable relative time string (e.g. "3 hours ago").
func RelativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}
