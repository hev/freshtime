package commands

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/hev/freshtime/internal/api"
	"github.com/hev/freshtime/internal/config"
	"github.com/hev/freshtime/internal/format"
)

// WeeklyCmd returns the weekly command.
func WeeklyCmd() *cobra.Command {
	var weekOf string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "weekly",
		Short: "Show weekly time summary grouped by client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWeekly(weekOf, jsonOutput)
		},
	}

	cmd.Flags().StringVar(&weekOf, "week-of", "", "Show week containing this date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func getWeekRange(ref time.Time) (weekStart, weekEnd string) {
	day := ref.Weekday()
	diffToMonday := int(time.Monday - day)
	if day == time.Sunday {
		diffToMonday = -6
	}
	monday := ref.AddDate(0, 0, diffToMonday)
	friday := monday.AddDate(0, 0, 4)

	return monday.Format("2006-01-02"), friday.Format("2006-01-02")
}

func buildSummary(entries []api.TimeEntry, clientNames map[int]string, weekStart string) *format.WeeklySummary {
	monday, _ := time.Parse("2006-01-02", weekStart)
	friday := monday.AddDate(0, 0, 4)
	weekEnd := friday.Format("2006-01-02")

	// Group by client
	byClient := make(map[int][]int) // client_id -> [5]seconds

	for _, entry := range entries {
		localDate := entry.LocalStartedAt
		if localDate == "" {
			localDate = entry.StartedAt
		}
		t, err := time.Parse("2006-01-02T15:04:05", localDate)
		if err != nil {
			// Try ISO with timezone
			t, err = time.Parse(time.RFC3339, localDate)
			if err != nil {
				continue
			}
		}
		dayOfWeek := t.Weekday()
		var dayIndex int
		if dayOfWeek == time.Sunday {
			dayIndex = 6
		} else {
			dayIndex = int(dayOfWeek) - 1 // 0=Mon...6=Sun
		}
		if dayIndex < 0 || dayIndex > 4 {
			continue // skip weekends
		}

		if _, ok := byClient[entry.ClientID]; !ok {
			byClient[entry.ClientID] = make([]int, 5)
		}
		byClient[entry.ClientID][dayIndex] += entry.Duration
	}

	// Build client summaries
	var clients []format.ClientSummary
	var grandTotal float64

	for clientID, dailySeconds := range byClient {
		dailyHours := make([]float64, 5)
		var total float64
		for i, s := range dailySeconds {
			h := math.Round(float64(s)/3600*100) / 100
			dailyHours[i] = h
			total += h
		}
		total = math.Round(total*100) / 100
		grandTotal += total

		name := clientNames[clientID]
		if name == "" {
			name = fmt.Sprintf("Client #%d", clientID)
		}
		clients = append(clients, format.ClientSummary{
			Name:  name,
			Daily: dailyHours,
			Total: total,
		})
	}

	sort.Slice(clients, func(i, j int) bool {
		return strings.ToLower(clients[i].Name) < strings.ToLower(clients[j].Name)
	})
	grandTotal = math.Round(grandTotal*100) / 100

	return &format.WeeklySummary{
		WeekStart:  weekStart,
		WeekEnd:    weekEnd,
		Clients:    clients,
		GrandTotal: grandTotal,
	}
}

func runWeekly(weekOf string, jsonOutput bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	http := api.NewClient(cfg)

	ref := time.Now()
	if weekOf != "" {
		ref, err = time.Parse("2006-01-02", weekOf)
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
	}
	weekStart, weekEnd := getWeekRange(ref)

	entries, err := api.ListTimeEntries(http, cfg.BusinessID, weekStart, weekEnd)
	if err != nil {
		return err
	}

	clientNames, err := api.ListClients(http, cfg.AccountID)
	if err != nil {
		return err
	}

	summary := buildSummary(entries, clientNames, weekStart)

	if jsonOutput {
		fmt.Println(format.JSON(summary))
	} else {
		fmt.Println(format.Table(summary))
	}
	return nil
}
