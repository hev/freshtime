package commands

import (
	"testing"
	"time"

	"github.com/hev/freshtime/internal/api"
)

func TestDailyRange(t *testing.T) {
	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)

	t.Run("defaults to trailing weeks ending today", func(t *testing.T) {
		start, end, err := dailyRange("", "", 53, now)
		if err != nil {
			t.Fatal(err)
		}
		if end != "2026-05-29" {
			t.Errorf("end = %q, want 2026-05-29", end)
		}
		// 53 weeks = 371 days inclusive, so start is 370 days before end.
		if start != "2025-05-24" {
			t.Errorf("start = %q, want 2025-05-24", start)
		}
	})

	t.Run("explicit from/to wins", func(t *testing.T) {
		start, end, err := dailyRange("2026-01-01", "2026-03-31", 53, now)
		if err != nil {
			t.Fatal(err)
		}
		if start != "2026-01-01" || end != "2026-03-31" {
			t.Errorf("got %q–%q, want 2026-01-01–2026-03-31", start, end)
		}
	})

	t.Run("from after to errors", func(t *testing.T) {
		if _, _, err := dailyRange("2026-04-01", "2026-01-01", 53, now); err == nil {
			t.Error("expected error when from is after to")
		}
	})

	t.Run("invalid date errors", func(t *testing.T) {
		if _, _, err := dailyRange("nonsense", "", 53, now); err == nil {
			t.Error("expected error on invalid --from")
		}
	})
}

func TestBuildDaily(t *testing.T) {
	t.Run("includes weekends and sums per day", func(t *testing.T) {
		entries := []api.TimeEntry{
			{ID: 1, ClientID: 1, Duration: 3600, StartedAt: "2026-02-14T09:00:00Z"}, // Saturday
			{ID: 2, ClientID: 1, Duration: 1800, StartedAt: "2026-02-14T14:00:00Z"}, // same Saturday
			{ID: 3, ClientID: 1, Duration: 7200, StartedAt: "2026-02-15T10:00:00Z"}, // Sunday
		}
		days, total := buildDaily(entries, "2026-02-09", "2026-02-15")
		if days["2026-02-14"] != 1.5 {
			t.Errorf("Sat = %v, want 1.5", days["2026-02-14"])
		}
		if days["2026-02-15"] != 2 {
			t.Errorf("Sun = %v, want 2", days["2026-02-15"])
		}
		if total != 3.5 {
			t.Errorf("total = %v, want 3.5", total)
		}
	})

	t.Run("drops entries outside the window", func(t *testing.T) {
		entries := []api.TimeEntry{
			{ID: 1, ClientID: 1, Duration: 3600, StartedAt: "2026-01-31T09:00:00Z"}, // before
			{ID: 2, ClientID: 1, Duration: 3600, StartedAt: "2026-02-10T09:00:00Z"}, // in
			{ID: 3, ClientID: 1, Duration: 3600, StartedAt: "2026-03-01T09:00:00Z"}, // after
		}
		days, total := buildDaily(entries, "2026-02-01", "2026-02-28")
		if len(days) != 1 || days["2026-02-10"] != 1 {
			t.Errorf("days = %v, want only 2026-02-10:1", days)
		}
		if total != 1 {
			t.Errorf("total = %v, want 1", total)
		}
	})

	t.Run("prefers local_started_at", func(t *testing.T) {
		entries := []api.TimeEntry{
			{ID: 1, ClientID: 1, Duration: 3600, StartedAt: "2026-02-11T02:00:00Z", LocalStartedAt: "2026-02-10T19:00:00"},
		}
		days, _ := buildDaily(entries, "2026-02-01", "2026-02-28")
		if days["2026-02-10"] != 1 {
			t.Errorf("expected entry bucketed to local date 2026-02-10, got %v", days)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		days, total := buildDaily(nil, "2026-02-01", "2026-02-28")
		if len(days) != 0 || total != 0 {
			t.Errorf("got days=%v total=%v, want empty", days, total)
		}
	})
}

func TestResolveClient(t *testing.T) {
	clients := map[int]string{
		598461:  "Trio Health",
		1264843: "Freeplay",
		807999:  "Ooli Data",
	}

	tests := []struct {
		name     string
		query    string
		wantID   int
		wantName string
		wantErr  bool
	}{
		{"empty means all", "", 0, "All clients", false},
		{"numeric id", "598461", 598461, "Trio Health", false},
		{"case-insensitive substring", "trio", 598461, "Trio Health", false},
		{"exact name", "Freeplay", 1264843, "Freeplay", false},
		{"no match", "acme", 0, "", true},
		{"unknown numeric id falls back", "999", 999, "Client #999", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, name, err := resolveClient(clients, tt.query)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.query)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if id != tt.wantID || name != tt.wantName {
				t.Errorf("got (%d, %q), want (%d, %q)", id, name, tt.wantID, tt.wantName)
			}
		})
	}
}
