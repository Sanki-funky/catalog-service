package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	adapter "github.com/Sanki-funky/catalog-service/internal/adapters"
)

const (
	defaultPort = "8080"

	// Kubernetes waits 30 seconds by default before killing the pod,
	// so in-flight requests get well under that to finish.
	shutdownTimeout = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	addr, err := listenAddr(os.Getenv("PORT"))
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	readiness := adapter.NewReadiness()
	server := &http.Server{
		Addr:              addr,
		Handler:           adapter.NewRouter(readiness),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Bind the port before reporting ready, so readiness never claims
	// traffic can be served while the listener doesn't exist yet.
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()

	readiness.SetReady(true)
	log.Printf("Server listening on %s", addr)

	select {
	case err := <-serveErr:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}

	// Restore default signal handling, so a second Ctrl+C or SIGTERM
	// stops the process immediately instead of waiting for the shutdown.
	stop()
	log.Println("Shutting down")
	readiness.SetReady(false)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	log.Println("Server stopped")
	return nil
}

// listenAddr turns the PORT environment variable into a listen address,
// falling back to the default port when it isn't set.
func listenAddr(port string) (string, error) {
	if port == "" {
		port = defaultPort
	}

	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return "", fmt.Errorf("invalid PORT %q: must be a number from 1 to 65535", port)
	}

	return ":" + strconv.Itoa(n), nil
}
