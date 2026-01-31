package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HTTPRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "opp",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total HTTP requests",
		}, []string{"service", "method", "route", "status"})

	HTTPRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "opp",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"service", "method", "route"})
)

func Register() {
	prometheus.MustRegister(
		HTTPRequestTotal,
		HTTPRequestDurationSeconds)
}
