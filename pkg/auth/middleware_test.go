package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sidDarthVader31/apexio/pkg/auth"
)

func TestAPIKeyDisabledWhenEmpty(t *testing.T) {
	called := false
	h := auth.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}), auth.APIKey("X-API-Key", ""))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/", nil))
	if !called || rr.Code != http.StatusOK {
		t.Fatalf("called=%v code=%d", called, rr.Code)
	}
}

func TestAPIKeyRejectsMissing(t *testing.T) {
	h := auth.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), auth.APIKey("X-API-Key", "secret"))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d", rr.Code)
	}
}

func TestAPIKeyAcceptsValid(t *testing.T) {
	h := auth.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}), auth.APIKey("X-API-Key", "secret"))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-API-Key", "secret")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("code=%d", rr.Code)
	}
}
