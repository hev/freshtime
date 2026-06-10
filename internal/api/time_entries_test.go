package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTimeEntry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/timetracking/business/42/time_entries/101" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"time_entry": map[string]any{
				"id": 101, "client_id": 9, "duration": 3600,
				"note": "pairing session", "billable": true,
				"started_at": "2026-06-09T19:00:00Z",
			},
		})
	}))
	defer srv.Close()

	origBase := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = origBase }()

	c := NewHttpClient("test-token")
	entry, err := GetTimeEntry(c, 42, 101)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.ID != 101 || entry.Duration != 3600 || entry.Note != "pairing session" {
		t.Errorf("unexpected entry: %+v", entry)
	}
}

func TestUpdateTimeEntry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/timetracking/business/42/time_entries/101" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body struct {
			TimeEntry map[string]any `json:"time_entry"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if body.TimeEntry["note"] != "updated note" {
			t.Errorf("unexpected note: %v", body.TimeEntry["note"])
		}
		if body.TimeEntry["is_logged"] != true {
			t.Errorf("expected is_logged=true, got %v", body.TimeEntry["is_logged"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"time_entry": map[string]any{"id": 101, "duration": 7200, "note": "updated note"},
		})
	}))
	defer srv.Close()

	origBase := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = origBase }()

	c := NewHttpClient("test-token")
	entry, err := UpdateTimeEntry(c, 42, 101, CreateTimeEntryRequest{
		ClientID: 9, Duration: 7200, Note: "updated note",
		Billable: true, StartedAt: "2026-06-09T19:00:00Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Duration != 7200 {
		t.Errorf("expected duration 7200, got %d", entry.Duration)
	}
}

func TestDeleteTimeEntry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/timetracking/business/42/time_entries/101" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	origBase := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = origBase }()

	c := NewHttpClient("test-token")
	if err := DeleteTimeEntry(c, 42, 101); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
