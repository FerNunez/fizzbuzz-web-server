package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fizzbuzz-web-server/internal/fizzbuzz/domain"
	"fizzbuzz-web-server/internal/fizzbuzz/handler"
	"fizzbuzz-web-server/internal/fizzbuzz/infrastructure/repository"
	"fizzbuzz-web-server/internal/fizzbuzz/service"
)

// newTestServer starts the real router with the real service and repository.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	repo := repository.NewInmemoryRepository()
	svc := service.NewService(repo, domain.Limits{MaxLimit: 100, MaxStrLength: 10})
	h := handler.NewFizzbuzzHandler(svc, slog.New(slog.DiscardHandler))

	server := httptest.NewServer(newRouter(h))
	t.Cleanup(server.Close)
	return server
}

func do(t *testing.T, method, url, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func TestRoutes(t *testing.T) {
	server := newTestServer(t)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{name: "health", method: http.MethodGet, path: "/health", wantStatus: http.StatusOK},
		{name: "statistics", method: http.MethodGet, path: "/fizzbuzz/statistics", wantStatus: http.StatusOK},
		{name: "wrong method", method: http.MethodGet, path: "/fizzbuzz", wantStatus: http.StatusMethodNotAllowed},
		{name: "unknown path", method: http.MethodGet, path: "/unknown", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := do(t, tt.method, server.URL+tt.path, "")
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("expects status %d, got %d", tt.wantStatus, resp.StatusCode)
			}
		})
	}
}

func TestStatisticsAfterRequests(t *testing.T) {
	server := newTestServer(t)
	bodyA := `{"int1":3,"int2":5,"limit":16,"str1":"fizz","str2":"buzz"}`
	bodyB := `{"int1":2,"int2":7,"limit":10,"str1":"foo","str2":"bar"}`

	for _, body := range []string{bodyA, bodyB, bodyA, bodyA} {
		resp := do(t, http.MethodPost, server.URL+"/fizzbuzz", body)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expects status %d on POST /fizzbuzz, got %d", http.StatusOK, resp.StatusCode)
		}
	}

	resp := do(t, http.MethodGet, server.URL+"/fizzbuzz/statistics", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expects status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	type params struct {
		Int1  int    `json:"int1"`
		Int2  int    `json:"int2"`
		Limit int    `json:"limit"`
		Str1  string `json:"str1"`
		Str2  string `json:"str2"`
	}
	var got struct {
		Data struct {
			Params params `json:"params"`
			Hits   int    `json:"hits"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("expects valid JSON body, got error: %v", err)
	}
	want := params{Int1: 3, Int2: 5, Limit: 16, Str1: "fizz", Str2: "buzz"}
	if got.Data.Params != want || got.Data.Hits != 3 {
		t.Fatalf("expects %+v with 3 hits, got %+v", want, got.Data)
	}
}

func TestRun(t *testing.T) {
	t.Run("invalid configuration returns an error", func(t *testing.T) {
		t.Setenv("MAX_LIMIT", "abc")
		if err := run(context.Background(), io.Discard); err == nil {
			t.Fatal("expects an error, got nil")
		}
	})

	t.Run("cancelled context shuts down cleanly", func(t *testing.T) {
		t.Setenv("HTTP_ADDR", "127.0.0.1:0")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := run(ctx, io.Discard); err != nil {
			t.Fatalf("expects nil error, got %v", err)
		}
	})
}
