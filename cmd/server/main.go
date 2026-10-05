package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fizzbuzz-web-server/internal/config"
	"fizzbuzz-web-server/internal/fizzbuzz/domain"
	"fizzbuzz-web-server/internal/fizzbuzz/handler"
	"fizzbuzz-web-server/internal/fizzbuzz/infrastructure/repository"
	"fizzbuzz-web-server/internal/fizzbuzz/service"
)

// shutdownTimeout bounds the graceful drain.
const shutdownTimeout = 10 * time.Second

// newRouter wires the HTTP routes. extracted as a function for easier testeability
func newRouter(h *handler.FizzbuzzHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("POST /fizzbuzz", h.HandleFizzbuzz)
	mux.HandleFunc("GET /fizzbuzz/statistics", h.HandleStatistics)
	return mux
}

// main run function that creates an listen a server
func run(ctx context.Context, stdout io.Writer) error {
	// config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// logger
	logger := slog.New(slog.NewTextHandler(stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	logger.Info("configuration loaded",
		"http_addr", cfg.HTTPAddr,
		"log_level", cfg.LogLevel,
		"max_limit", cfg.MaxLimit,
		"max_str_length", cfg.MaxStrLength,
	)

	// repo -> service -> handler
	repo := repository.NewInmemoryRepository()
	svc := service.NewService(repo, domain.Limits{
		MaxLimit:     cfg.MaxLimit,
		MaxStrLength: cfg.MaxStrLength,
	})
	fizzbuzzHandler := handler.NewFizzbuzzHandler(svc, logger)
	mux := newRouter(fizzbuzzHandler)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Launch server and listen for errors in channel.
	// ErrServerClosed is the expected result of Shutdown, not a failure.
	serverErrs := make(chan error, 1)
	go func() {
		logger.Info("fizzbuzz server listening", "address", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrs <- err
		}
		close(serverErrs)
	}()

	select {
	case err := <-serverErrs:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		closeErr := server.Close()
		return fmt.Errorf("graceful shutdown failed, server force closed: %w", errors.Join(err, closeErr))
	}
	logger.Info("server closed")
	return nil
}

// main entrypoint
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
