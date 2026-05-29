package format

import (
	"encoding/json"
	"fmt"
)

// DailySummary holds per-day hours over a date range. The Days map is keyed by
// YYYY-MM-DD and weekends are included — it's the data source for the hours wall.
type DailySummary struct {
	Client     string             `json:"client"`
	ClientID   int                `json:"clientId,omitempty"`
	From       string             `json:"from"`
	To         string             `json:"to"`
	TotalHours float64            `json:"totalHours"`
	DaysLogged int                `json:"daysLogged"`
	Days       map[string]float64 `json:"days"`
}

// DailyJSON renders a DailySummary as indented JSON.
func DailyJSON(s *DailySummary) string {
	data, _ := json.MarshalIndent(s, "", "  ")
	return string(data)
}

// DailyText renders a one-line human summary of a DailySummary.
func DailyText(s *DailySummary) string {
	return fmt.Sprintf("%s · %s – %s\n%.1fh across %d day(s)",
		s.Client, s.From, s.To, s.TotalHours, s.DaysLogged)
}
