package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type RequestMetric struct {
	Method     string
	Handler    string
	StatusCode int
	Duration   float64
	Timestamp  time.Time
}

type MetricsCollector interface {
	Record(metric RequestMetric)
	Instrumentator() Instrumentator
}

type Instrumentator interface {
	InstrumentHandlerFunc(handler string, handlerFunc interface{}) interface{}
}

type PrometheusCollector struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func NewPrometheusCollector() *PrometheusCollector {
	collector := &PrometheusCollector{
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "handler", "status"},
		),
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "handler", "status"},
		),
	}

	return collector
}

func (p *PrometheusCollector) Record(metric RequestMetric) {
	labels := prometheus.Labels{
		"method":  metric.Method,
		"handler": metric.Handler,
		"status":  fmt.Sprintf("%d", metric.StatusCode),
	}

	p.requestsTotal.With(labels).Inc()
	p.requestDuration.With(labels).Observe(metric.Duration)
}

func (p *PrometheusCollector) Instrumentator() Instrumentator {
	return &PrometheusInstrumentator{}
}

type PrometheusInstrumentator struct{}

func (p *PrometheusInstrumentator) InstrumentHandlerFunc(handler string, handlerFunc interface{}) interface{} {
	return handlerFunc
}

func NewRequestCollector() MetricsCollector {
	return NewPrometheusCollector()
}
