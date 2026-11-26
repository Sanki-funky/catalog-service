package adapter

import "net/http"

func NewRouter(readiness *Readiness) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", LivenessHandler)
	mux.Handle("GET /health/ready", readiness)

	return mux
}
