package handlers

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

const secondsPerDay int64 = 24 * 60 * 60

//go:embed swagger.json
var swaggerSpec []byte

type EpochHandler struct {
	now func() time.Time
}

type EpochResponse struct {
	NowEpoch    int64 `json:"now_epoch"`
	FutureEpoch int64 `json:"future_epoch"`
	DaysAdded   int   `json:"days_added"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewEpochHandler(now func() time.Time) *EpochHandler {
	if now == nil {
		now = time.Now
	}
	return &EpochHandler{now: now}
}

func (h *EpochHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /swagger", h.GetSwagger)
	mux.HandleFunc("GET /health", h.GetHealth)
	mux.HandleFunc("GET /epoch/{days}", h.GetEpoch)
}

func (h *EpochHandler) GetSwagger(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(swaggerSpec); err != nil {
		http.Error(w, `{"error":"failed to write response"}`, http.StatusInternalServerError)
	}
}

func (h *EpochHandler) GetHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *EpochHandler) GetEpoch(w http.ResponseWriter, r *http.Request) {
	daysRaw := r.PathValue("days")

	days, err := strconv.Atoi(daysRaw)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: "invalid days: must be an integer",
		})
		return
	}

	nowEpoch, futureEpoch, err := calculateEpochs(r.Context(), h.now().UTC().Unix(), days)
	if err != nil {
		status := http.StatusInternalServerError
		message := err.Error()

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			status = http.StatusRequestTimeout
			message = "request context canceled"
		}

		writeJSON(w, status, ErrorResponse{Error: message})
		return
	}

	writeJSON(w, http.StatusOK, EpochResponse{
		NowEpoch:    nowEpoch,
		FutureEpoch: futureEpoch,
		DaysAdded:   days,
	})
}

func calculateEpochs(ctx context.Context, nowEpoch int64, days int) (int64, int64, error) {
	select {
	case <-ctx.Done():
		return 0, 0, ctx.Err()
	default:
	}

	deltaSeconds, err := daysToSeconds(days)
	if err != nil {
		return 0, 0, err
	}

	if (deltaSeconds > 0 && nowEpoch > math.MaxInt64-deltaSeconds) ||
		(deltaSeconds < 0 && nowEpoch < math.MinInt64-deltaSeconds) {
		return 0, 0, errors.New("integer overflow when adding days")
	}

	return nowEpoch, nowEpoch + deltaSeconds, nil
}

func daysToSeconds(days int) (int64, error) {
	days64 := int64(days)
	if days64 > math.MaxInt64/secondsPerDay || days64 < math.MinInt64/secondsPerDay {
		return 0, errors.New("days value out of supported range")
	}
	return days64 * secondsPerDay, nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, `{"error":"failed to write response"}`, http.StatusInternalServerError)
	}
}

// Middleware

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, lrw.statusCode, time.Since(start))
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
