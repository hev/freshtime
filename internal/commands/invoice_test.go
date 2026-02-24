package commands

import (
	"testing"

	"github.com/hev/freshtime/internal/api"
)

var sampleEntries = []api.TimeEntry{
	{
		ID:             1,
		ClientID:       100,
		Duration:       7200, // 2 hours
		StartedAt:      "2026-02-09T09:00:00Z",
		LocalStartedAt: "2026-02-09T09:00:00",
		Note:           "Frontend work",
		Billable:       true,
	},
	{
		ID:             2,
		ClientID:       100,
		Duration:       5400, // 1.5 hours
		StartedAt:      "2026-02-10T10:00:00Z",
		LocalStartedAt: "2026-02-10T10:00:00",
		Note:           "",
		Billable:       true,
	},
}

func TestBuildInvoiceLines(t *testing.T) {
	t.Run("creates one line per entry with correct fields", func(t *testing.T) {
		lines := buildInvoiceLines(sampleEntries, "150.00", "USD")

		if len(lines) != 2 {
			t.Fatalf("expected 2 lines, got %d", len(lines))
		}
		line := lines[0]
		if line.Type != 0 {
			t.Errorf("type = %d, want 0", line.Type)
		}
		if line.Name != "Frontend work" {
			t.Errorf("name = %q, want %q", line.Name, "Frontend work")
		}
		if line.Description != "2026-02-09" {
			t.Errorf("description = %q, want %q", line.Description, "2026-02-09")
		}
		if line.Qty != "2.00" {
			t.Errorf("qty = %q, want %q", line.Qty, "2.00")
		}
		if line.UnitCost.Amount != "150.00" {
			t.Errorf("unit_cost.amount = %q, want %q", line.UnitCost.Amount, "150.00")
		}
		if line.UnitCost.Code != "USD" {
			t.Errorf("unit_cost.code = %q, want %q", line.UnitCost.Code, "USD")
		}
	})

	t.Run("uses Consulting when note is empty", func(t *testing.T) {
		lines := buildInvoiceLines(sampleEntries, "150.00", "USD")
		if lines[1].Name != "Consulting" {
			t.Errorf("name = %q, want %q", lines[1].Name, "Consulting")
		}
	})

	t.Run("converts duration to hours with 2 decimal places", func(t *testing.T) {
		entries := []api.TimeEntry{
			{
				ID:             1,
				ClientID:       100,
				Duration:       2700, // 0.75 hours
				StartedAt:      "2026-02-09T09:00:00Z",
				LocalStartedAt: "2026-02-09T09:00:00",
				Note:           "Task",
				Billable:       true,
			},
		}
		lines := buildInvoiceLines(entries, "100.00", "CAD")
		if lines[0].Qty != "0.75" {
			t.Errorf("qty = %q, want %q", lines[0].Qty, "0.75")
		}
		if lines[0].UnitCost.Code != "CAD" {
			t.Errorf("unit_cost.code = %q, want %q", lines[0].UnitCost.Code, "CAD")
		}
	})
}

func TestSplitDateTime(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2026-02-09T09:00:00", "2026-02-09"},
		{"2026-02-09", "2026-02-09"},
		{"short", "short"},
		{"", ""},
	}
	for _, tt := range tests {
		got := splitDateTime(tt.input)
		if got != tt.want {
			t.Errorf("splitDateTime(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
