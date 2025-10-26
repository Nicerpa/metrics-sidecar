package metrics

import (
	"fmt"
	"regexp"
	"strings"
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

type PathNormalizer struct {
	UUIDPattern         *regexp.Regexp
	IntegerPattern      *regexp.Regexp
	AlphanumericPattern *regexp.Regexp
	Base64Pattern       *regexp.Regexp
}

func NewPathNormalizer() *PathNormalizer {
	return &PathNormalizer{
		UUIDPattern:         regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`),
		IntegerPattern:      regexp.MustCompile(`\b\d+\b`),
		AlphanumericPattern: regexp.MustCompile(`\b[a-zA-Z0-9]{8,}\b`),
		Base64Pattern:       regexp.MustCompile(`[A-Za-z0-9+/]{16,}={0,2}`),
	}
}

func (pn *PathNormalizer) NormalizePath(path string) string {
	if path == "" || path == "/" {
		return path
	}

	var queryPart string
	if idx := strings.Index(path, "?"); idx != -1 {
		queryPart = path[idx:]
		path = path[:idx]
	}

	segments := strings.Split(path, "/")
	normalized := make([]string, len(segments))

	for i, segment := range segments {
		if segment == "" {
			normalized[i] = segment
			continue
		}

		switch {
		case pn.UUIDPattern.MatchString(segment):
			normalized[i] = ":uuid"
		case pn.Base64Pattern.MatchString(segment):
			normalized[i] = ":token"
		case pn.IntegerPattern.MatchString(segment) && len(segment) > 2: // Avoid replacing small numbers that might be version numbers
			normalized[i] = ":id"
		case pn.AlphanumericPattern.MatchString(segment) && !pn.isLikelyStaticSegment(segment):
			normalized[i] = ":key"
		default:
			normalized[i] = segment
		}
	}

	result := strings.Join(normalized, "/")
	if queryPart != "" {
		result += queryPart
	}

	return result
}

func (pn *PathNormalizer) isLikelyStaticSegment(segment string) bool {
	staticSegments := []string{
		"api", "v1", "v2", "v3", "admin", "auth", "login", "logout",
		"health", "metrics", "status", "info", "docs", "swagger",
		"users", "posts", "comments", "orders", "products", "items",
		"search", "create", "update", "delete", "list", "get", "put", "post",
		"public", "private", "static", "assets", "images", "css", "js",
	}

	segmentLower := strings.ToLower(segment)
	for _, static := range staticSegments {
		if segmentLower == static {
			return true
		}
	}

	if strings.Contains(segment, ".") {
		return true
	}

	if regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(segment) && len(segment) < 20 {
		return true
	}

	return false
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
	pathNormalizer  *PathNormalizer
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
		pathNormalizer: NewPathNormalizer(),
	}

	return collector
}

func (p *PrometheusCollector) Record(metric RequestMetric) {
	status := fmt.Sprintf("%dxx", metric.StatusCode/100)

	normalizedHandler := p.pathNormalizer.NormalizePath(metric.Handler)

	labels := prometheus.Labels{
		"method":  metric.Method,
		"handler": normalizedHandler,
		"status":  status,
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

func NewRequestCollectorWithNormalizer(normalizer *PathNormalizer) MetricsCollector {
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
		pathNormalizer: normalizer,
	}
	return collector
}

func (p *PrometheusCollector) SetPathNormalizer(normalizer *PathNormalizer) {
	p.pathNormalizer = normalizer
}

func (p *PrometheusCollector) GetPathNormalizer() *PathNormalizer {
	return p.pathNormalizer
}
