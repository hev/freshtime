package commands

import (
	"testing"
	"time"

	"github.com/hev/freshtime/internal/api"
	"github.com/hev/freshtime/internal/format"
)

func TestGetWeekRange(t *testing.T) {
	tests := []struct {
		name      string
		input     time.Time
		wantStart string
		wantEnd   string
	}{
		{
			name:      "Wednesday",
			input:     time.Date(2026, 2, 11, 12, 0, 0, 0, time.UTC),
			wantStart: "2026-02-09",
			wantEnd:   "2026-02-13",
		},
		{
			name:      "Monday",
			input:     time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC),
			wantStart: "2026-02-09",
			wantEnd:   "2026-02-13",
		},
		{
			name:      "Friday",
			input:     time.Date(2026, 2, 13, 23, 59, 59, 0, time.UTC),
			wantStart: "2026-02-09",
			wantEnd:   "2026-02-13",
		},
		{
			name:      "Sunday goes to previous week",
			input:     time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC),
			wantStart: "2026-02-09",
			wantEnd:   "2026-02-13",
		},
		{
			name:      "Saturday",
			input:     time.Date(2026, 2, 14, 12, 0, 0, 0, time.UTC),
			wantStart: "2026-02-09",
			wantEnd:   "2026-02-13",
		},
		{
			name:      "month boundary",
			input:     time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC),
			wantStart: "2026-03-02",
			wantEnd:   "2026-03-06",
		},
		{
			name:      "year boundary",
			input:     time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC),
			wantStart: "2025-12-29",
			wantEnd:   "2026-01-02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := getWeekRange(tt.input)
			if start != tt.wantStart {
				t.Errorf("weekStart = %q, want %q", start, tt.wantStart)
			}
			if end != tt.wantEnd {
				t.Errorf("weekEnd = %q, want %q", end, tt.wantEnd)
			}
		})
	}
}

func findClient(clients []format.ClientSummary, name string) *format.ClientSummary {
	for i := range clients {
		if clients[i].Name == name {
			return &clients[i]
		}
	}
	return nil
}

func TestBuildSummary(t *testing.T) {
	clientNames := map[int]string{
		1: "Acme Corp",
		2: "Globex Inc",
	}

	t.Run("groups entries by client", func(t *testing.T) {
		entries := []api.TimeEntry{
			{ID: 1, ClientID: 1, Duration: 7200, StartedAt: "2026-02-09T09:00:00Z"},
			{ID: 2, ClientID: 1, Duration: 3600, StartedAt: "2026-02-10T10:00:00Z"},
			{ID: 3, ClientID: 2, Duration: 5400, StartedAt: "2026-02-09T14:00:00Z"},
		}

		summary := buildSummary(entries, clientNames, "2026-02-09")

		if len(summary.Clients) != 2 {
			t.Fatalf("expected 2 clients, got %d", len(summary.Clients))
		}

		acme := findClient(summary.Clients, "Acme Corp")
		globex := findClient(summary.Clients, "Globex Inc")
		if acme == nil || globex == nil {
			t.Fatal("missing expected client")
		}
		if acme.Daily[0] != 2 {
			t.Errorf("Acme Mon = %v, want 2", acme.Daily[0])
		}
		if acme.Daily[1] != 1 {
			t.Errorf("Acme Tue = %v, want 1", acme.Daily[1])
		}
		if acme.Total != 3 {
			t.Errorf("Acme total = %v, want 3", acme.Total)
		}
		if globex.Daily[0] != 1.5 {
			t.Errorf("Globex Mon = %v, want 1.5", globex.Daily[0])
		}
		if globex.Total != 1.5 {
			t.Errorf("Globex total = %v, want 1.5", globex.Total)
		}
	})

	t.Run("zero-entry week", func(t *testing.T) {
		summary := buildSummary(nil, clientNames, "2026-02-09")
		if len(summary.Clients) != 0 {
			t.Errorf("expected 0 clients, got %d", len(summary.Clients))
		}
		if summary.GrandTotal != 0 {
			t.Errorf("grandTotal = %v, want 0", summary.GrandTotal)
		}
		if summary.WeekStart != "2026-02-09" {
			t.Errorf("weekStart = %q, want %q", summary.WeekStart, "2026-02-09")
		}
		if summary.WeekEnd != "2026-02-13" {
			t.Errorf("weekEnd = %q, want %q", summary.WeekEnd, "2026-02-13")
		}
	})

	t.Run("converts duration seconds to hours", func(t *testing.T) {
		entries := []api.TimeEntry{
			{ID: 1, ClientID: 1, Duration: 5400, StartedAt: "2026-02-09T09:00:00Z"},
			{ID: 2, ClientID: 1, Duration: 900, StartedAt: "2026-02-10T10:00:00Z"},
		}
		summary := buildSummary(entries, clientNames, "2026-02-09")
		acme := summary.Clients[0]
		if acme.Daily[0] != 1.5 {
			t.Errorf("Mon = %v, want 1.5", acme.Daily[0])
		}
		if acme.Daily[1] != 0.25 {
			t.Errorf("Tue = %v, want 0.25", acme.Daily[1])
		}
		if acme.Total != 1.75 {
			t.Errorf("total = %v, want 1.75", acme.Total)
		}
	})

	t.Run("unknown client_id", func(t *testing.T) {
		entries := []api.TimeEntry{
			{ID: 1, ClientID: 999, Duration: 3600, StartedAt: "2026-02-09T09:00:00Z"},
		}
		summary := buildSummary(entries, clientNames, "2026-02-09")
		if summary.Clients[0].Name != "Client #999" {
			t.Errorf("name = %q, want %q", summary.Clients[0].Name, "Client #999")
		}
	})

	t.Run("sums multiple entries same client same day", func(t *testing.T) {
		entries := []api.TimeEntry{
			{ID: 1, ClientID: 1, Duration: 3600, StartedAt: "2026-02-09T09:00:00Z"},
			{ID: 2, ClientID: 1, Duration: 3600, StartedAt: "2026-02-09T14:00:00Z"},
		}
		summary := buildSummary(entries, clientNames, "2026-02-09")
		acme := summary.Clients[0]
		if acme.Daily[0] != 2 {
			t.Errorf("Mon = %v, want 2", acme.Daily[0])
		}
		if acme.Total != 2 {
			t.Errorf("total = %v, want 2", acme.Total)
		}
	})

	t.Run("skips weekend entries", func(t *testing.T) {
		entries := []api.TimeEntry{
			{ID: 1, ClientID: 1, Duration: 3600, StartedAt: "2026-02-14T09:00:00Z"}, // Saturday
			{ID: 2, ClientID: 1, Duration: 3600, StartedAt: "2026-02-15T09:00:00Z"}, // Sunday
		}
		summary := buildSummary(entries, clientNames, "2026-02-09")
		if len(summary.Clients) != 0 {
			t.Errorf("expected 0 clients (weekends skipped), got %d", len(summary.Clients))
		}
		if summary.GrandTotal != 0 {
			t.Errorf("grandTotal = %v, want 0", summary.GrandTotal)
		}
	})

	t.Run("sorts clients alphabetically", func(t *testing.T) {
		entries := []api.TimeEntry{
			{ID: 1, ClientID: 2, Duration: 3600, StartedAt: "2026-02-09T09:00:00Z"},
			{ID: 2, ClientID: 1, Duration: 3600, StartedAt: "2026-02-09T10:00:00Z"},
		}
		summary := buildSummary(entries, clientNames, "2026-02-09")
		if summary.Clients[0].Name != "Acme Corp" {
			t.Errorf("first client = %q, want %q", summary.Clients[0].Name, "Acme Corp")
		}
		if summary.Clients[1].Name != "Globex Inc" {
			t.Errorf("second client = %q, want %q", summary.Clients[1].Name, "Globex Inc")
		}
	})

	t.Run("calculates grandTotal across clients", func(t *testing.T) {
		entries := []api.TimeEntry{
			{ID: 1, ClientID: 1, Duration: 7200, StartedAt: "2026-02-09T09:00:00Z"},
			{ID: 2, ClientID: 2, Duration: 5400, StartedAt: "2026-02-10T10:00:00Z"},
		}
		summary := buildSummary(entries, clientNames, "2026-02-09")
		if summary.GrandTotal != 3.5 {
			t.Errorf("grandTotal = %v, want 3.5", summary.GrandTotal)
		}
	})
}
