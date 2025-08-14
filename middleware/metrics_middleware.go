package middleware

import (
	"proxytrack/api"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func (p *PromMetrics) WithMetrics(h fiber.Handler, handlerName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := h(c)
		if err != nil {
			if apiErr, ok := err.(api.Error); ok {
				p.TotalErrors.WithLabelValues(handlerName).Inc()
				_ = apiErr.Code
			}
		}
		p.TotalRequests.WithLabelValues(handlerName).Inc()
		time := time.Since(c.Context().Time())
		p.RequestLatency.WithLabelValues(handlerName).Observe(float64(time.Milliseconds()))
		return err
	}
}

type PromMetrics struct {
	TotalRequests  *prometheus.CounterVec   `json:"total_requests"`
	RequestLatency *prometheus.HistogramVec `json:"request_latency"`
	TotalErrors    *prometheus.CounterVec   `json:"total_requests_errors"`
}

func NewPromMetrics() *PromMetrics {
	reqCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_requests",
			Help: "Total number of requests",
		},
		[]string{"handler"})

	reqLatency := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_latency",
			Help:    "Request latency in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0},
		},
		[]string{"handler"})

	reqErrCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_requests_errors",
			Help: "Total number of errors",
		},
		[]string{"handler"})

	return &PromMetrics{
		TotalRequests:  reqCounter,
		RequestLatency: reqLatency,
		TotalErrors:    reqErrCounter,
	}
}
