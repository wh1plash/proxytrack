package middleware

import (
	"proxytrack/api"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func (p *PromMetrics) WithMetrics(h fiber.Handler, handlerName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := h(c)
		if err != nil {
			if apiErr, ok := err.(api.Error); ok {
				p.TotalErrors.WithLabelValues(handlerName, "Timeout").Inc()
				_ = apiErr.Code
			} else {
				if valErr, ok := err.(api.ValidationError); ok {
					p.TotalErrors.WithLabelValues(handlerName, "Validation").Inc()
					_ = valErr.Status
				}
			}
		}
		p.TotalRequests.WithLabelValues(handlerName).Inc()
		latency := time.Since(start).Seconds()
		p.RequestLatency.WithLabelValues(handlerName).Observe(latency)

		status := c.Response().StatusCode()
		p.CountStatuses.WithLabelValues(strconv.Itoa(status)).Inc()
		return err
	}
}

type PromMetrics struct {
	TotalRequests  *prometheus.CounterVec   `json:"total_requests"`
	RequestLatency *prometheus.HistogramVec `json:"request_latency"`
	TotalErrors    *prometheus.CounterVec   `json:"total_requests_errors"`
	CountStatuses  *prometheus.CounterVec   `json:"count_statuses"`
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
			Buckets: []float64{0.1, 0.5, 5.0, 10, 30},
		},
		[]string{"handler"})

	reqErrCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_requests_errors",
			Help: "Total number of errors",
		},
		[]string{"handler", "type"})

	CountStatuses := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "count_statuses",
			Help: "Nnumber responses by status",
		},
		[]string{"status"})

	return &PromMetrics{
		TotalRequests:  reqCounter,
		RequestLatency: reqLatency,
		TotalErrors:    reqErrCounter,
		CountStatuses:  CountStatuses,
	}
}
