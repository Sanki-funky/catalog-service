// Command api starts catalog-service: it reads the configuration, sets up
// JSON logging and serves HTTP until it receives SIGINT or SIGTERM.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sanki-funky/catalog-service/internal/config"
	"github.com/Sanki-funky/catalog-service/internal/web/httpapi"
)

const serviceName = "catalog-service"

func main() {
	cfg, err := config.Load()
	if err != nil {
		newLogger(slog.LevelInfo).Error("invalid configuration", slog.Any("error", err))
		os.Exit(1)
	}
	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, logger); err != nil {
		logger.Error("service stopped with an error", slog.Any("error", err))
		os.Exit(1)
	}
}

func newLogger(level slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler).With(slog.String("service", serviceName))
}

// run wires the service together. Dependencies such as the database will be
// created here and their health added to the readiness checks.
func run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	readiness := httpapi.NewReadiness()
	metrics := httpapi.NewMetrics()
	server := httpapi.NewServer(httpapi.NewRouter(logger, readiness, metrics), logger)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return err
	}
	return httpapi.Serve(ctx, server, listener, readiness, cfg.ShutdownTimeout, logger)
}
