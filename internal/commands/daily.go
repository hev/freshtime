package commands

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/hev/freshtime/internal/api"
	"github.com/hev/freshtime/internal/config"
	"github.com/hev/freshtime/internal/format"
)

// DailyCmd returns the daily command. It emits per-day hours (weekends
// included) over a date range — the data source for the hours wall.
func DailyCmd() *cobra.Command {
	var from, to, client string
	var weeks int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "daily",
		Short: "Show per-day hours over a date range (data source for the hours wall)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDaily(from, to, client, weeks, jsonOutput)
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Start date YYYY-MM-DD (default: <weeks> back from --to)")
	cmd.Flags().StringVar(&to, "to", "", "End date YYYY-MM-DD (default: today)")
	cmd.Flags().StringVar(&client, "client", "", "Filter by client name (substring, case-insensitive) or numeric ID")
	cmd.Flags().IntVar(&weeks, "weeks", 53, "Weeks back from --to when --from is omitted")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

// dailyRange resolves the [from, to] window, defaulting to the trailing
// `weeks` window ending at `now` when the flags are omitted.
func dailyRange(from, to string, weeks int, now time.Time) (start, end string, err error) {
	endT := now
	if to != "" {
		endT, err = time.Parse("2006-01-02", to)
		if err != nil {
			return "", "", fmt.Errorf("invalid --to date: %w", err)
		}
	}
	startT := endT.AddDate(0, 0, -(weeks*7 - 1))
	if from != "" {
		startT, err = time.Parse("2006-01-02", from)
		if err != nil {
			return "", "", fmt.Errorf("invalid --from date: %w", err)
		}
	}
	if startT.After(endT) {
		return "", "", fmt.Errorf("--from %s is after --to %s",
			startT.Format("2006-01-02"), endT.Format("2006-01-02"))
	}
	return startT.Format("2006-01-02"), endT.Format("2006-01-02"), nil
}

// entryDate extracts the YYYY-MM-DD local date from a time entry, preferring
// the local timestamp so days bucket the way they appear in FreshBooks.
func entryDate(e api.TimeEntry) (string, bool) {
	raw := e.LocalStartedAt
	if raw == "" {
		raw = e.StartedAt
	}
	t, err := time.Parse("2006-01-02T15:04:05", raw)
	if err != nil {
		t, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			return "", false
		}
	}
	return t.Format("2006-01-02"), true
}

// buildDaily buckets entries into per-day hours (weekends included), keeping
// only days within [from, to]. ISO dates compare lexicographically.
func buildDaily(entries []api.TimeEntry, from, to string) (map[string]float64, float64) {
	secByDay := make(map[string]int)
	totalSec := 0
	for _, e := range entries {
		d, ok := entryDate(e)
		if !ok || d < from || d > to {
			continue
		}
		secByDay[d] += e.Duration
		totalSec += e.Duration
	}
	days := make(map[string]float64, len(secByDay))
	for d, s := range secByDay {
		days[d] = math.Round(float64(s)/3600*100) / 100
	}
	return days, math.Round(float64(totalSec)/3600*100) / 100
}

// resolveClient turns a name/ID query into a client ID + display name.
// An empty query means "all clients" (id 0).
func resolveClient(clientNames map[int]string, query string) (int, string, error) {
	if query == "" {
		return 0, "All clients", nil
	}
	if id, err := strconv.Atoi(query); err == nil {
		if name, ok := clientNames[id]; ok {
			return id, name, nil
		}
		return id, fmt.Sprintf("Client #%d", id), nil
	}

	q := strings.ToLower(query)
	var matches []int
	exact := -1
	for id, name := range clientNames {
		ln := strings.ToLower(name)
		if ln == q {
			exact = id
		}
		if strings.Contains(ln, q) {
			matches = append(matches, id)
		}
	}
	if exact >= 0 {
		return exact, clientNames[exact], nil
	}
	if len(matches) == 1 {
		return matches[0], clientNames[matches[0]], nil
	}
	if len(matches) == 0 {
		return 0, "", fmt.Errorf("no client matches %q", query)
	}
	names := make([]string, 0, len(matches))
	for _, id := range matches {
		names = append(names, clientNames[id])
	}
	sort.Strings(names)
	return 0, "", fmt.Errorf("multiple clients match %q: %s", query, strings.Join(names, ", "))
}

func runDaily(from, to, client string, weeks int, jsonOutput bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	start, end, err := dailyRange(from, to, weeks, time.Now())
	if err != nil {
		return err
	}

	http := api.NewClient(cfg)

	entries, err := api.ListTimeEntries(http, cfg.BusinessID, start, end)
	if err != nil {
		return err
	}

	clientNames, err := api.ListClients(http, cfg.AccountID)
	if err != nil {
		return err
	}

	clientID, clientName, err := resolveClient(clientNames, client)
	if err != nil {
		return err
	}

	if clientID != 0 {
		filtered := make([]api.TimeEntry, 0, len(entries))
		for _, e := range entries {
			if e.ClientID == clientID {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	days, total := buildDaily(entries, start, end)
	summary := &format.DailySummary{
		Client:     clientName,
		ClientID:   clientID,
		From:       start,
		To:         end,
		TotalHours: total,
		DaysLogged: len(days),
		Days:       days,
	}

	if jsonOutput {
		fmt.Println(format.DailyJSON(summary))
	} else {
		fmt.Println(format.DailyText(summary))
	}
	return nil
}
