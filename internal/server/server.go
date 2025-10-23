package server

import (
	"fmt"
	"log"
	"metrics-sidecard/internal/config"
	"metrics-sidecard/pkg/metrics"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	config    *config.Config
	proxy     *httputil.ReverseProxy
	collector metrics.MetricsCollector
	targetURL *url.URL
}

func New(cfg *config.Config) *Server {
	targetURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", cfg.ProxyHost, cfg.ProxyPort),
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	collector := metrics.NewRequestCollector()

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
	}

	server := &Server{
		config:    cfg,
		proxy:     proxy,
		collector: collector,
		targetURL: targetURL,
	}

	proxy.Transport = &MetricsTransport{
		collector: collector,
		transport: http.DefaultTransport,
	}

	return server
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc(s.config.HealthEndpoint, s.handleHealth)

	mux.HandleFunc(s.config.MetricsEndpoint, s.handleMetrics)

	mux.HandleFunc("/", s.handleProxy)

	log.Printf("Metrics Sidecard listening on :%d", s.config.ListenPort)
	log.Printf("Proxying requests to %s", s.targetURL.String())
	log.Printf("Health endpoint available at %s", s.config.HealthEndpoint)
	log.Printf("Metrics endpoint available at %s", s.config.MetricsEndpoint)

	return http.ListenAndServe(fmt.Sprintf(":%d", s.config.ListenPort), mux)
}

func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	recorder := &responseRecorder{
		ResponseWriter: w,
		statusCode:     200,
	}

	s.proxy.ServeHTTP(recorder, r)

	duration := time.Since(start).Seconds()
	s.collector.Record(metrics.RequestMetric{
		Method:     r.Method,
		Handler:    r.URL.Path,
		StatusCode: recorder.statusCode,
		Duration:   duration,
		Timestamp:  start,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "healthy", "proxy_target": "%s"}`, s.targetURL.String())
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = s.config.MetricsFormat
	}

	// TODO: Add other formats here
	switch strings.ToLower(format) {
	case "prometheus":
		promhttp.Handler().ServeHTTP(w, r)
	default:
		promhttp.Handler().ServeHTTP(w, r)
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

type MetricsTransport struct {
	collector metrics.MetricsCollector
	transport http.RoundTripper
}

func (t *MetricsTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// For now, just pass through to the default transport
	// Additional metrics can be collected here if needed
	return t.transport.RoundTrip(req)
}
