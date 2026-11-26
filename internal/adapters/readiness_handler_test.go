package adapter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadinessRoute(t *testing.T) {
	tests := []struct {
		name  string
		ready bool
		want  int
	}{
		{name: "not ready", ready: false, want: http.StatusServiceUnavailable},
		{name: "ready", ready: true, want: http.StatusOK},
	}
	for _, tt := range tests {
		for _, method := range []string{http.MethodGet, http.MethodHead} {
			t.Run(tt.name+"/"+method, func(t *testing.T) {
				readiness := NewReadiness()
				readiness.SetReady(tt.ready)
				router := NewRouter(readiness)

				request := httptest.NewRequest(method, "/health/ready", nil)
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)

				if response.Code != tt.want {
					t.Errorf("status = %d, want %d", response.Code, tt.want)
				}
				if response.Body.Len() != 0 {
					t.Errorf("body = %q, want empty body", response.Body.String())
				}
			})
		}
	}
}

func TestReadinessCanBeWithdrawn(t *testing.T) {
	readiness := NewReadiness()
	router := NewRouter(readiness)

	readiness.SetReady(true)
	readiness.SetReady(false)

	request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}
