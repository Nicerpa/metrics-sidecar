package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ListenPort      int
	ProxyPort       int
	ProxyHost       string
	LogLevel        string
	DBPath          string
	MetricsEndpoint string
	HealthEndpoint  string
}

func LoadFromCLI() *Config {
	var (
		listenPort      = flag.Int("listen-port", 8080, "Port for the sidecard server to listen on")
		proxyPort       = flag.Int("proxy-port", 0, "Port of the target service to proxy to (required)")
		proxyHost       = flag.String("proxy-host", "localhost", "Host of the target service to proxy to")
		logLevel        = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		metricsEndpoint = flag.String("metrics-endpoint", "/metrics", "Endpoint path for metrics collection")
		healthEndpoint  = flag.String("health-endpoint", "/health", "Endpoint path for health checks")
		help            = flag.Bool("help", false, "Show help message")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nMetrics Sidecard - HTTP proxy server with metrics collection\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -listen-port 8080 -proxy-port 3000\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -listen-port 9090 -proxy-host example.com -proxy-port 80\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -listen-port 8080 -proxy-port 3000 -metrics-endpoint /api/metrics -health-endpoint /api/health\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *proxyPort == 0 {
		fmt.Fprintf(os.Stderr, "Error: -proxy-port is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	return &Config{
		ListenPort:      *listenPort,
		ProxyPort:       *proxyPort,
		ProxyHost:       *proxyHost,
		LogLevel:        *logLevel,
		DBPath:          getEnv("DB_PATH", "./metrics.db"),
		MetricsEndpoint: *metricsEndpoint,
		HealthEndpoint:  *healthEndpoint,
	}
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))

	return &Config{
		ListenPort: port,
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		DBPath:     getEnv("DB_PATH", "./data.db"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
