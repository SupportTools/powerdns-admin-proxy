package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/supporttools/powerdns-admin-proxy/pkg/config"
	"github.com/supporttools/powerdns-admin-proxy/pkg/logging"
)

var log = logging.SetupLogging()

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latencies in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status_code"},
	)
)

func init() {
	// Register the metrics
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

func main() {
	flag.Parse() // Parse command-line flags
	config := config.LoadConfigFromEnv()

	logging.SetupLogging()
	var err error

	if config.Debug {
		log.Debugf("Debug mode enabled")
		log.Debugf("Port: %s", config.Port)
		log.Debugf("Backend URL: %s", config.BackendURL)
	}

	// The address of the backend server.
	backendURL, err := url.Parse(config.BackendURL)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	// Start the HTTP server
	http.HandleFunc("/", proxyHandler(proxy))
	http.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))

	// Start the HTTP server for the application
	go func() {
		listenAddress := ":" + config.Port
		log.Printf("Starting server on %s", listenAddress)
		if err := http.ListenAndServe(listenAddress, nil); err != nil {
			log.Fatal(err)
		}
	}()

	// Start a new HTTP server just for Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())
	metricsAddress := ":9000"
	log.Printf("Starting metrics server on %s", metricsAddress)
	if err := http.ListenAndServe(metricsAddress, nil); err != nil {
		log.Fatal(err)
	}
}

func proxyHandler(proxy *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Use a ResponseWriter wrapper to capture the status code
		lrw := NewLoggingResponseWriter(w)

		// Serve the proxy request and capture the backend response
		logRequestResponse(proxy, lrw, r)

		duration := time.Since(start).Seconds()

		// Update metrics
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", lrw.statusCode)).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", lrw.statusCode)).Observe(duration)
	}
}

// Log and serve proxy request and capture backend response
func logRequestResponse(proxy *httputil.ReverseProxy, w *LoggingResponseWriter, r *http.Request) {
	recorder := httptest.NewRecorder()
	proxy.ServeHTTP(recorder, r)

	// Copy the recorded response to the original response writer
	for k, v := range recorder.HeaderMap {
		w.Header()[k] = v
	}
	w.WriteHeader(recorder.Code)
	_, _ = w.Write(recorder.Body.Bytes())

	// Set the status code
	w.statusCode = recorder.Code

	// Log request and response details
	log.Infof("Client IP: %s, HTTP Method: %s, URL: %s, Backend Response Status: %s",
		r.RemoteAddr,
		r.Method,
		r.URL,
		http.StatusText(recorder.Code),
	)
}

// LoggingResponseWriter is a wrapper to capture the status code
type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
