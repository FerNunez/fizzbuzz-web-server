package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fizzbuzz-web-server/internal/fizzbuzz/handler"
	"fizzbuzz-web-server/internal/fizzbuzz/infrastructure/repository"
	"fizzbuzz-web-server/internal/fizzbuzz/service"
	"fizzbuzz-web-server/pkg/env"
)

var (
	httpAddr = env.GetString("HTTP_ADDR", ":8081")
	logLevel = env.GetString("LOG_LEVEL", "INFO")
)

func newLogger(level string) *slog.Logger {
	var slogLevel slog.Level
	err := slogLevel.UnmarshalText([]byte(level))
	if err != nil {
		slogLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel}))

	if err != nil {
		logger.Warn("wrong configured logger level, defaulting to: INFO", "cfg_logger", level)
	}
	return logger
}

func main() {
	logger := newLogger(logLevel)
	repo := repository.NewInmemoryRepository()
	svc := service.NewService(repo)
	handler := handler.NewFizzbuzzHandler(svc, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /fizzbuzz", handler.HandleFizzbuzz)
	mux.HandleFunc("GET /fizzbuzz/statistics", handler.HandleStatistics)

	server := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	// Launch server and listen for errors
	serverErrs := make(chan error, 1)
	go func() {
		logger.Info("fizzbuzz server listening", "address", httpAddr)
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
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("couln't shutdown gracefully, force closing server", "err", err)
			server.Close()
		}
		logger.Info("server closed")
	}
}
