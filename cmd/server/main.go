package main

import (
	"context"
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

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	logger.Info("configuration loaded",
		"http_addr", cfg.HTTPAddr,
		"log_level", cfg.LogLevel,
		"max_limit", cfg.MaxLimit,
		"max_str_length", cfg.MaxStrLength,
	)

	repo := repository.NewInmemoryRepository()
	svc := service.NewService(repo, domain.Limits{
		MaxLimit:     cfg.MaxLimit,
		MaxStrLength: cfg.MaxStrLength,
	})
	handler := handler.NewFizzbuzzHandler(svc, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /fizzbuzz", handler.HandleFizzbuzz)
	mux.HandleFunc("GET /fizzbuzz/statistics", handler.HandleStatistics)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Launch server and listen for errors
	serverErrs := make(chan error, 1)
	go func() {
		logger.Info("fizzbuzz server listening", "address", cfg.HTTPAddr)
		serverErrs <- server.ListenAndServe()
	}()

	// Listen to shutdoiwn signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-serverErrs:
		logger.Error("Server error, exiting..", "err", err)
		os.Exit(1)
	case sig := <-shutdown:
		logger.Info("Received signal to shutting down", "signal", sig)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("couln't shutdown gracefully, force closing server", "err", err)
			if err := server.Close(); err != nil {
				logger.Error("force close failed", "err", err)
			}
		}
		logger.Info("server closed")
	}
}
