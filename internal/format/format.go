package format

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// WeeklySummary holds the weekly time data.
type WeeklySummary struct {
	WeekStart  string          `json:"weekStart"`
	WeekEnd    string          `json:"weekEnd"`
	Clients    []ClientSummary `json:"clients"`
	GrandTotal float64         `json:"grandTotal"`
}

// ClientSummary holds per-client daily hours.
type ClientSummary struct {
	Name  string    `json:"name"`
	Daily []float64 `json:"daily"` // Mon–Fri (5 elements)
	Total float64   `json:"total"`
}

var dayHeaders = []string{"Mon", "Tue", "Wed", "Thu", "Fri"}

func formatHours(h float64) string {
	if h == 0 {
		return "—"
	}
	return fmt.Sprintf("%.1f", h)
}

func formatDateRange(start, end string) string {
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	s, _ := time.Parse("2006-01-02", start)
	e, _ := time.Parse("2006-01-02", end)
	return fmt.Sprintf("%s %d – %s %d, %d", months[s.Month()-1], s.Day(), months[e.Month()-1], e.Day(), e.Year())
}

// Table renders a WeeklySummary as a formatted text table.
func Table(summary *WeeklySummary) string {
	const colWidth = 6
	const nameWidth = 20

	var lines []string

	lines = append(lines, fmt.Sprintf("Week of %s", formatDateRange(summary.WeekStart, summary.WeekEnd)))
	lines = append(lines, "")

	// Header
	header := fmt.Sprintf("%-*s", nameWidth, "Client")
	for _, d := range dayHeaders {
		header += fmt.Sprintf("%*s", colWidth, d)
	}
	header += "  Total"
	lines = append(lines, header)

	separator := strings.Repeat("─", len(header))
	lines = append(lines, separator)

	// Client rows
	for _, client := range summary.Clients {
		name := client.Name
		if len(name) > nameWidth {
			name = name[:nameWidth]
		}
		row := fmt.Sprintf("%-*s", nameWidth, name)
		for _, h := range client.Daily {
			row += fmt.Sprintf("%*s", colWidth, formatHours(h))
		}
		row += fmt.Sprintf("%*s", colWidth+1, formatHours(client.Total))
		row += "h"
		lines = append(lines, row)
	}

	lines = append(lines, separator)

	// Totals row
	dailyTotals := make([]float64, 5)
	for _, client := range summary.Clients {
		for i := 0; i < 5; i++ {
			dailyTotals[i] += client.Daily[i]
		}
	}
	totals := fmt.Sprintf("%-*s", nameWidth, "Total")
	for _, h := range dailyTotals {
		totals += fmt.Sprintf("%*s", colWidth, formatHours(h))
	}
	totals += fmt.Sprintf("%*s", colWidth+1, formatHours(summary.GrandTotal))
	totals += "h"
	lines = append(lines, totals)

	return strings.Join(lines, "\n")
}

// JSON renders a WeeklySummary as indented JSON.
func JSON(summary *WeeklySummary) string {
	data, _ := json.MarshalIndent(summary, "", "  ")
	return string(data)
}
