package adapter

import (
	"net/http"
	"sync/atomic"
)

// Readiness reports whether the service should receive traffic.
// It starts out not ready. The server marks it ready once it is listening
// and not ready again when it starts shutting down.
type Readiness struct {
	ready atomic.Bool
}

func NewReadiness() *Readiness {
	return &Readiness{}
}

func (r *Readiness) SetReady(ready bool) {
	r.ready.Store(ready)
}

func (r *Readiness) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	if !r.ready.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}
