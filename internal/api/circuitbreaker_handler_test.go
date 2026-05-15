package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronlens/cronlens/internal/circuitbreaker"
)

func newCBHandler() *CircuitBreakerHandler {
	return NewCircuitBreakerHandler(circuitbreaker.New())
}

func registerRule(t *testing.T, h *CircuitBreakerHandler, job string, threshold int, resetSec float64) {
	t.Helper()
	body, _ := json.Marshal(map[string]interface{}{
		"job_name":      job,
		"threshold":     threshold,
		"reset_seconds": resetSec,
	})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/circuitbreaker/rules", bytes.NewReader(body))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCBHandler_MethodNotAllowed(t *testing.T) {
	h := newCBHandler()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/circuitbreaker", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestCBHandler_RegisterInvalidJSON(t *testing.T) {
	h := newCBHandler()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/circuitbreaker/rules", bytes.NewBufferString("not-json"))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCBHandler_RegisterAndQueryState(t *testing.T) {
	b := circuitbreaker.New()
	_ = b.Register(circuitbreaker.Rule{JobName: "nightly", Threshold: 2, ResetTimeout: time.Minute})
	h := NewCircuitBreakerHandler(b)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/circuitbreaker?job=nightly", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp["state"] != "closed" {
		t.Fatalf("expected closed, got %s", resp["state"])
	}
}

func TestCBHandler_StateOpenAfterFailures(t *testing.T) {
	b := circuitbreaker.New()
	_ = b.Register(circuitbreaker.Rule{JobName: "sync", Threshold: 2, ResetTimeout: time.Minute})
	b.RecordFailure("sync")
	b.RecordFailure("sync")
	h := NewCircuitBreakerHandler(b)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/circuitbreaker?job=sync", nil)
	h.ServeHTTP(rr, req)
	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp["state"] != "open" {
		t.Fatalf("expected open, got %s", resp["state"])
	}
}

func TestCBHandler_MissingJobParam(t *testing.T) {
	h := newCBHandler()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/circuitbreaker", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
