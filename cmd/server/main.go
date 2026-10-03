package main

import (
	"context"
	"log"
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

var httpAddr = env.GetString("HTTP_ADDR", ":8081")

func main() {
	repo := repository.NewInmemoryRepository()
	svc := service.NewService(repo)
	handler := handler.NewFizzbuzzHandler(svc)

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

	serverErrs := make(chan error, 1)
	go func() {
		log.Printf("fizzbuzz server listening on %s", httpAddr)
		serverErrs <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrs:
		log.Printf("Server error: %v", err)
	case sig := <-shutdown:
		log.Printf("Received %v, shutting down...", sig)
		shutdownCtx, sshutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer sshutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			server.Close()
		}
	}
}
