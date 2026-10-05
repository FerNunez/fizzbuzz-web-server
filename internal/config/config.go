// Package config loads the service configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

// Config holds every setting the server needs to start.
type Config struct {
	// HTTPAddr is the address the server listens on, for example ":8081".
	HTTPAddr string
	// LogLevel is the minimum level written by the logger.
	LogLevel slog.Level
	// MaxLimit is the maximum accepted fizzbuzz limit.
	MaxLimit int
	// MaxStrLength is the maximum accepted length of str1 and str2.
	MaxStrLength int
}

// Load reads the configuration from the environment, using defaults for unset variables.
func Load() (Config, error) {
	var l loader
	cfg := Config{
		HTTPAddr:     l.string("HTTP_ADDR", ":8081"),
		LogLevel:     l.logLevel("LOG_LEVEL", slog.LevelInfo),
		MaxLimit:     l.int("MAX_LIMIT", 10000),
		MaxStrLength: l.int("MAX_STR_LENGTH", 100),
	}

	// TODO: Should I move this domain relates check inside the domain? add a Validate() func into domain.Limits
	if cfg.MaxLimit <= 0 {
		l.errorf("MAX_LIMIT must be greater than 0, got %d", cfg.MaxLimit)
	}
	if cfg.MaxStrLength <= 0 {
		l.errorf("MAX_STR_LENGTH must be greater than 0, got %d", cfg.MaxStrLength)
	}

	return cfg, errors.Join(l.errs...)
}

// loader reads environment variables and collects parse errors so all of them
// are reported together.
type loader struct {
	errs []error
}

func (l *loader) errorf(format string, args ...any) {
	l.errs = append(l.errs, fmt.Errorf(format, args...))
}

func (l *loader) string(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func (l *loader) int(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		l.errorf("%s=%q: not an integer", key, val)
		return fallback
	}
	return n
}

func (l *loader) logLevel(key string, fallback slog.Level) slog.Level {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	var level slog.Level
	if err := level.UnmarshalText([]byte(val)); err != nil {
		l.errorf("%s=%q: must be DEBUG, INFO, WARN or ERROR", key, val)
		return fallback
	}
	return level
}
