package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// NewServer returns an HTTP server with timeouts, so slow clients can't hold
// connections open forever.
func NewServer(handler http.Handler, logger *slog.Logger) *http.Server {
	return &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
}

// Serve runs server on listener until ctx is cancelled (normally by SIGTERM),
// then shuts down gracefully: readiness starts failing, the listener stops
// accepting connections, and in-flight requests get up to shutdownTimeout to
// finish. It returns nil after a clean shutdown.
func Serve(
	ctx context.Context,
	server *http.Server,
	listener net.Listener,
	readiness *Readiness,
	shutdownTimeout time.Duration,
	logger *slog.Logger,
) error {
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	logger.Info("server started", slog.String("address", listener.Addr().String()))

	select {
	case err := <-serveErr:
		return err // the server failed on its own; nothing to shut down
	case <-ctx.Done():
	}

	logger.Info("shutting down", slog.String("timeout", shutdownTimeout.String()))
	readiness.Drain()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return errors.Join(errors.New("graceful shutdown did not finish"), err, server.Close())
	}
	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("server stopped")
	return nil
}
