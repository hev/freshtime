package api

import (
	"encoding/json"
	"fmt"
)

// TimeEntry represents a FreshBooks time entry.
type TimeEntry struct {
	ID             int    `json:"id"`
	ClientID       int    `json:"client_id"`
	Duration       int    `json:"duration"` // seconds
	StartedAt      string `json:"started_at"`
	LocalStartedAt string `json:"local_started_at"`
	Note           string `json:"note"`
	Billable       bool   `json:"billable"`
}

// ListTimeEntries fetches time entries for a date range.
func ListTimeEntries(c *HttpClient, businessID int, startedFrom, startedTo string) ([]TimeEntry, error) {
	path := fmt.Sprintf("/timetracking/business/%d/time_entries", businessID)
	raw, err := c.GetPaginated(path, "time_entries", map[string]string{
		"started_from": startedFrom + "T00:00:00",
		"started_to":   startedTo + "T23:59:59",
	})
	if err != nil {
		return nil, err
	}

	entries := make([]TimeEntry, 0, len(raw))
	for _, r := range raw {
		var te TimeEntry
		if err := json.Unmarshal(r, &te); err != nil {
			continue
		}
		entries = append(entries, te)
	}
	return entries, nil
}

// ListUnbilledEntries fetches unbilled, billable time entries for a client.
func ListUnbilledEntries(c *HttpClient, businessID, clientID int) ([]TimeEntry, error) {
	path := fmt.Sprintf("/timetracking/business/%d/time_entries", businessID)
	raw, err := c.GetPaginated(path, "time_entries", map[string]string{
		"client_id": fmt.Sprintf("%d", clientID),
		"billed":    "false",
		"billable":  "true",
	})
	if err != nil {
		return nil, err
	}

	entries := make([]TimeEntry, 0, len(raw))
	for _, r := range raw {
		var te TimeEntry
		if err := json.Unmarshal(r, &te); err != nil {
			continue
		}
		entries = append(entries, te)
	}
	return entries, nil
}

// MarkEntriesAsBilled marks each entry as billed via the API.
func MarkEntriesAsBilled(c *HttpClient, businessID int, entries []TimeEntry) error {
	for _, entry := range entries {
		path := fmt.Sprintf("/timetracking/business/%d/time_entries/%d", businessID, entry.ID)
		body := map[string]any{
			"time_entry": map[string]any{
				"billed":     true,
				"started_at": entry.StartedAt,
				"is_logged":  true,
				"duration":   entry.Duration,
			},
		}
		if err := c.Put(path, body, nil); err != nil {
			return err
		}
	}
	return nil
}
