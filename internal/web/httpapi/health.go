package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// readyCheckTimeout bounds how long all readiness checks may take together.
const readyCheckTimeout = 2 * time.Second

// Live reports that the process is up and can answer HTTP. It never checks
// dependencies: a failing liveness probe makes Kubernetes restart the pod,
// which wouldn't fix a database outage.
func Live(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Check is one dependency the service needs before it can serve traffic,
// such as a database ping.
type Check struct {
	Name string
	Run  func(context.Context) error
}

// Readiness answers /health/ready. It fails while any check fails, and from
// the moment Drain is called during shutdown, so Kubernetes stops sending
// new requests before the server closes.
type Readiness struct {
	checks   []Check
	draining atomic.Bool
}

func NewReadiness(checks ...Check) *Readiness {
	return &Readiness{checks: checks}
}

// Drain marks the service as not ready for the rest of its life.
func (r *Readiness) Drain() {
	r.draining.Store(true)
}

type readyResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

func (r *Readiness) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.draining.Load() {
		writeJSON(w, http.StatusServiceUnavailable, readyResponse{Status: "shutting down"})
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), readyCheckTimeout)
	defer cancel()

	response, status := readyResponse{Status: "ok"}, http.StatusOK
	for _, check := range r.checks {
		if err := check.Run(ctx); err != nil {
			if response.Checks == nil {
				response.Checks = map[string]string{}
			}
			response.Checks[check.Name] = err.Error()
			response.Status, status = "unavailable", http.StatusServiceUnavailable
		}
	}
	writeJSON(w, status, response)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
