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

	"github.com/ArkXero/termfolio/internal/tui"
)

const (
	devHost    = "127.0.0.1"
	devPort    = "23234"
	prodHost   = "0.0.0.0"
	prodPort   = "422"
	devKeyPath = ".ssh/id_ed25519"
	prodKeyPath = "/var/lib/termfolio/id_ed25519"
)

func main() {
	isProd := os.Getenv("RONIT_ENV") == "production"

	host, port, keyPath := devHost, devPort, devKeyPath
	if isProd {
		host, port, keyPath = prodHost, prodPort, prodKeyPath
	}

	if !isProd {
		if err := os.MkdirAll(".ssh", 0700); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir .ssh: %v\n", err)
			os.Exit(1)
		}
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(keyPath),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
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

// teaHandler constructs a new root model for each SSH session.
func teaHandler(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, ok := sess.Pty()
	if !ok {
		return nil, nil
	}
	cfg := tui.Config{
		Term:   pty.Term,
		Width:  pty.Window.Width,
		Height: pty.Window.Height,
	}
	m := tui.NewRootModel(cfg)
	return m, nil
}
