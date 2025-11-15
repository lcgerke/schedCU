// Package result defines spike result types and reporting.
package result

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Status represents the outcome of a spike.
type Status string

const (
	StatusSuccess Status = "success"
	StatusWarning Status = "warning"
	StatusFailure Status = "failure"
)

// SpikeResult represents the result of a single spike execution.
type SpikeResult struct {
	// Metadata
	SpikeID      string    `json:"spike_id"`      // e.g., "spike1"
	SpikeName    string    `json:"spike_name"`    // e.g., "Amion HTML Scraping"
	ExecutedAt   time.Time `json:"executed_at"`
	ExecutedBy   string    `json:"executed_by"`
	Duration     int64     `json:"duration_ms"`    // Milliseconds
	Environment  string    `json:"environment"`    // "mock" or "real"

	// Results
	Status          Status                 `json:"status"`           // success, warning, failure
	Findings        map[string]interface{} `json:"findings"`         // Key metrics
	Recommendation  string                 `json:"recommendation"`   // Next steps
	TimecostWeeks   int                    `json:"timecost_weeks"`   // If fallback needed
	Evidence        []string               `json:"evidence"`         // Links to logs/graphs
	DetailedResults string                 `json:"detailed_results"` // Full markdown doc

	// Errors
	Errors []string `json:"errors,omitempty"`
}

// WriteJSON writes the result as JSON to the specified file.
func (r *SpikeResult) WriteJSON(filepath string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("write json: %w", err)
	}

	return nil
}

// WriteMarkdown writes the result as a markdown document.
func (r *SpikeResult) WriteMarkdown(filepath string) error {
	markdown := fmt.Sprintf(`# %s Results

**Spike ID**: %s
**Status**: %s
**Executed**: %s
**Duration**: %dms
**Environment**: %s

## Summary

**Recommendation**: %s

**Timeline Impact**: +%d weeks (if fallback needed)

## Findings

`, r.SpikeName, r.SpikeID, r.Status, r.ExecutedAt.Format(time.RFC3339), r.Duration, r.Environment,
		r.Recommendation, r.TimecostWeeks)

	// Add findings as key-value pairs
	for key, val := range r.Findings {
		markdown += fmt.Sprintf("- **%s**: %v\n", key, val)
	}

	markdown += "\n## Evidence\n\n"
	if len(r.Evidence) > 0 {
		for _, evidence := range r.Evidence {
			markdown += fmt.Sprintf("- %s\n", evidence)
		}
	} else {
		markdown += "None yet.\n"
	}

	if len(r.Errors) > 0 {
		markdown += "\n## Errors\n\n"
		for _, err := range r.Errors {
			markdown += fmt.Sprintf("- %s\n", err)
		}
	}

	markdown += fmt.Sprintf("\n## Detailed Results\n\n%s\n", r.DetailedResults)

	if err := os.WriteFile(filepath, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}

	return nil
}

// WriteResults writes both JSON and Markdown results to the results directory.
func (r *SpikeResult) WriteResults(resultsDir string) error {
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return fmt.Errorf("create results dir: %w", err)
	}

	jsonFile := filepath.Join(resultsDir, fmt.Sprintf("%s_results.json", r.SpikeID))
	if err := r.WriteJSON(jsonFile); err != nil {
		return fmt.Errorf("write json: %w", err)
	}

	mdFile := filepath.Join(resultsDir, fmt.Sprintf("%s_results.md", r.SpikeID))
	if err := r.WriteMarkdown(mdFile); err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}

	return nil
}

// Summary generates a brief text summary of the result.
func (r *SpikeResult) Summary() string {
	return fmt.Sprintf("[%s] %s: %s (timeline cost: +%d weeks)", r.SpikeID, r.SpikeName, r.Status, r.TimecostWeeks)
}

// NewResult creates a new spike result.
func NewResult(spikeID, spikeName, environment string) *SpikeResult {
	return &SpikeResult{
		SpikeID:      spikeID,
		SpikeName:    spikeName,
		ExecutedAt:   time.Now(),
		ExecutedBy:   os.Getenv("USER"),
		Environment:  environment,
		Findings:     make(map[string]interface{}),
		Evidence:     make([]string, 0),
		Errors:       make([]string, 0),
		Status:       StatusSuccess,
		TimecostWeeks: 0,
	}
}

// AddFinding adds a finding to the result.
func (r *SpikeResult) AddFinding(key string, value interface{}) {
	r.Findings[key] = value
}

// AddError adds an error message to the result.
func (r *SpikeResult) AddError(msg string, args ...interface{}) {
	r.Errors = append(r.Errors, fmt.Sprintf(msg, args...))
}

// AddEvidence adds an evidence link to the result.
func (r *SpikeResult) AddEvidence(link string) {
	r.Evidence = append(r.Evidence, link)
}

// FailWith marks the result as failed with a message and optional timeline cost.
func (r *SpikeResult) FailWith(recommendation string, timeWeeks int) {
	r.Status = StatusFailure
	r.Recommendation = recommendation
	r.TimecostWeeks = timeWeeks
}

// WarnWith marks the result as warning with a message and optional timeline cost.
func (r *SpikeResult) WarnWith(recommendation string, timeWeeks int) {
	r.Status = StatusWarning
	r.Recommendation = recommendation
	r.TimecostWeeks = timeWeeks
}

// SucceedWith marks the result as successful with a message.
func (r *SpikeResult) SucceedWith(recommendation string) {
	r.Status = StatusSuccess
	r.Recommendation = recommendation
	r.TimecostWeeks = 0
}
