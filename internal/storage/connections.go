package storage

import (
	"fmt"
	"strings"
	"time"
)

// ConnectionRecord holds the data logged for one SSH session.
type ConnectionRecord struct {
	ID        int
	StartedAt time.Time
	EndedAt   *time.Time
	IPPrefix  string
	KeyFP     string
	Term      string
	Width     int
	Height    int
}

// LogConnection inserts a new connection record and returns its ID.
// Called at the start of each SSH session.
func (d *DB) LogConnection(ip, keyFP, term string, width, height int) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	res, err := d.db.Exec(`
		INSERT INTO connections (started_at, ip_prefix, key_fp, term, width, height)
		VALUES (datetime('now'), ?, ?, ?, ?, ?)
	`, ip, keyFP, term, width, height)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// EndConnection stamps ended_at on an existing connection record.
func (d *DB) EndConnection(id int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec(
		`UPDATE connections SET ended_at = datetime('now') WHERE id = ?`, id,
	)
	return err
}

// GetVisitorCount returns the total number of connections ever logged.
func (d *DB) GetVisitorCount() (int, error) {
	var n int
	err := d.db.QueryRow(`SELECT COUNT(*) FROM connections`).Scan(&n)
	return n, err
}

// ListConnections returns recent connection records, optionally filtered by
// a minimum start time. Pass zero Time for no filter.
func (d *DB) ListConnections(since time.Time, limit int) ([]ConnectionRecord, error) {
	query := `SELECT id, started_at, ended_at, ip_prefix, key_fp, term, width, height
	          FROM connections`
	args := []any{}
	if !since.IsZero() {
		query += ` WHERE started_at >= ?`
		args = append(args, since.Format("2006-01-02 15:04:05"))
	}
	query += ` ORDER BY started_at DESC`
	if limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, limit)
	}

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ConnectionRecord
	for rows.Next() {
		var r ConnectionRecord
		var startStr string
		var endStr *string
		if err := rows.Scan(
			&r.ID, &startStr, &endStr,
			&r.IPPrefix, &r.KeyFP, &r.Term, &r.Width, &r.Height,
		); err != nil {
			return nil, err
		}
		r.StartedAt, _ = time.Parse("2006-01-02 15:04:05", startStr)
		if endStr != nil {
			t, _ := time.Parse("2006-01-02 15:04:05", *endStr)
			r.EndedAt = &t
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// MaskIP masks the last octet of an IPv4 address or the last group of IPv6.
// e.g. "192.168.0.42:1234" -> "192.168.0.xxx"
func MaskIP(addr string) string {
	// Strip port.
	host := addr
	if i := strings.LastIndex(addr, ":"); i != -1 {
		host = addr[:i]
	}
	host = strings.Trim(host, "[]") // IPv6 brackets

	parts := strings.Split(host, ".")
	if len(parts) == 4 {
		parts[3] = "xxx"
		return strings.Join(parts, ".")
	}
	// IPv6: mask last group.
	parts6 := strings.Split(host, ":")
	if len(parts6) > 1 {
		parts6[len(parts6)-1] = "xxxx"
		return strings.Join(parts6, ":")
	}
	return host
}

// TruncateFingerprint returns the first 16 characters of a pre-computed
// SHA-256 fingerprint string (e.g. "SHA256:XXXXXX..."). The caller should
// compute the full fingerprint using golang.org/x/crypto/ssh.FingerprintSHA256.
func TruncateFingerprint(fp string) string {
	if len(fp) > 16 {
		return fp[:16]
	}
	return fp
}
