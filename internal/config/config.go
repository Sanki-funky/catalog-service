// Package config reads the service configuration from environment variables.
//
// Every setting has a default that works for local development, so the service
// starts with no environment at all. A value that is set but invalid stops the
// service at startup instead of being silently replaced by the default.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

// DefaultPort is this service's port when PORT is unset. Each service has its
// own, so they can all run side by side on one machine.
const DefaultPort = 8080

// Config holds every setting the service reads at startup.
type Config struct {
	// Port is the HTTP port for the API, health checks and metrics. PORT, default DefaultPort.
	Port int
	// LogLevel is the minimum level written to the log. LOG_LEVEL, default info.
	LogLevel slog.Level
	// ShutdownTimeout is how long in-flight requests get to finish after
	// SIGTERM. SHUTDOWN_TIMEOUT, default 10s.
	ShutdownTimeout time.Duration
}

// Load reads the configuration from the process environment.
func Load() (Config, error) {
	return FromLookup(os.LookupEnv)
}

// FromLookup reads the configuration through lookup, which behaves like
// os.LookupEnv. Tests pass a map-backed function instead of touching the
// real environment. All invalid values are reported together.
func FromLookup(lookup func(string) (string, bool)) (Config, error) {
	r := reader{lookup: lookup}
	cfg := Config{
		Port:            read(&r, "PORT", DefaultPort, parsePort),
		LogLevel:        read(&r, "LOG_LEVEL", slog.LevelInfo, parseLevel),
		ShutdownTimeout: read(&r, "SHUTDOWN_TIMEOUT", 10*time.Second, parsePositiveDuration),
	}
	return cfg, errors.Join(r.errs...)
}

type reader struct {
	lookup func(string) (string, bool)
	errs   []error
}

// read returns the parsed value of key, or fallback when key is unset or empty.
func read[T any](r *reader, key string, fallback T, parse func(string) (T, error)) T {
	raw, ok := r.lookup(key)
	if !ok || raw == "" {
		return fallback
	}
	value, err := parse(raw)
	if err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s=%q: %w", key, raw, err))
		return fallback
	}
	return value
}

func parsePort(raw string) (int, error) {
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return 0, errors.New("must be a number from 1 to 65535")
	}
	return port, nil
}

func parseLevel(raw string) (slog.Level, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(raw)); err != nil {
		return level, errors.New("must be debug, info, warn or error")
	}
	return level, nil
}

func parsePositiveDuration(raw string) (time.Duration, error) {
	duration, err := time.ParseDuration(raw)
	if err != nil || duration <= 0 {
		return 0, errors.New("must be a positive duration such as 10s or 1m")
	}
	return duration, nil
}
