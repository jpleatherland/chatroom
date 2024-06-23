package main

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jpleatherland/chatroom/internal/auth"
	"github.com/jpleatherland/chatroom/internal/config"
	"github.com/jpleatherland/chatroom/internal/tui"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg, err := config.NewConfig()
	file := cfg.DatabaseFile
	sqliteDatabase, _ := sql.Open("sqlite3", file)
	defer sqliteDatabase.Close()
	if err != nil {
		log.Error("Could not read config file", "error", err)
		os.Exit(1)
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(cfg.Host, cfg.Port)),
		wish.WithHostKeyPath(cfg.HostKeyPath),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			if key.Type() != "ssh-ed25519" {
				return false
			}
			authorised, err := auth.GetUser(key, sqliteDatabase)
			if err != nil {
				if err.Error() == "Unauthorised" {
					return false
				}
				log.Error(err.Error())
				return false
			}
			if authorised {
				return true
			}
			if !authorised {
				err := auth.AddUser(ctx.User(), key, sqliteDatabase)
				if err != nil {
					log.Error(err.Error())
					return false
				}
				return true
			}
			return false
		}),
		wish.WithMiddleware(
			bubbletea.Middleware(tui.TeaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", cfg.Host, "port", cfg.Port)

	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()
	<-done

	log.Info("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() { cancel() }()
	if err = s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not shutdown server", "error", err)
	}
}
