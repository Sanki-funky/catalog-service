package adapter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLivenessHandler(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		t.Run(method, func(t *testing.T) {
			request := httptest.NewRequest(method, "/livez", nil)
			response := httptest.NewRecorder()

			LivenessHandler(response, request)

			if response.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", response.Code, http.StatusOK)
			}
			if response.Body.Len() != 0 {
				t.Errorf("body = %q, want empty body", response.Body.String())
			}
		})
	}
}
