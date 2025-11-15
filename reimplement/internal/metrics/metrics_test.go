package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// TestNewMetricsRegistry tests that metrics registry is properly initialized
func TestNewMetricsRegistry(t *testing.T) {
	// Create a custom registry for this test
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	if registry == nil {
		t.Fatal("Expected non-nil MetricsRegistry")
	}

	// Verify the registry can be used without panic
	registry.RecordHTTPRequest("GET", "/test", 200, 0.1)
}

// TestRecordHTTPRequest records HTTP request metrics and verifies they're tracked
func TestRecordHTTPRequest(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	// Record multiple requests
	registry.RecordHTTPRequest("GET", "/api/schedules", 200, 0.05)
	registry.RecordHTTPRequest("GET", "/api/schedules", 200, 0.08)
	registry.RecordHTTPRequest("POST", "/api/schedules", 201, 0.15)
	registry.RecordHTTPRequest("GET", "/api/schedules", 404, 0.02)

	// Verify handler returns metrics
	handler := registry.GetHandler()
	if handler == nil {
		t.Fatal("Expected non-nil metrics handler")
	}

	// Create a test request to /metrics
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()

	// Check for expected metrics in output
	if !strings.Contains(body, "http_requests_total") {
		t.Error("Expected http_requests_total metric in output")
	}
	if !strings.Contains(body, "http_request_duration_seconds") {
		t.Error("Expected http_request_duration_seconds metric in output")
	}
}

// TestRecordHTTPError records HTTP error metrics
func TestRecordHTTPError(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	registry.RecordHTTPError("validation_error")
	registry.RecordHTTPError("validation_error")
	registry.RecordHTTPError("database_error")
	registry.RecordHTTPError("internal_error")

	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "http_errors_total") {
		t.Error("Expected http_errors_total metric in output")
	}
	if !strings.Contains(body, `error_type="validation_error"`) {
		t.Error("Expected validation_error label in output")
	}
}

// TestRecordDatabaseQuery records database query metrics including query count
func TestRecordDatabaseQuery(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	// Record queries with different query counts (to detect N+1 issues)
	registry.RecordDatabaseQuery("select", 0.05, 1)  // Good: 1 query
	registry.RecordDatabaseQuery("select", 0.08, 1)  // Good: 1 query
	registry.RecordDatabaseQuery("select", 0.12, 5)  // Bad: 5 queries (N+1 issue)
	registry.RecordDatabaseQuery("insert", 0.10, 1)  // Good: 1 query

	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "database_operations_total") {
		t.Error("Expected database_operations_total metric in output")
	}
	if !strings.Contains(body, "database_query_duration_seconds") {
		t.Error("Expected database_query_duration_seconds metric in output")
	}
	if !strings.Contains(body, "query_count_per_operation") {
		t.Error("Expected query_count_per_operation metric in output")
	}
}

// TestRecordServiceOperation records service operation metrics
func TestRecordServiceOperation(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	registry.RecordServiceOperation("ods", "import", 0.25, false)
	registry.RecordServiceOperation("amion", "scrape", 0.50, false)
	registry.RecordServiceOperation("coverage", "calculate", 0.15, false)
	registry.RecordServiceOperation("ods", "import", 0.30, true)  // With error

	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "service_operation_duration_seconds") {
		t.Error("Expected service_operation_duration_seconds metric in output")
	}
	if !strings.Contains(body, `service="ods"`) {
		t.Error("Expected service label in output")
	}
}

// TestRecordValidationError records validation error metrics
func TestRecordValidationError(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	registry.RecordValidationError("INVALID_FORMAT")
	registry.RecordValidationError("INVALID_FORMAT")
	registry.RecordValidationError("MISSING_FIELD")
	registry.RecordValidationError("DUPLICATE_ENTRY")

	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "validation_errors_total") {
		t.Error("Expected validation_errors_total metric in output")
	}
	if !strings.Contains(body, `error_code="INVALID_FORMAT"`) {
		t.Error("Expected INVALID_FORMAT label in output")
	}
}

// TestIncrementDecrementActiveJobs tests gauge increment and decrement
func TestIncrementDecrementActiveJobs(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	// Start with no active jobs
	registry.IncrementActiveJobs("amion")
	registry.IncrementActiveJobs("amion")
	registry.IncrementActiveJobs("ods")

	// Now decrement
	registry.DecrementActiveJobs("amion")

	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "active_scrape_jobs") {
		t.Error("Expected active_scrape_jobs metric in output")
	}
}

// TestSetQueueDepth tests gauge set operation for queue depth
func TestSetQueueDepth(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	registry.SetQueueDepth("import_jobs", 5)
	registry.SetQueueDepth("scrape_jobs", 10)
	registry.SetQueueDepth("import_jobs", 3)  // Update the value

	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "queue_depth") {
		t.Error("Expected queue_depth metric in output")
	}
	if !strings.Contains(body, `queue_name="import_jobs"`) {
		t.Error("Expected import_jobs queue_name label in output")
	}
}

// TestSetDatabaseConnectionPoolSize tests gauge set for connection pool
func TestSetDatabaseConnectionPoolSize(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	registry.SetDatabaseConnectionPoolSize("main", 10)
	registry.SetDatabaseConnectionPoolSize("read_replica", 5)
	registry.SetDatabaseConnectionPoolSize("main", 8)  // Update

	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "database_connection_pool_size") {
		t.Error("Expected database_connection_pool_size metric in output")
	}
}

// TestHTTPMiddleware tests that middleware properly records metrics
func TestHTTPMiddleware(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with middleware
	wrapped := registry.HTTPMiddleware(testHandler)

	// Make a request
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify metrics were recorded
	handler := registry.GetHandler()
	metricsReq := httptest.NewRequest("GET", "/metrics", nil)
	metricsW := httptest.NewRecorder()
	handler.ServeHTTP(metricsW, metricsReq)

	body := metricsW.Body.String()
	if !strings.Contains(body, "http_requests_total") {
		t.Error("Expected http_requests_total metric in middleware")
	}
}

// TestHTTPMiddlewareErrorHandling tests middleware with various status codes
func TestHTTPMiddlewareErrorHandling(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	tests := []struct {
		name       string
		statusCode int
		expected   string
	}{
		{"OK", http.StatusOK, "2xx"},
		{"Redirect", http.StatusMovedPermanently, "3xx"},
		{"NotFound", http.StatusNotFound, "4xx"},
		{"ServerError", http.StatusInternalServerError, "5xx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			wrapped := registry.HTTPMiddleware(testHandler)
			req := httptest.NewRequest("GET", "/api/test", nil)
			w := httptest.NewRecorder()
			wrapped.ServeHTTP(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}

// TestMetricsPrometheusFormat verifies output is valid Prometheus format
func TestMetricsPrometheusFormat(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	// Record some metrics
	registry.RecordHTTPRequest("GET", "/api/test", 200, 0.1)
	registry.RecordDatabaseQuery("select", 0.05, 1)

	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body := w.Body.String()

	// Prometheus metrics format checks
	if w.Header().Get("Content-Type") == "" {
		t.Error("Expected Content-Type header")
	}

	// Should contain HELP lines
	if !strings.Contains(body, "# HELP") {
		t.Error("Expected HELP comments in Prometheus format")
	}

	// Should contain TYPE lines
	if !strings.Contains(body, "# TYPE") {
		t.Error("Expected TYPE comments in Prometheus format")
	}

	// Should contain actual metric lines
	if !strings.Contains(body, "{") || !strings.Contains(body, "}") {
		t.Error("Expected metric lines with labels in Prometheus format")
	}
}

// TestConcurrentMetricRecording tests thread-safe metric recording
func TestConcurrentMetricRecording(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Launch multiple goroutines recording metrics concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				registry.RecordHTTPRequest("GET", "/api/test", 200, 0.01)
				registry.RecordDatabaseQuery("select", 0.01, 1)
				registry.IncrementActiveJobs("amion")
				registry.DecrementActiveJobs("amion")
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify metrics can be queried without error
	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Should not panic or error
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestStatusCodeLabel tests the status code label function
func TestStatusCodeLabel(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{200, "2xx"},
		{201, "2xx"},
		{299, "2xx"},
		{300, "3xx"},
		{301, "3xx"},
		{399, "3xx"},
		{400, "4xx"},
		{404, "4xx"},
		{499, "4xx"},
		{500, "5xx"},
		{502, "5xx"},
		{599, "5xx"},
		{0, "unknown"},
		{99, "unknown"},
		{600, "5xx"}, // Above 500 goes to 5xx
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.code)), func(t *testing.T) {
			result := statusCodeLabel(tt.code)
			if result != tt.expected {
				t.Errorf("statusCodeLabel(%d) = %q, want %q", tt.code, result, tt.expected)
			}
		})
	}
}

// TestResponseWriterStatusCapture tests that response writer correctly captures status
func TestResponseWriterStatusCapture(t *testing.T) {
	rw := httptest.NewRecorder()
	wrapped := &responseWriter{ResponseWriter: rw, statusCode: http.StatusOK}

	// Write header
	wrapped.WriteHeader(http.StatusNotFound)

	if wrapped.statusCode != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, wrapped.statusCode)
	}

	// Writing same header twice should not change status
	wrapped.WriteHeader(http.StatusOK)
	if wrapped.statusCode != http.StatusNotFound {
		t.Errorf("Expected status to remain %d, got %d", http.StatusNotFound, wrapped.statusCode)
	}
}

// TestMetricsWithZeroValues tests metrics with edge case values
func TestMetricsWithZeroValues(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	// Record with zero duration
	registry.RecordHTTPRequest("GET", "/test", 200, 0.0)
	registry.RecordDatabaseQuery("select", 0.0, 0)
	registry.SetQueueDepth("queue", 0)
	registry.SetDatabaseConnectionPoolSize("pool", 0)

	// Should not panic
	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// BenchmarkRecordHTTPRequest benchmarks the HTTP request recording
func BenchmarkRecordHTTPRequest(b *testing.B) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.RecordHTTPRequest("GET", "/api/test", 200, 0.05)
	}
}

// BenchmarkRecordDatabaseQuery benchmarks the database query recording
func BenchmarkRecordDatabaseQuery(b *testing.B) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.RecordDatabaseQuery("select", 0.05, 1)
	}
}

// BenchmarkHTTPMiddleware benchmarks the HTTP middleware
func BenchmarkHTTPMiddleware(b *testing.B) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := registry.HTTPMiddleware(testHandler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
	}
}

// TestMetricsIntegration is an integration test recording various metric types together
func TestMetricsIntegration(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	// Simulate a complete workflow

	// HTTP request comes in
	registry.RecordHTTPRequest("POST", "/api/schedules/import", 202, 0.01)

	// Service operation starts
	registry.IncrementActiveJobs("ods")

	// Database queries happen
	registry.RecordDatabaseQuery("insert", 0.05, 1)
	registry.RecordDatabaseQuery("select", 0.03, 1)

	// Service operation completes
	registry.RecordServiceOperation("ods", "import", 0.10, false)
	registry.DecrementActiveJobs("ods")

	// Queue depth changes
	registry.SetQueueDepth("import_jobs", 5)

	// Connection pool status
	registry.SetDatabaseConnectionPoolSize("main", 10)

	// Get all metrics
	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()

	// Verify all metric types are present
	requiredMetrics := []string{
		"http_requests_total",
		"http_request_duration_seconds",
		"database_operations_total",
		"database_query_duration_seconds",
		"service_operation_duration_seconds",
		"active_scrape_jobs",
		"queue_depth",
		"database_connection_pool_size",
	}

	for _, metric := range requiredMetrics {
		if !strings.Contains(body, metric) {
			t.Errorf("Expected metric %q in output", metric)
		}
	}
}

// TestMetricsWithLargeNumbers tests metrics with various magnitude values
func TestMetricsWithLargeNumbers(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	// Large durations
	registry.RecordHTTPRequest("GET", "/slow", 200, 120.5)

	// Large query counts (detecting severe N+1 issues)
	registry.RecordDatabaseQuery("select", 1.0, 1000)

	// Large queue depths
	registry.SetQueueDepth("queue", 10000)

	// Large connection pool
	registry.SetDatabaseConnectionPoolSize("pool", 100)

	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestMetricsOutputReadable verifies output can be fully read
func TestMetricsOutputReadable(t *testing.T) {
	customRegistry := prometheus.NewRegistry()
	registry := NewMetricsRegistryWithRegistry(customRegistry)

	registry.RecordHTTPRequest("GET", "/test", 200, 0.1)

	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	// Ensure body can be fully read
	body, err := io.ReadAll(w.Body)
	if err != nil {
		t.Errorf("Failed to read response body: %v", err)
	}

	if len(body) == 0 {
		t.Error("Response body is empty")
	}
}
