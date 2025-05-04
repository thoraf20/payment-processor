package observability

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	paymentRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_requests_total",
			Help: "Total number of payment requests",
		},
		[]string{"status", "method"},
	)
	
	paymentProcessingTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_processing_time_seconds",
			Help:    "Time taken to process payments",
			Buckets: []float64{0.1, 0.5, 1, 2, 5},
		},
		[]string{"method"},
	)
)

func init() {
	prometheus.MustRegister(paymentRequests)
	prometheus.MustRegister(paymentProcessingTime)
}