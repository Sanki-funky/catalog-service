// Package httpapi is the HTTP edge of the service: routing, health checks,
// metrics, request logging and the server lifecycle.
package httpapi

import (
	"log/slog"
	"net/http"
)

// NewRouter registers every route and wraps them in logging and metrics.
func NewRouter(logger *slog.Logger, readiness *Readiness, metrics *Metrics) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", Live)
	mux.Handle("GET /health/ready", readiness)
	mux.Handle("GET /metrics", metrics.Handler())

	return Instrument(mux, logger, metrics)
}
