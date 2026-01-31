package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ChernykhITMO/order-processing-platform/gateway/internal/metrics"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func Instrument(service, route string, h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		h(sw, r)

		status := strconv.Itoa(sw.status)

		metrics.HTTPRequestTotal.WithLabelValues(
			service,
			r.Method,
			route,
			status,
		).Inc()

		metrics.HTTPRequestDurationSeconds.WithLabelValues(
			service,
			r.Method,
			route,
		).Observe(time.Since(start).Seconds())

	})
}
