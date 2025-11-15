// Package metrics provides Prometheus metrics infrastructure for the application.
// It exports metrics via an HTTP endpoint in Prometheus format.
package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsRegistry holds all application metrics and provides helper methods
// for recording various metric types.
type MetricsRegistry struct {
	registry prometheus.Registerer

	// Counter metrics
	httpRequestsTotal       prometheus.CounterVec
	httpErrorsTotal         prometheus.CounterVec
	validationErrorsTotal   prometheus.CounterVec
	databaseOperationsTotal prometheus.CounterVec

	// Histogram metrics
	httpRequestDuration       prometheus.HistogramVec
	databaseQueryDuration     prometheus.HistogramVec
	serviceOperationDuration  prometheus.HistogramVec
	queryCountPerOperation    prometheus.HistogramVec

	// Gauge metrics
	activeScrapeJobs           prometheus.GaugeVec
	queueDepth                 prometheus.GaugeVec
	databaseConnectionPoolSize prometheus.GaugeVec

	mu sync.RWMutex
}

// NewMetricsRegistry creates and registers all application metrics using the global registry.
// It panics if any metric fails to register.
func NewMetricsRegistry() *MetricsRegistry {
	return NewMetricsRegistryWithRegistry(prometheus.DefaultRegisterer)
}

// NewMetricsRegistryWithRegistry creates and registers all application metrics with a custom registry.
// This is mainly used for testing. It panics if any metric fails to register.
func NewMetricsRegistryWithRegistry(registerer prometheus.Registerer) *MetricsRegistry {
	m := &MetricsRegistry{
		registry: registerer,
	}

	// Initialize counter metrics
	m.httpRequestsTotal = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests by method and path",
		},
		[]string{"method", "path"},
	)
	m.registry.MustRegister(&m.httpRequestsTotal)

	m.httpErrorsTotal = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total HTTP errors by error type",
		},
		[]string{"error_type"},
	)
	m.registry.MustRegister(&m.httpErrorsTotal)

	m.validationErrorsTotal = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "validation_errors_total",
			Help: "Total validation failures by error code",
		},
		[]string{"error_code"},
	)
	m.registry.MustRegister(&m.validationErrorsTotal)

	m.databaseOperationsTotal = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_operations_total",
			Help: "Total database operations by operation type",
		},
		[]string{"operation"},
	)
	m.registry.MustRegister(&m.databaseOperationsTotal)

	// Initialize histogram metrics
	m.httpRequestDuration = *prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
	m.registry.MustRegister(&m.httpRequestDuration)

	m.databaseQueryDuration = *prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
	m.registry.MustRegister(&m.databaseQueryDuration)

	m.serviceOperationDuration = *prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "service_operation_duration_seconds",
			Help:    "Service operation duration in seconds (ODS, Amion, Coverage)",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "operation"},
	)
	m.registry.MustRegister(&m.serviceOperationDuration)

	m.queryCountPerOperation = *prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "query_count_per_operation",
			Help:    "Number of database queries per operation (tracks N+1 opportunities)",
			Buckets: []float64{1, 2, 5, 10, 20, 50, 100, 500},
		},
		[]string{"operation"},
	)
	m.registry.MustRegister(&m.queryCountPerOperation)

	// Initialize gauge metrics
	m.activeScrapeJobs = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_scrape_jobs",
			Help: "Concurrent Amion scrapers",
		},
		[]string{"service"},
	)
	m.registry.MustRegister(&m.activeScrapeJobs)

	m.queueDepth = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_depth",
			Help: "Pending job queue length",
		},
		[]string{"queue_name"},
	)
	m.registry.MustRegister(&m.queueDepth)

	m.databaseConnectionPoolSize = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_connection_pool_size",
			Help: "Active database connections",
		},
		[]string{"pool_name"},
	)
	m.registry.MustRegister(&m.databaseConnectionPoolSize)

	return m
}

// RecordHTTPRequest records an HTTP request metric.
// This includes both request count and latency histogram.
// method: HTTP method (GET, POST, etc.)
// path: request path
// statusCode: HTTP response status code
// duration: request duration in seconds
func (m *MetricsRegistry) RecordHTTPRequest(method, path string, statusCode int, duration float64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.httpRequestsTotal.WithLabelValues(method, path).Inc()
	m.httpRequestDuration.WithLabelValues(method, path, statusCodeLabel(statusCode)).Observe(duration)
}

// RecordHTTPError records an HTTP error metric.
// errorType: type of error (e.g., "validation_error", "database_error", "internal_error")
func (m *MetricsRegistry) RecordHTTPError(errorType string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.httpErrorsTotal.WithLabelValues(errorType).Inc()
}

// RecordDatabaseQuery records a database query metric.
// operation: database operation (e.g., "select", "insert", "update", "delete")
// duration: query duration in seconds
// queryCount: number of individual queries executed (for N+1 detection)
func (m *MetricsRegistry) RecordDatabaseQuery(operation string, duration float64, queryCount int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.databaseOperationsTotal.WithLabelValues(operation).Inc()
	m.databaseQueryDuration.WithLabelValues(operation).Observe(duration)
	m.queryCountPerOperation.WithLabelValues(operation).Observe(float64(queryCount))
}

// RecordServiceOperation records a service operation metric.
// service: service name (e.g., "ods", "amion", "coverage")
// operation: operation name (e.g., "import", "scrape", "calculate")
// duration: operation duration in seconds
// hasError: whether the operation resulted in an error
func (m *MetricsRegistry) RecordServiceOperation(service, operation string, duration float64, hasError bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.serviceOperationDuration.WithLabelValues(service, operation).Observe(duration)
	if hasError {
		m.RecordHTTPError(service + "_error")
	}
}

// RecordValidationError records a validation error metric.
// errorCode: validation error code (e.g., "INVALID_FORMAT", "MISSING_FIELD")
func (m *MetricsRegistry) RecordValidationError(errorCode string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.validationErrorsTotal.WithLabelValues(errorCode).Inc()
}

// IncrementActiveJobs increments the active scrape job counter.
// service: service name (e.g., "amion")
func (m *MetricsRegistry) IncrementActiveJobs(service string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.activeScrapeJobs.WithLabelValues(service).Inc()
}

// DecrementActiveJobs decrements the active scrape job counter.
// service: service name (e.g., "amion")
func (m *MetricsRegistry) DecrementActiveJobs(service string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.activeScrapeJobs.WithLabelValues(service).Dec()
}

// SetQueueDepth sets the queue depth metric to a specific value.
// queueName: name of the queue (e.g., "import_jobs", "scrape_jobs")
// depth: current queue depth
func (m *MetricsRegistry) SetQueueDepth(queueName string, depth int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.queueDepth.WithLabelValues(queueName).Set(float64(depth))
}

// SetDatabaseConnectionPoolSize sets the database connection pool size to a specific value.
// poolName: name of the database pool (e.g., "main", "read_replica")
// size: current pool size
func (m *MetricsRegistry) SetDatabaseConnectionPoolSize(poolName string, size int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.databaseConnectionPoolSize.WithLabelValues(poolName).Set(float64(size))
}

// GetHandler returns an HTTP handler that serves Prometheus metrics from this registry.
func (m *MetricsRegistry) GetHandler() http.Handler {
	return promhttp.HandlerFor(m.registry.(prometheus.Gatherer), promhttp.HandlerOpts{})
}

// statusCodeLabel converts an HTTP status code to a label string.
// It groups status codes into categories for better metric grouping.
func statusCodeLabel(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}

// HTTPMiddleware returns an HTTP middleware that records request metrics.
// It wraps an http.Handler and records metrics for each request.
func (m *MetricsRegistry) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Record the start time
		startTime := prometheus.NewTimer(prometheus.ObserverFunc(func(seconds float64) {
			m.RecordHTTPRequest(r.Method, r.URL.Path, wrapped.statusCode, seconds)
		}))

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Stop the timer and record metrics
		startTime.ObserveDuration()
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader implements http.ResponseWriter.WriteHeader.
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

// Write implements http.ResponseWriter.Write.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}
