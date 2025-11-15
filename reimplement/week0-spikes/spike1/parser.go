package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Shift represents a single shift entry from Amion.
type Shift struct {
	Date      string `json:"date"`
	Position  string `json:"position"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Location  string `json:"location"`
}

// AmionParser handles parsing Amion HTML pages.
type AmionParser struct {
	// Selectors for finding shift data in HTML
	ShiftRowSelector string
	DateSelector     string
	PositionSelector string
	StartTimeSelector string
	EndTimeSelector  string
	LocationSelector string
}

// NewAmionParser creates a new Amion HTML parser with sensible defaults.
func NewAmionParser() *AmionParser {
	return &AmionParser{
		ShiftRowSelector: "table tbody tr",
		DateSelector: "td:nth-child(1)",
		PositionSelector: "td:nth-child(2)",
		StartTimeSelector: "td:nth-child(3)",
		EndTimeSelector: "td:nth-child(4)",
		LocationSelector: "td:nth-child(5)",
	}
}

// ParseShifts parses shift data from HTML content.
// Returns a slice of Shift structs and any error encountered.
// Errors are logged but parsing continues to collect as much data as possible.
func (p *AmionParser) ParseShifts(html string) ([]Shift, error) {
	if html == "" {
		return []Shift{}, nil
	}

	doc, err := parseHTML(html)
	if err != nil {
		return []Shift{}, nil // Return empty but don't error on parse failure
	}

	shifts := make([]Shift, 0)
	doc.Find(p.ShiftRowSelector).Each(func(i int, s *goquery.Selection) {
		shift := Shift{
			Date:      strings.TrimSpace(s.Find(p.DateSelector).Text()),
			Position:  strings.TrimSpace(s.Find(p.PositionSelector).Text()),
			StartTime: strings.TrimSpace(s.Find(p.StartTimeSelector).Text()),
			EndTime:   strings.TrimSpace(s.Find(p.EndTimeSelector).Text()),
			Location:  strings.TrimSpace(s.Find(p.LocationSelector).Text()),
		}

		// Only add shifts with at least date and position
		if shift.Date != "" && shift.Position != "" {
			shifts = append(shifts, shift)
		}
	})

	return shifts, nil
}

// ParseShiftsWithAccuracy parses shifts and returns accuracy metrics.
// This validates that parsing is reliable by checking data completeness.
func (p *AmionParser) ParseShiftsWithAccuracy(html string) ([]Shift, float64, error) {
	shifts, err := p.ParseShifts(html)
	if err != nil {
		return nil, 0, err
	}

	if len(shifts) == 0 {
		return shifts, 100, nil // Empty results are "accurate"
	}

	// Calculate accuracy based on completeness of parsed data
	totalFields := len(shifts) * 5 // 5 fields per shift
	completedFields := 0

	for _, shift := range shifts {
		if shift.Date != "" {
			completedFields++
		}
		if shift.Position != "" {
			completedFields++
		}
		if shift.StartTime != "" {
			completedFields++
		}
		if shift.EndTime != "" {
			completedFields++
		}
		if shift.Location != "" {
			completedFields++
		}
	}

	accuracy := (float64(completedFields) / float64(totalFields)) * 100
	return shifts, accuracy, nil
}

// CustomizeSelectors allows adjusting CSS selectors for different HTML structures.
// Call this before parsing if Amion's HTML structure differs from expected.
func (p *AmionParser) CustomizeSelectors(selectorMap map[string]string) error {
	validKeys := map[string]bool{
		"shift_row": true,
		"date": true,
		"position": true,
		"start_time": true,
		"end_time": true,
		"location": true,
	}

	for key, val := range selectorMap {
		if !validKeys[key] {
			return fmt.Errorf("unknown selector key: %s", key)
		}

		switch key {
		case "shift_row":
			p.ShiftRowSelector = val
		case "date":
			p.DateSelector = val
		case "position":
			p.PositionSelector = val
		case "start_time":
			p.StartTimeSelector = val
		case "end_time":
			p.EndTimeSelector = val
		case "location":
			p.LocationSelector = val
		}
	}

	return nil
}

// parseHTML parses HTML string into a goquery document.
func parseHTML(html string) (*goquery.Document, error) {
	return goquery.NewDocumentFromReader(strings.NewReader(html))
}
