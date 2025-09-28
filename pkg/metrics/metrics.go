package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

type Metric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

type RequestMetric struct {
	Method     string
	Path       string
	StatusCode int
	Duration   time.Duration
	Timestamp  time.Time
}

type RequestStats struct {
	TotalRequests   int64
	RequestRate     float64
	AvgResponseTime time.Duration
	StatusCodes     map[int]int64
}

type RequestCollector struct {
	mu            sync.RWMutex
	totalRequests int64
	statusCodes   map[int]int64
	totalDuration int64
	firstRequest  time.Time
	lastRequest   time.Time
}

func NewRequestCollector() *RequestCollector {
	return &RequestCollector{
		statusCodes: make(map[int]int64),
	}
}

func (c *RequestCollector) Record(metric RequestMetric) {
	c.mu.Lock()
	defer c.mu.Unlock()

	atomic.AddInt64(&c.totalRequests, 1)

	c.statusCodes[metric.StatusCode]++

	atomic.AddInt64(&c.totalDuration, int64(metric.Duration))

	if c.firstRequest.IsZero() {
		c.firstRequest = metric.Timestamp
	}
	c.lastRequest = metric.Timestamp
}

func (c *RequestCollector) GetMetrics() RequestStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := atomic.LoadInt64(&c.totalRequests)
	totalDur := atomic.LoadInt64(&c.totalDuration)

	var avgResponseTime time.Duration
	var requestRate float64

	if total > 0 {
		avgResponseTime = time.Duration(totalDur / total)

		if !c.firstRequest.IsZero() && !c.lastRequest.IsZero() {
			duration := c.lastRequest.Sub(c.firstRequest)
			if duration > 0 {
				requestRate = float64(total) / duration.Seconds()
			}
		}
	}

	statusCodes := make(map[int]int64, len(c.statusCodes))
	for code, count := range c.statusCodes {
		statusCodes[code] = count
	}

	return RequestStats{
		TotalRequests:   total,
		RequestRate:     requestRate,
		AvgResponseTime: avgResponseTime,
		StatusCodes:     statusCodes,
	}
}

type MetricCollector interface {
	Collect() ([]Metric, error)
	Name() string
}

type Registry struct {
	collectors []MetricCollector
}

func NewRegistry() *Registry {
	return &Registry{
		collectors: make([]MetricCollector, 0),
	}
}

func (r *Registry) Register(collector MetricCollector) {
	r.collectors = append(r.collectors, collector)
}

func (r *Registry) CollectAll() ([]Metric, error) {
	var allMetrics []Metric

	for _, collector := range r.collectors {
		metrics, err := collector.Collect()
		if err != nil {
			return nil, err
		}
		allMetrics = append(allMetrics, metrics...)
	}

	return allMetrics, nil
}
