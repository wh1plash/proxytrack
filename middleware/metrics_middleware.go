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

		urlLabel := c.OriginalURL()
		if target, ok := c.Locals("target_url").(string); ok {
			urlLabel = target
		}

		if err != nil {
			// fmt.Println("Find error", err)
			if apiErr, ok := err.(api.Error); ok {
				p.TotalErrors.WithLabelValues(handlerName, strconv.Itoa(apiErr.Code), urlLabel).Inc()
			} else {
				if valErr, ok := err.(api.ValidationError); ok {
					p.TotalErrors.WithLabelValues(handlerName, urlLabel, "Validation").Inc()
					_ = valErr.Status
				}
			}
		}
		p.TotalRequests.WithLabelValues(handlerName, urlLabel).Inc()
		latency := time.Since(start).Seconds()
		p.RequestLatency.WithLabelValues(handlerName, urlLabel).Observe(latency)

		status := c.Response().StatusCode()
		p.CountStatuses.WithLabelValues(strconv.Itoa(status), urlLabel).Inc()

		code, _ := c.Locals("resp_code").(string)
		p.CountRespCodes.WithLabelValues(urlLabel, code).Inc()

		return err
	}
}

type PromMetrics struct {
	TotalRequests  *prometheus.CounterVec   `json:"total_requests"`
	RequestLatency *prometheus.HistogramVec `json:"request_latency"`
	TotalErrors    *prometheus.CounterVec   `json:"total_requests_errors"`
	CountStatuses  *prometheus.CounterVec   `json:"count_statuses"`
	CountRespCodes *prometheus.CounterVec   `json:"count_resp_codes"`
}

func NewPromMetrics() *PromMetrics {
	reqCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_requests",
			Help: "Total number of requests",
		},
		[]string{"handler", "urlLabel"})

	reqLatency := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_latency",
			Help:    "Request latency in seconds",
			Buckets: []float64{0.1, 0.5, 5.0, 10, 30},
		},
		[]string{"handler", "urlLabel"})

	reqErrCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_requests_errors",
			Help: "Total number of errors",
		},
		[]string{"handler", "urlLabel", "type"})

	CountStatuses := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "count_statuses",
			Help: "Nnumber responses by status",
		},
		[]string{"status", "urlLabel"})

	CountRespCodes := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "count_responses",
			Help: "Nnumber responses by code",
		},
		[]string{"urlLabel", "code"})

	return &PromMetrics{
		TotalRequests:  reqCounter,
		RequestLatency: reqLatency,
		TotalErrors:    reqErrCounter,
		CountStatuses:  CountStatuses,
		CountRespCodes: CountRespCodes,
	}
}
