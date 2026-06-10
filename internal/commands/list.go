package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/hev/freshtime/internal/api"
	"github.com/hev/freshtime/internal/config"
)

// ListCmd returns the list command, showing recent time entries with their IDs.
func ListCmd() *cobra.Command {
	var days int
	var client string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent time entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(days, client)
		},
	}

	cmd.Flags().IntVar(&days, "days", 7, "How many days back to list")
	cmd.Flags().StringVar(&client, "client", "", "Filter by client name (substring, case-insensitive) or numeric ID")

	return cmd
}

func runList(days int, client string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	now := time.Now()
	from := now.AddDate(0, 0, -(days - 1)).Format("2006-01-02")
	to := now.Format("2006-01-02")

	http := api.NewClient(cfg)
	entries, err := api.ListTimeEntries(http, cfg.BusinessID, from, to)
	if err != nil {
		return err
	}

	clientNames, err := api.ListClients(http, cfg.AccountID)
	if err != nil {
		return err
	}

	clientID, _, err := resolveClient(clientNames, client)
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

	if len(entries) == 0 {
		fmt.Printf("No entries between %s and %s.\n", from, to)
		return nil
	}

	sort.Slice(entries, func(i, j int) bool {
		di, _ := entryDate(entries[i])
		dj, _ := entryDate(entries[j])
		if di != dj {
			return di < dj
		}
		return entries[i].ID < entries[j].ID
	})

	fmt.Printf("%-12s%-12s%7s  %-20s%-4s%s\n", "ID", "Date", "Hours", "Client", "", "Note")
	fmt.Println(strings.Repeat("─", 80))
	totalSec := 0
	for _, e := range entries {
		d, _ := entryDate(e)
		name := clientNames[e.ClientID]
		if name == "" {
			name = fmt.Sprintf("#%d", e.ClientID)
		}
		if len(name) > 18 {
			name = name[:18]
		}
		billable := ""
		if !e.Billable {
			billable = "NB"
		}
		fmt.Printf("%-12d%-12s%7.2f  %-20s%-4s%s\n",
			e.ID, d, float64(e.Duration)/3600, name, billable, e.Note)
		totalSec += e.Duration
	}
	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("%d entries, %.2fh total\n", len(entries), float64(totalSec)/3600)
	return nil
}
