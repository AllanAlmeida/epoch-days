package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"epoch-days/internal/handlers"
)

func TestIntegrationEpochEndpoint(t *testing.T) {
	fixedNow := time.Unix(1_800_000_000, 0).UTC()
	h := handlers.NewEpochHandler(func() time.Time { return fixedNow })

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/epoch/1")
	if err != nil {
		t.Fatalf("failed to call integration endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var body handlers.EpochResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.NowEpoch != fixedNow.Unix() {
		t.Fatalf("expected now_epoch %d, got %d", fixedNow.Unix(), body.NowEpoch)
	}

	expectedFuture := fixedNow.Unix() + 24*60*60
	if body.FutureEpoch != expectedFuture {
		t.Fatalf("expected future_epoch %d, got %d", expectedFuture, body.FutureEpoch)
	}
}
