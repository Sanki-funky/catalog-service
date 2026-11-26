package main

import "testing"

func TestListenAddr(t *testing.T) {
	tests := []struct {
		port string
		want string
	}{
		{port: "", want: ":8080"},
		{port: "9090", want: ":9090"},
		{port: "1", want: ":1"},
		{port: "65535", want: ":65535"},
	}
	for _, tt := range tests {
		t.Run(tt.port, func(t *testing.T) {
			got, err := listenAddr(tt.port)
			if err != nil {
				t.Fatalf("listenAddr(%q) returned error: %v", tt.port, err)
			}
			if got != tt.want {
				t.Errorf("listenAddr(%q) = %q, want %q", tt.port, got, tt.want)
			}
		})
	}
}

func TestListenAddrRejectsInvalidPort(t *testing.T) {
	for _, port := range []string{"abc", "0", "-1", "65536", " 8080", "8080:"} {
		t.Run(port, func(t *testing.T) {
			if got, err := listenAddr(port); err == nil {
				t.Errorf("listenAddr(%q) = %q, want an error", port, got)
			}
		})
	}
}
