package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetEpochSuccess(t *testing.T) {
	fixedNow := time.Unix(1_700_000_000, 0).UTC()
	h := NewEpochHandler(func() time.Time { return fixedNow })

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/epoch/2", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	var resp EpochResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.NowEpoch != fixedNow.Unix() {
		t.Fatalf("expected now_epoch %d, got %d", fixedNow.Unix(), resp.NowEpoch)
	}

	expectedFuture := fixedNow.Unix() + (2 * secondsPerDay)
	if resp.FutureEpoch != expectedFuture {
		t.Fatalf("expected future_epoch %d, got %d", expectedFuture, resp.FutureEpoch)
	}

	if resp.DaysAdded != 2 {
		t.Fatalf("expected days_added %d, got %d", 2, resp.DaysAdded)
	}
}

func TestGetEpochInvalidDays(t *testing.T) {
	h := NewEpochHandler(func() time.Time { return time.Unix(0, 0).UTC() })

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/epoch/not-an-int", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	expected := "invalid days: must be an integer"
	if resp.Error != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.Error)
	}
}

func TestGetEpochContextCanceled(t *testing.T) {
	h := NewEpochHandler(func() time.Time { return time.Unix(0, 0).UTC() })

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := httptest.NewRequest(http.MethodGet, "/epoch/1", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestTimeout {
		t.Fatalf("expected status %d, got %d", http.StatusRequestTimeout, rr.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if resp.Error != "request context canceled" {
		t.Fatalf("expected cancellation error message, got %q", resp.Error)
	}
}

func TestGetSwaggerSuccess(t *testing.T) {
	h := NewEpochHandler(func() time.Time { return time.Unix(0, 0).UTC() })

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/epoch/swagger", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	if !strings.Contains(rr.Body.String(), "\"openapi\"") {
		t.Fatalf("expected swagger payload to contain openapi field")
	}
}
