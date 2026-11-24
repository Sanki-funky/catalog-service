package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"
)

// An in-flight request finishes after shutdown starts, readiness flips to
// failing, and Serve returns nil.
func TestServeShutsDownGracefully(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	readiness := NewReadiness()
	started, release := make(chan struct{}), make(chan struct{})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		_, _ = io.WriteString(w, "done")
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	served := make(chan error, 1)
	go func() {
		served <- Serve(ctx, NewServer(mux, logger), listener, readiness, 5*time.Second, logger)
	}()

	type result struct {
		body string
		err  error
	}
	responses := make(chan result, 1)
	go func() {
		response, err := http.Get("http://" + listener.Addr().String() + "/slow")
		if err != nil {
			responses <- result{err: err}
			return
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		responses <- result{string(body), err}
	}()

	<-started
	cancel() // what SIGTERM does in main
	waitFor(t, func() bool { return readiness.draining.Load() })
	close(release)

	if got := <-responses; got.err != nil || got.body != "done" {
		t.Errorf("in-flight request = %q, %v; want it to finish", got.body, got.err)
	}
	select {
	case err := <-served:
		if err != nil {
			t.Errorf("Serve returned %v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return after shutdown")
	}
}

func TestServeReturnsListenerErrors(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	listener.Close() // Serve fails straight away on a closed listener
	err = Serve(context.Background(), NewServer(http.NewServeMux(), logger), listener,
		NewReadiness(), time.Second, logger)
	if err == nil {
		t.Error("expected an error from a closed listener")
	}
}

func waitFor(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatal("condition not met in time")
		}
		time.Sleep(5 * time.Millisecond)
	}
}
