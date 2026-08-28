package api

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

var latencyBounds = [...]time.Duration{time.Millisecond, 5 * time.Millisecond, 10 * time.Millisecond, 25 * time.Millisecond, 50 * time.Millisecond, 100 * time.Millisecond, 250 * time.Millisecond, 500 * time.Millisecond, time.Second}

type runtimeMetrics struct {
	requests   atomic.Uint64
	errors     atomic.Uint64
	totalNanos atomic.Uint64
	buckets    [len(latencyBounds) + 1]atomic.Uint64
}

func newRuntimeMetrics() *runtimeMetrics { return &runtimeMetrics{} }

func (m *runtimeMetrics) observe(duration time.Duration, status int) {
	m.requests.Add(1)
	m.totalNanos.Add(uint64(max(duration, 0)))
	if status >= 400 {
		m.errors.Add(1)
	}
	index := len(latencyBounds)
	for i, bound := range latencyBounds {
		if duration <= bound {
			index = i
			break
		}
	}
	m.buckets[index].Add(1)
}

func (m *runtimeMetrics) handle(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = fmt.Fprintf(writer, "eventframed_http_requests_total %d\neventframed_http_errors_total %d\neventframed_http_duration_seconds_sum %.9f\n", m.requests.Load(), m.errors.Load(), float64(m.totalNanos.Load())/float64(time.Second))
	var cumulative uint64
	for i, bound := range latencyBounds {
		cumulative += m.buckets[i].Load()
		_, _ = fmt.Fprintf(writer, "eventframed_http_duration_seconds_bucket{le=\"%.3f\"} %d\n", bound.Seconds(), cumulative)
	}
	cumulative += m.buckets[len(latencyBounds)].Load()
	_, _ = fmt.Fprintf(writer, "eventframed_http_duration_seconds_bucket{le=\"+Inf\"} %d\n", cumulative)
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
