package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testAPI struct {
	handler   http.Handler
	readiness *Readiness
	logs      *bytes.Buffer
}

func newTestAPI(checks ...Check) testAPI {
	logs := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	readiness := NewReadiness(checks...)
	return testAPI{NewRouter(logger, readiness, NewMetrics()), readiness, logs}
}

func (api testAPI) do(method, path string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	api.handler.ServeHTTP(response, httptest.NewRequest(method, path, nil))
	return response
}

func TestLiveReturnsEmptyOK(t *testing.T) {
	api := newTestAPI()
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		t.Run(method, func(t *testing.T) {
			response := api.do(method, "/health/live")
			if response.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", response.Code, http.StatusOK)
			}
			if response.Body.Len() != 0 {
				t.Errorf("body = %q, want empty body", response.Body.String())
			}
		})
	}
}

func readyBody(t *testing.T, response *httptest.ResponseRecorder) readyResponse {
	t.Helper()
	var body readyResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("body %q is not JSON: %v", response.Body.String(), err)
	}
	return body
}

func TestReadyWithPassingChecks(t *testing.T) {
	api := newTestAPI(Check{Name: "database", Run: func(context.Context) error { return nil }})
	response := api.do(http.MethodGet, "/health/ready")
	if response.Code != http.StatusOK || readyBody(t, response).Status != "ok" {
		t.Errorf("got %d %s, want 200 ok", response.Code, response.Body.String())
	}
}

func TestReadyFailsWhenACheckFails(t *testing.T) {
	api := newTestAPI(
		Check{Name: "database", Run: func(context.Context) error { return errors.New("connection refused") }},
		Check{Name: "nats", Run: func(context.Context) error { return nil }},
	)
	response := api.do(http.MethodGet, "/health/ready")
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.Code)
	}
	body := readyBody(t, response)
	if body.Checks["database"] != "connection refused" {
		t.Errorf("checks = %v, want the database error", body.Checks)
	}
	if _, listed := body.Checks["nats"]; listed {
		t.Errorf("checks = %v, passing checks should not be listed", body.Checks)
	}
}

func TestReadyFailsAfterDrain(t *testing.T) {
	api := newTestAPI()
	api.readiness.Drain()
	response := api.do(http.MethodGet, "/health/ready")
	if response.Code != http.StatusServiceUnavailable || readyBody(t, response).Status != "shutting down" {
		t.Errorf("got %d %s, want 503 shutting down", response.Code, response.Body.String())
	}
}

func TestMetricsCountRequestsByRoute(t *testing.T) {
	api := newTestAPI()
	api.do(http.MethodGet, "/health/live")
	api.do(http.MethodGet, "/health/live")
	api.do(http.MethodGet, "/does-not-exist")

	response := api.do(http.MethodGet, "/metrics")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	body, _ := io.ReadAll(response.Body)
	for _, want := range []string{
		`http_requests_total{method="GET",route="/health/live",status="200"} 2`,
		`http_requests_total{method="GET",route="unmatched",status="404"} 1`,
		`http_request_duration_seconds_bucket{method="GET",route="/health/live"`,
		"go_goroutines",
	} {
		if !strings.Contains(string(body), want) {
			t.Errorf("metrics output is missing %q", want)
		}
	}
}

func TestRequestsAreLoggedAsJSON(t *testing.T) {
	api := newTestAPI()
	api.do(http.MethodGet, "/does-not-exist")

	var entry map[string]any
	if err := json.Unmarshal(api.logs.Bytes(), &entry); err != nil {
		t.Fatalf("log line %q is not JSON: %v", api.logs.String(), err)
	}
	if entry["msg"] != "request" || entry["path"] != "/does-not-exist" || entry["status"] != float64(404) {
		t.Errorf("log entry = %v", entry)
	}
}

func TestRouteLabel(t *testing.T) {
	for pattern, want := range map[string]string{
		"":                     "unmatched",
		"GET /products/{id}":   "/products/{id}",
		"/static/":             "/static/",
		"POST example.com/x/y": "example.com/x/y",
	} {
		if got := routeLabel(pattern); got != want {
			t.Errorf("routeLabel(%q) = %q, want %q", pattern, got, want)
		}
	}
}
