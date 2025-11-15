# Metrics Infrastructure Documentation

## Overview

The metrics infrastructure provides comprehensive Prometheus-compatible monitoring for the schedCU application. All metrics are exported via an HTTP endpoint in Prometheus exposition format.

## Metric Types

### Counter Metrics

Counters measure the total number of events that have occurred. They only increase and reset when the application restarts.

#### `http_requests_total`
- **Description**: Total HTTP requests by method and path
- **Labels**: `method` (HTTP method), `path` (request path)
- **Example Query**: `rate(http_requests_total[5m])` - requests per second over 5 minutes

#### `http_errors_total`
- **Description**: Total HTTP errors by error type
- **Labels**: `error_type` (e.g., "validation_error", "database_error", "internal_error")
- **Example Query**: `increase(http_errors_total[1h])` - total errors in last hour

#### `validation_errors_total`
- **Description**: Total validation failures by error code
- **Labels**: `error_code` (e.g., "INVALID_FORMAT", "MISSING_FIELD", "DUPLICATE_ENTRY")
- **Example Query**: `topk(5, sum by (error_code) (rate(validation_errors_total[5m])))` - top 5 validation errors

#### `database_operations_total`
- **Description**: Total database operations by operation type
- **Labels**: `operation` (e.g., "select", "insert", "update", "delete")
- **Example Query**: `sum by (operation) (rate(database_operations_total[5m]))` - operations per second by type

### Histogram Metrics

Histograms measure the distribution of values in buckets. They provide sum, count, and bucket information.

#### `http_request_duration_seconds`
- **Description**: HTTP request latency in seconds
- **Labels**: `method` (HTTP method), `path` (request path), `status` (status class: 2xx, 3xx, 4xx, 5xx)
- **Buckets**: Prometheus default buckets (0.005s, 0.01s, 0.025s, 0.05s, 0.075s, 0.1s, 0.25s, 0.5s, 0.75s, 1s, 2.5s, 5s, 7.5s, 10s, +Inf)
- **Example Queries**:
  - `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))` - 95th percentile latency
  - `sum(rate(http_request_duration_seconds_sum[5m])) / sum(rate(http_request_duration_seconds_count[5m]))` - average latency

#### `database_query_duration_seconds`
- **Description**: Database query duration in seconds
- **Labels**: `operation` (database operation type)
- **Buckets**: Prometheus default buckets
- **Example Queries**:
  - `histogram_quantile(0.99, rate(database_query_duration_seconds_bucket[5m]))` - 99th percentile query latency
  - `rate(database_query_duration_seconds_sum[5m]) / rate(database_query_duration_seconds_count[5m])` - average query duration

#### `service_operation_duration_seconds`
- **Description**: Service operation duration in seconds (ODS, Amion, Coverage)
- **Labels**: `service` (e.g., "ods", "amion", "coverage"), `operation` (e.g., "import", "scrape", "calculate")
- **Buckets**: Prometheus default buckets
- **Example Queries**:
  - `histogram_quantile(0.95, rate(service_operation_duration_seconds_bucket{service="amion"}[5m]))` - 95th percentile Amion scrape time
  - `sum by (service) (rate(service_operation_duration_seconds_sum[5m])) / sum by (service) (rate(service_operation_duration_seconds_count[5m]))` - average operation time by service

#### `query_count_per_operation`
- **Description**: Number of database queries per operation (tracks N+1 opportunities)
- **Labels**: `operation` (operation name)
- **Buckets**: [1, 2, 5, 10, 20, 50, 100, 500] - designed to detect N+1 query issues
- **Example Queries**:
  - `histogram_quantile(0.95, rate(query_count_per_operation_bucket[5m]))` - 95th percentile query count
  - `max by (operation) (histogram_quantile(0.99, rate(query_count_per_operation_bucket[5m])))` - worst-case query counts

### Gauge Metrics

Gauges measure values that can go up and down. They represent the current state.

#### `active_scrape_jobs`
- **Description**: Number of concurrent Amion scrapers currently running
- **Labels**: `service` (e.g., "amion")
- **Usage**: Monitoring concurrent scraper load and capacity
- **Example Query**: `max(active_scrape_jobs)` - peak concurrent scrapers

#### `queue_depth`
- **Description**: Pending job queue length
- **Labels**: `queue_name` (e.g., "import_jobs", "scrape_jobs")
- **Usage**: Queue depth monitoring and alerting
- **Example Queries**:
  - `queue_depth{queue_name="scrape_jobs"}` - current scrape job queue depth
  - `max(queue_depth{queue_name="import_jobs"})` - peak import job queue depth

#### `database_connection_pool_size`
- **Description**: Number of active database connections
- **Labels**: `pool_name` (e.g., "main", "read_replica")
- **Usage**: Database connection pool monitoring
- **Example Query**: `database_connection_pool_size{pool_name="main"}` - current main pool connections

## Usage Examples

### Recording HTTP Requests

```go
package main

import (
    "net/http"
    "time"
    "github.com/schedcu/reimplement/internal/metrics"
)

func main() {
    // Create metrics registry
    metricsRegistry := metrics.NewMetricsRegistry()

    // Record an HTTP request
    metricsRegistry.RecordHTTPRequest("GET", "/api/schedules", 200, 0.125)
    metricsRegistry.RecordHTTPRequest("POST", "/api/schedules", 201, 0.250)

    // Record an HTTP error
    metricsRegistry.RecordHTTPError("validation_error")

    // Use middleware to automatically record metrics
    mux := http.NewServeMux()
    wrappedMux := metricsRegistry.HTTPMiddleware(mux)
    http.ListenAndServe(":8080", wrappedMux)

    // Expose metrics endpoint
    http.Handle("/metrics", metricsRegistry.GetHandler())
}
```

### Recording Database Operations

```go
func QueryDatabase(operation string) {
    startTime := time.Now()
    queryCount := 0

    // Simulate queries
    queryCount += 1  // Query 1: SELECT * FROM schedules
    queryCount += 3  // Query 2-4: SELECT shifts for each schedule (N+1 issue)

    duration := time.Since(startTime).Seconds()
    metricsRegistry.RecordDatabaseQuery(operation, duration, queryCount)
}
```

### Recording Service Operations

```go
func ImportODSFile(filePath string) {
    startTime := time.Now()
    hasError := false

    // Perform import
    if err := performImport(filePath); err != nil {
        hasError = true
    }

    duration := time.Since(startTime).Seconds()
    metricsRegistry.RecordServiceOperation("ods", "import", duration, hasError)
}
```

### Monitoring Active Jobs

```go
func ScrapeAmionSchedules() {
    metricsRegistry.IncrementActiveJobs("amion")
    defer metricsRegistry.DecrementActiveJobs("amion")

    // Perform scraping
    scrapeAmion()
}
```

### Tracking Queue Depth

```go
func UpdateQueueStatus(jobQueue chan Job) {
    depth := len(jobQueue)
    metricsRegistry.SetQueueDepth("import_jobs", depth)
}
```

## Prometheus Configuration

### Scrape Configuration

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'schedcu'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
    scrape_timeout: 10s
```

## Alert Examples

### High Error Rate

```yaml
alert: HighHTTPErrorRate
expr: |
  (
    sum(rate(http_errors_total[5m]))
    /
    sum(rate(http_requests_total[5m]))
  ) > 0.05
for: 5m
annotations:
  summary: "High HTTP error rate ({{ $value | humanizePercentage }})"
  description: "Error rate is above 5% threshold for 5 minutes"
```

### N+1 Query Detection

```yaml
alert: N1QueryPatternDetected
expr: |
  histogram_quantile(0.95, rate(query_count_per_operation_bucket[5m])) > 10
for: 10m
annotations:
  summary: "Possible N+1 query pattern detected"
  description: "95th percentile query count per operation exceeds 10"
```

### Excessive Queue Depth

```yaml
alert: HighQueueDepth
expr: queue_depth{queue_name="scrape_jobs"} > 100
for: 5m
annotations:
  summary: "High scrape job queue depth"
  description: "Queue depth: {{ $value }} jobs"
```

### Slow Service Operations

```yaml
alert: SlowServiceOperation
expr: |
  histogram_quantile(0.95, rate(service_operation_duration_seconds_bucket{service="amion"}[5m])) > 300
for: 10m
annotations:
  summary: "Slow Amion service operations"
  description: "95th percentile duration: {{ $value }}s (threshold: 300s)"
```

## Grafana Dashboard Configuration Example

### JSON Model for Dashboard

```json
{
  "dashboard": {
    "title": "schedCU Metrics",
    "panels": [
      {
        "title": "HTTP Requests per Second",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total[5m])) by (method)"
          }
        ],
        "type": "graph"
      },
      {
        "title": "HTTP Error Rate",
        "targets": [
          {
            "expr": "sum(rate(http_errors_total[5m]))"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Request Latency (95th percentile)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Database Operations per Second",
        "targets": [
          {
            "expr": "sum(rate(database_operations_total[5m])) by (operation)"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Service Operation Duration",
        "targets": [
          {
            "expr": "sum(rate(service_operation_duration_seconds_sum[5m])) by (service) / sum(rate(service_operation_duration_seconds_count[5m])) by (service)"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Active Scrape Jobs",
        "targets": [
          {
            "expr": "active_scrape_jobs{service=\"amion\"}"
          }
        ],
        "type": "gauge"
      },
      {
        "title": "Queue Depths",
        "targets": [
          {
            "expr": "queue_depth"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Database Connection Pool",
        "targets": [
          {
            "expr": "database_connection_pool_size"
          }
        ],
        "type": "gauge"
      },
      {
        "title": "Query Count per Operation (95th %ile)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(query_count_per_operation_bucket[5m])) by (operation)"
          }
        ],
        "type": "graph"
      }
    ]
  }
}
```

## Performance

Metrics recording is highly optimized:

- **HTTP Request**: ~155 ns per operation
- **Database Query**: ~180 ns per operation
- **HTTP Middleware**: ~3.8 µs per request (negligible impact)

All operations are thread-safe and can be called concurrently from multiple goroutines.

## Testing

The metrics package includes comprehensive tests covering:

- Metric creation and registration
- Counter increments
- Histogram observations
- Gauge updates
- HTTP middleware integration
- Concurrent metric recording
- Prometheus format validation
- Edge cases (zero values, large numbers)

Run tests with:

```bash
go test ./internal/metrics/... -v
```

Run benchmarks with:

```bash
go test ./internal/metrics/... -bench=.
```

## Best Practices

1. **Always include context labels**: Use meaningful labels for filtering metrics
2. **Monitor N+1 queries**: Use `query_count_per_operation` histogram to detect inefficient queries
3. **Set queue depth alerts**: Alert on high queue depths to detect processing bottlenecks
4. **Track service operation errors**: Use `RecordServiceOperation` with hasError=true for error tracking
5. **Expose metrics on a separate port**: In production, expose `/metrics` on a different port than your API
6. **Use metric recording helpers**: Use the provided helper functions rather than recording individual metrics
7. **Be mindful of cardinality**: Avoid creating labels with unbounded values (e.g., user IDs, request UUIDs)

## Troubleshooting

### Metrics Not Appearing in Prometheus

1. Verify the `/metrics` endpoint is accessible: `curl http://localhost:8080/metrics`
2. Check Prometheus scrape configuration in `prometheus.yml`
3. Ensure metrics have been recorded (check application logs)
4. Verify the endpoint is responding with valid Prometheus format

### High Cardinality Issues

If you see high cardinality warnings in Prometheus:

1. Audit your label usage - avoid labels with unbounded values
2. Use `statusCodeLabel()` function to group status codes into classes (2xx, 3xx, etc.)
3. Consider using Prometheus relabeling rules to drop high-cardinality labels

### Performance Impact

The metrics recording has minimal performance impact (~155 ns per HTTP request, negligible in typical web request contexts). If you notice performance issues:

1. Ensure metrics recording is not happening in critical paths
2. Use middleware instead of manual recording for HTTP metrics
3. Consider reducing scrape frequency in Prometheus configuration

## Future Enhancements

Potential improvements for future versions:

1. Custom bucket configurations per metric type
2. Metrics summary export (OpenMetrics format)
3. Metrics persistence across restarts
4. Custom metric types (e.g., Summary metrics)
5. Distributed tracing integration
