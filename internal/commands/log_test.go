package commands

import (
	"strings"
	"testing"
	"time"
)

func TestEntryStartedAtDefaultsToNow(t *testing.T) {
	now := time.Date(2026, 6, 10, 15, 30, 45, 0, time.UTC)
	got, err := entryStartedAt("", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "2026-06-10T15:30:45Z" {
		t.Errorf("expected now timestamp, got %q", got)
	}
}

func TestEntryStartedAtBackdates(t *testing.T) {
	now := time.Date(2026, 6, 10, 15, 30, 45, 0, time.UTC)
	got, err := entryStartedAt("2026-06-01", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Pinned to noon local on the given date; the UTC date stays within a
	// day of the requested one for any sane timezone offset.
	if !strings.HasPrefix(got, "2026-06-01T") && !strings.HasPrefix(got, "2026-06-02T") && !strings.HasPrefix(got, "2026-05-31T") {
		t.Errorf("expected timestamp near 2026-06-01, got %q", got)
	}
	if !strings.HasSuffix(got, "Z") {
		t.Errorf("expected UTC timestamp, got %q", got)
	}
}

func TestEntryStartedAtRejectsBadDate(t *testing.T) {
	now := time.Date(2026, 6, 10, 15, 30, 45, 0, time.UTC)
	if _, err := entryStartedAt("June 1", now); err == nil {
		t.Error("expected error for invalid date")
	}
	if _, err := entryStartedAt("2026-13-40", now); err == nil {
		t.Error("expected error for out-of-range date")
	}
}
