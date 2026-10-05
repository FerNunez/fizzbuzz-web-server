package config

import (
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	for _, key := range []string{"HTTP_ADDR", "LOG_LEVEL", "MAX_LIMIT", "MAX_STR_LENGTH"} {
		t.Setenv(key, "") // restores the original value after the test
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expects no error, got: %v", err)
	}
	want := Config{
		HTTPAddr:     ":8081",
		LogLevel:     slog.LevelInfo,
		MaxLimit:     10000,
		MaxStrLength: 100,
	}
	if cfg != want {
		t.Fatalf("expects %+v, got %+v", want, cfg)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9000")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("MAX_LIMIT", "50")
	t.Setenv("MAX_STR_LENGTH", "8")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expects no error, got: %v", err)
	}
	want := Config{
		HTTPAddr:     ":9000",
		LogLevel:     slog.LevelDebug,
		MaxLimit:     50,
		MaxStrLength: 8,
	}
	if cfg != want {
		t.Fatalf("expects %+v, got %+v", want, cfg)
	}
}

func TestLoadInvalid(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr []string
	}{
		{
			name:    "not an integer",
			env:     map[string]string{"MAX_LIMIT": "abc"},
			wantErr: []string{`MAX_LIMIT="abc"`},
		},
		{
			name:    "out of range",
			env:     map[string]string{"MAX_STR_LENGTH": "0"},
			wantErr: []string{"MAX_STR_LENGTH must be greater than 0"},
		},
		{
			name:    "all errors reported together",
			env:     map[string]string{"LOG_LEVEL": "LOUD", "MAX_LIMIT": "-5"},
			wantErr: []string{`LOG_LEVEL="LOUD"`, "MAX_LIMIT must be greater than 0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			_, err := Load()
			if err == nil {
				t.Fatal("expects error, got none")
			}
			for _, want := range tt.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("expects error to contain %q, got: %v", want, err)
				}
			}
		})
	}
}
