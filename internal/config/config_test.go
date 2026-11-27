package config

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

func env(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

func TestDefaultsWhenNothingIsSet(t *testing.T) {
	cfg, err := FromLookup(env(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Config{Port: DefaultPort, LogLevel: slog.LevelInfo, ShutdownTimeout: 10 * time.Second}
	if cfg != want {
		t.Errorf("config = %+v, want %+v", cfg, want)
	}
}

func TestReadsValuesFromEnvironment(t *testing.T) {
	cfg, err := FromLookup(env(map[string]string{
		"PORT":             "9090",
		"LOG_LEVEL":        "DEBUG",
		"SHUTDOWN_TIMEOUT": "30s",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Config{Port: 9090, LogLevel: slog.LevelDebug, ShutdownTimeout: 30 * time.Second}
	if cfg != want {
		t.Errorf("config = %+v, want %+v", cfg, want)
	}
}

func TestEmptyValueMeansDefault(t *testing.T) {
	cfg, err := FromLookup(env(map[string]string{"PORT": ""}))
	if err != nil || cfg.Port != DefaultPort {
		t.Errorf("port = %d, err = %v; want %d and no error", cfg.Port, err, DefaultPort)
	}
}

func TestReportsEveryInvalidValue(t *testing.T) {
	_, err := FromLookup(env(map[string]string{
		"PORT":             "80a",
		"LOG_LEVEL":        "loud",
		"SHUTDOWN_TIMEOUT": "-5s",
	}))
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, key := range []string{"PORT", "LOG_LEVEL", "SHUTDOWN_TIMEOUT"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error %q does not mention %s", err, key)
		}
	}
}

func TestRejectsOutOfRangePort(t *testing.T) {
	for _, port := range []string{"0", "65536", "-1"} {
		if _, err := FromLookup(env(map[string]string{"PORT": port})); err == nil {
			t.Errorf("PORT=%s: expected an error", port)
		}
	}
}
