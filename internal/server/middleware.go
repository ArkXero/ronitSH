// Package server provides Wish middleware for the SSH server.
package server

import (
	"github.com/charmbracelet/ssh"
	gossh "golang.org/x/crypto/ssh"

	"github.com/ArkXero/termfolio/internal/storage"
)

// connectionLogger is the storage interface required by ConnectionLogger.
type connectionLogger interface {
	LogConnection(ip, keyFP, term string, width, height int) (int64, error)
	EndConnection(id int64) error
}

// ConnectionLogger returns a Wish middleware that records each SSH session in
// the connections table. It inserts a row when the session begins and stamps
// ended_at when the session handler returns.
func ConnectionLogger(db connectionLogger) func(ssh.Handler) ssh.Handler {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			term, w, h := "", 0, 0
			if pty, _, ok := sess.Pty(); ok {
				term = pty.Term
				w = pty.Window.Width
				h = pty.Window.Height
			}

			ip := storage.MaskIP(sess.RemoteAddr().String())
			fp := ""
			if pk := sess.PublicKey(); pk != nil {
				fp = storage.TruncateFingerprint(gossh.FingerprintSHA256(pk))
			}

			connID, _ := db.LogConnection(ip, fp, term, w, h)
			defer func() {
				if connID > 0 {
					_ = db.EndConnection(connID)
				}
			}()

			next(sess)
		}
	}
}
