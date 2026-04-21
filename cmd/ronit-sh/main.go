package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"
	"github.com/charmbracelet/ssh"
	gossh "golang.org/x/crypto/ssh"

	termfolio "github.com/ArkXero/termfolio"
	"github.com/ArkXero/termfolio/internal/content"
	"github.com/ArkXero/termfolio/internal/ratelimit"
	"github.com/ArkXero/termfolio/internal/server"
	"github.com/ArkXero/termfolio/internal/storage"
	"github.com/ArkXero/termfolio/internal/tui"
)

const (
	devHost     = "127.0.0.1"
	devPort     = "23234"
	prodHost    = "127.0.0.1"
	prodPort    = "2222"
	devKeyPath  = ".ssh/id_ed25519"
	prodKeyPath = "/var/lib/termfolio/id_ed25519"
	devDBPath   = "./termfolio-dev.db"
	prodDBPath  = "/var/lib/termfolio/termfolio.db"
)

func main() {
	isProd := os.Getenv("RONIT_ENV") == "production"

	host, port, keyPath := devHost, devPort, devKeyPath
	dbPath := devDBPath
	if isProd {
		host, port, keyPath = prodHost, prodPort, prodKeyPath
		dbPath = prodDBPath
	}

	if !isProd {
		if err := os.MkdirAll(".ssh", 0700); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir .ssh: %v\n", err)
			os.Exit(1)
		}
	}

	// Open SQLite database.
	db, err := storage.Open(dbPath)
	if err != nil {
		log.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("database ready", "path", dbPath)

	// Shared rate limiter: 1 guestbook entry per 10 minutes per IP prefix.
	rl := ratelimit.New(10 * time.Minute)

	// Load all embedded content once at startup.
	loader, err := content.NewLoader(termfolio.ContentFS)
	if err != nil {
		log.Error("failed to load content", "error", err)
		os.Exit(1)
	}
	log.Info("content loaded",
		"projects", len(loader.Projects),
		"posts", len(loader.Posts),
	)

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(keyPath),
		wish.WithMiddleware(
			bubbletea.Middleware(makeTeaHandler(loader, db, rl)),
			activeterm.Middleware(),
			server.ConnectionLogger(db),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("could not create server", "error", err)
		os.Exit(1)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	log.Info("starting SSH server", "host", host, "port", port)

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, net.ErrClosed) {
			log.Error("server error", "error", err)
			done <- syscall.SIGTERM
		}
	}()

	<-done
	log.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, net.ErrClosed) {
		log.Error("shutdown error", "error", err)
	}
}

// makeTeaHandler returns a Wish bubbletea handler that closes over the shared
// content loader, database, and rate limiter. A new root model is created per
// SSH session.
func makeTeaHandler(loader *content.Loader, db *storage.DB, rl *ratelimit.Limiter) bubbletea.Handler {
	return func(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
		pty, _, ok := sess.Pty()
		if !ok {
			return nil, nil
		}

		ip := storage.MaskIP(sess.RemoteAddr().String())
		fp := ""
		if sess.PublicKey() != nil {
			fp = storage.TruncateFingerprint(gossh.FingerprintSHA256(sess.PublicKey()))
		}

		cfg := tui.Config{
			Term:   pty.Term,
			Width:  pty.Window.Width,
			Height: pty.Window.Height,
			IP:     ip,
			KeyFP:  fp,
			RL:     rl,
			DB:     db,
		}
		m := tui.NewRootModel(cfg, loader)
		return m, nil
	}
}
