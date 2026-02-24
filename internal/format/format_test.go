package format

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatHours(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "—"},
		{1.0, "1.0"},
		{8.5, "8.5"},
		{0.5, "0.5"},
		{10.25, "10.2"}, // rounds to 1 decimal
	}
	for _, tt := range tests {
		got := formatHours(tt.input)
		if got != tt.want {
			t.Errorf("formatHours(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatDateRange(t *testing.T) {
	got := formatDateRange("2026-02-23", "2026-02-27")
	want := "Feb 23 – Feb 27, 2026"
	if got != want {
		t.Errorf("formatDateRange = %q, want %q", got, want)
	}
}

func TestTable(t *testing.T) {
	summary := &WeeklySummary{
		WeekStart:  "2026-02-23",
		WeekEnd:    "2026-02-27",
		GrandTotal: 24.0,
		Clients: []ClientSummary{
			{Name: "Acme Corp", Daily: []float64{8.0, 8.0, 0, 0, 0}, Total: 16.0},
			{Name: "Beta Inc", Daily: []float64{0, 0, 4.0, 4.0, 0}, Total: 8.0},
		},
	}

	result := Table(summary)

	// Check structural elements
	if !strings.Contains(result, "Week of Feb 23 – Feb 27, 2026") {
		t.Error("missing date range header")
	}
	if !strings.Contains(result, "Client") {
		t.Error("missing Client column header")
	}
	for _, day := range []string{"Mon", "Tue", "Wed", "Thu", "Fri"} {
		if !strings.Contains(result, day) {
			t.Errorf("missing day header: %s", day)
		}
	}
	if !strings.Contains(result, "Total") {
		t.Error("missing Total column")
	}
	if !strings.Contains(result, "Acme Corp") {
		t.Error("missing client name Acme Corp")
	}
	if !strings.Contains(result, "Beta Inc") {
		t.Error("missing client name Beta Inc")
	}
	// Zero hours should show dash
	if !strings.Contains(result, "—") {
		t.Error("zero hours should display as dash")
	}
	// Separator lines
	if !strings.Contains(result, "─") {
		t.Error("missing separator line")
	}
}

func TestTableTruncatesLongNames(t *testing.T) {
	summary := &WeeklySummary{
		WeekStart:  "2026-02-23",
		WeekEnd:    "2026-02-27",
		GrandTotal: 8.0,
		Clients: []ClientSummary{
			{Name: "Very Long Client Name That Exceeds Twenty Chars", Daily: []float64{8.0, 0, 0, 0, 0}, Total: 8.0},
		},
	}

	result := Table(summary)
	// Name should be truncated to 20 chars
	if strings.Contains(result, "Very Long Client Name That Exceeds Twenty Chars") {
		t.Error("long client name should be truncated")
	}
}

func TestJSON(t *testing.T) {
	summary := &WeeklySummary{
		WeekStart:  "2026-02-23",
		WeekEnd:    "2026-02-27",
		GrandTotal: 8.0,
		Clients: []ClientSummary{
			{Name: "Acme Corp", Daily: []float64{8.0, 0, 0, 0, 0}, Total: 8.0},
		},
	}

	result := JSON(summary)

	// Should be valid JSON
	var parsed WeeklySummary
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("JSON output is not valid JSON: %v", err)
	}

	// Verify round-trip
	if parsed.WeekStart != "2026-02-23" {
		t.Errorf("weekStart = %q, want %q", parsed.WeekStart, "2026-02-23")
	}
	if parsed.GrandTotal != 8.0 {
		t.Errorf("grandTotal = %v, want 8.0", parsed.GrandTotal)
	}
	if len(parsed.Clients) != 1 {
		t.Fatalf("clients len = %d, want 1", len(parsed.Clients))
	}
	if parsed.Clients[0].Name != "Acme Corp" {
		t.Errorf("client name = %q, want %q", parsed.Clients[0].Name, "Acme Corp")
	}
}
