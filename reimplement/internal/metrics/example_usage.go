package metrics

// This file contains example usage patterns for the metrics infrastructure.
// These are documentation examples and are not executed as tests.

/*
Example 1: Basic HTTP Server with Metrics

	func main() {
		// Initialize metrics registry
		metricsRegistry := NewMetricsRegistry()

		// Create a basic HTTP handler
		mux := http.NewServeMux()
		mux.HandleFunc("/api/schedules", func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			defer func() {
				duration := time.Since(start).Seconds()
				metricsRegistry.RecordHTTPRequest(r.Method, r.URL.Path, http.StatusOK, duration)
			}()

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		// Wrap with metrics middleware
		wrappedMux := metricsRegistry.HTTPMiddleware(mux)

		// Expose metrics endpoint
		http.Handle("/metrics", metricsRegistry.GetHandler())

		// Start server
		http.ListenAndServe(":8080", wrappedMux)
	}

Example 2: Recording Database Operations

	func FetchScheduleWithShifts(scheduleID string) error {
		startTime := time.Now()
		queryCount := 0

		// Query 1: Get schedule
		schedule, err := db.Query("SELECT * FROM schedules WHERE id = ?", scheduleID)
		queryCount++
		if err != nil {
			metricsRegistry.RecordDatabaseQuery("select", time.Since(startTime).Seconds(), queryCount)
			metricsRegistry.RecordHTTPError("database_error")
			return err
		}

		// Query 2-N: Get shifts for schedule (N+1 opportunity)
		for _, shift := range schedule.ShiftIDs {
			_, err := db.Query("SELECT * FROM shifts WHERE id = ?", shift)
			queryCount++
			if err != nil {
				metricsRegistry.RecordDatabaseQuery("select", time.Since(startTime).Seconds(), queryCount)
				return err
			}
		}

		metricsRegistry.RecordDatabaseQuery("select", time.Since(startTime).Seconds(), queryCount)
		return nil
	}

Example 3: Service Operation Tracking

	func ImportODSSchedule(filePath string) error {
		startTime := time.Now()
		metricsRegistry.IncrementActiveJobs("ods")
		defer metricsRegistry.DecrementActiveJobs("ods")

		hasError := false
		defer func() {
			duration := time.Since(startTime).Seconds()
			metricsRegistry.RecordServiceOperation("ods", "import", duration, hasError)
		}()

		// Parse ODS file
		schedules, err := parseODSFile(filePath)
		if err != nil {
			metricsRegistry.RecordValidationError("PARSE_ERROR")
			hasError = true
			return err
		}

		// Insert schedules
		for _, schedule := range schedules {
			if err := validateSchedule(schedule); err != nil {
				metricsRegistry.RecordValidationError("INVALID_SCHEDULE")
				metricsRegistry.RecordHTTPError("validation_error")
				hasError = true
				return err
			}

			if err := insertSchedule(schedule); err != nil {
				metricsRegistry.RecordHTTPError("database_error")
				hasError = true
				return err
			}
		}

		return nil
	}

Example 4: Concurrent Job Processing

	func ProcessScrapeQueue(jobQueue chan Job, numWorkers int) {
		for i := 0; i < numWorkers; i++ {
			go func(workerID int) {
				for job := range jobQueue {
					metricsRegistry.IncrementActiveJobs("amion")
					startTime := time.Now()
					hasError := false

					// Process job
					if err := performScrape(job); err != nil {
						hasError = true
						metricsRegistry.RecordValidationError(err.Error())
					}

					duration := time.Since(startTime).Seconds()
					metricsRegistry.RecordServiceOperation("amion", "scrape", duration, hasError)
					metricsRegistry.DecrementActiveJobs("amion")

					// Update queue depth
					metricsRegistry.SetQueueDepth("scrape_jobs", len(jobQueue))
				}
			}(i)
		}
	}

Example 5: Connection Pool Monitoring

	type DatabasePool struct {
		metricsRegistry *MetricsRegistry
		maxConns        int
		activeConns     int
	}

	func (p *DatabasePool) AcquireConnection() (*Connection, error) {
		if p.activeConns >= p.maxConns {
			metricsRegistry.RecordHTTPError("pool_exhausted")
			return nil, errors.New("connection pool exhausted")
		}

		p.activeConns++
		p.metricsRegistry.SetDatabaseConnectionPoolSize("main", p.activeConns)
		return createConnection(), nil
	}

	func (p *DatabasePool) ReleaseConnection(conn *Connection) {
		p.activeConns--
		p.metricsRegistry.SetDatabaseConnectionPoolSize("main", p.activeConns)
		conn.Close()
	}

Example 6: Validation Error Tracking

	func ValidateSchedule(schedule Schedule) error {
		if schedule.StartDate.IsZero() {
			metricsRegistry.RecordValidationError("MISSING_START_DATE")
			return errors.New("start date is required")
		}

		if schedule.EndDate.Before(schedule.StartDate) {
			metricsRegistry.RecordValidationError("INVALID_DATE_RANGE")
			return errors.New("end date must be after start date")
		}

		if len(schedule.Shifts) == 0 {
			metricsRegistry.RecordValidationError("EMPTY_SCHEDULE")
			return errors.New("schedule must contain at least one shift")
		}

		return nil
	}

Example 7: Prometheus Queries

Query 1: Requests Per Second (Last 5 Minutes)
	rate(http_requests_total[5m])

Query 2: 95th Percentile Request Latency
	histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

Query 3: Error Rate Percentage
	(
		sum(rate(http_errors_total[5m]))
		/
		sum(rate(http_requests_total[5m]))
	) * 100

Query 4: Average Queries Per Operation
	sum(rate(query_count_per_operation_sum[5m]))
	/
	sum(rate(query_count_per_operation_count[5m]))

Query 5: Slow Operations (>5 seconds)
	avg(rate(service_operation_duration_seconds_sum[5m]) / rate(service_operation_duration_seconds_count[5m]) > 5)

Query 6: Peak Active Scrapers
	max(active_scrape_jobs)

Query 7: Total Validation Errors By Type (1 Hour)
	sum by (error_code) (increase(validation_errors_total[1h]))

Query 8: Database Operations By Type (Last 5 Minutes)
	sum by (operation) (rate(database_operations_total[5m]))

Example 8: Custom Middleware Implementation

	type MetricsMiddleware struct {
		metrics *MetricsRegistry
		next    http.Handler
	}

	func (m *MetricsMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Call next handler
		m.next.ServeHTTP(w, r)

		// Record metrics
		duration := time.Since(startTime).Seconds()
		m.metrics.RecordHTTPRequest(r.Method, r.URL.Path, http.StatusOK, duration)

		// Additional context recording
		if r.Header.Get("X-Error") == "true" {
			m.metrics.RecordHTTPError("client_error")
		}
	}

Example 9: Batch Operation Monitoring

	func ImportMultipleSchedules(files []string) error {
		total := len(files)
		succeeded := 0
		failed := 0

		for _, file := range files {
			startTime := time.Now()
			if err := ImportODSSchedule(file); err != nil {
				failed++
				metricsRegistry.RecordServiceOperation("ods", "batch_import",
					time.Since(startTime).Seconds(), true)
			} else {
				succeeded++
				metricsRegistry.RecordServiceOperation("ods", "batch_import",
					time.Since(startTime).Seconds(), false)
			}
		}

		metricsRegistry.RecordHTTPRequest("POST", "/api/schedules/batch",
			200, time.Since(time.Now()).Seconds())

		return fmt.Errorf("imported %d/%d files, %d failed", succeeded, total, failed)
	}

Example 10: Real-time Queue Monitoring Loop

	func MonitorQueueDepth(jobQueue chan Job, ticker *time.Ticker) {
		for range ticker.C {
			depth := len(jobQueue)
			metricsRegistry.SetQueueDepth("import_jobs", depth)

			// Alert if queue is getting too deep
			if depth > 100 {
				metricsRegistry.RecordHTTPError("queue_backlog_warning")
				log.Printf("WARNING: Queue depth at %d", depth)
			}
		}
	}
*/
