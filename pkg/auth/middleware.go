// Package auth provides optional HTTP middleware hooks for Apexio services.
package auth

import (
	"net/http"
	"strings"
)

// Middleware wraps an HTTP handler (stdlib-compatible).
type Middleware func(http.Handler) http.Handler

// Chain applies middleware outer-to-inner; returns the original handler if chain is empty.
func Chain(h http.Handler, chain ...Middleware) http.Handler {
	for i := len(chain) - 1; i >= 0; i-- {
		if chain[i] != nil {
			h = chain[i](h)
		}
	}
	return h
}

// APIKey enforces X-API-Key when expected is non-empty. Disabled when expected is "".
func APIKey(header, expected string) Middleware {
	if expected == "" {
		return nil
	}
	if header == "" {
		header = "X-API-Key"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := strings.TrimSpace(r.Header.Get(header))
			if got == "" || got != expected {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
