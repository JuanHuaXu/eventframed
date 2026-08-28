package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/api"
	"github.com/JuanHuaXu/eventframed/internal/config"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "eventframed:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	settings, err := config.Parse(args)
	if err != nil {
		return err
	}
	if err := settings.EnsureDirectories(); err != nil {
		return err
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: parseLogLevel(settings.LogLevel)}))
	eventStore, err := libravdbstore.Open(libravdbstore.Config{
		Path:          settings.DatabasePath,
		Dimension:     settings.Dimension,
		Quantization:  settings.Quantization,
		MemoryMapping: true,
	})
	if err != nil {
		return err
	}
	embedder, err := embed.NewHashEmbedder(settings.Dimension)
	if err != nil {
		_ = eventStore.Close()
		return err
	}
	runtime, err := service.New(eventStore, embedder, service.Config{
		DefaultRecallK:      settings.RecallK,
		DefaultPackK:        settings.PackK,
		DefaultTokenBudget:  settings.TokenBudget,
		OverfetchMultiplier: 4,
		Quantization:        settings.Quantization,
	})
	if err != nil {
		_ = eventStore.Close()
		return err
	}
	defer runtime.Close()

	listener, cleanup, err := listen(settings.Listen)
	if err != nil {
		return err
	}
	defer cleanup()

	server := &http.Server{
		Handler:           api.NewServer(runtime, logger).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	logger.Info("eventframed started", "version", version, "listen", settings.Listen, "database", settings.DatabasePath, "dimension", settings.Dimension, "quantization", settings.Quantization, "embedder", embedder.Name())

	shutdownContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-shutdownContext.Done():
		context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(context); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
	}
	logger.Info("eventframed stopped")
	return nil
}

func listen(endpoint string) (net.Listener, func(), error) {
	if strings.HasPrefix(endpoint, "tcp://") {
		listener, err := net.Listen("tcp", strings.TrimPrefix(endpoint, "tcp://"))
		return listener, func() {}, err
	}
	path := strings.TrimPrefix(endpoint, "unix://")
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSocket == 0 {
			return nil, func() {}, fmt.Errorf("refusing to replace non-socket path %s", path)
		}
		if err := os.Remove(path); err != nil {
			return nil, func() {}, fmt.Errorf("remove stale socket: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, func() {}, err
	}
	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, func() {}, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		_ = listener.Close()
		return nil, func() {}, err
	}
	return listener, func() { _ = os.Remove(path) }, nil
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
