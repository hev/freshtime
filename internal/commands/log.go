package commands

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/hev/freshtime/internal/api"
	"github.com/hev/freshtime/internal/config"
)

// LogCmd returns the log command.
func LogCmd() *cobra.Command {
	var (
		message   string
		duration  string
		client    int
		project   int
		service   int
		noBillable bool
	)

	cmd := &cobra.Command{
		Use:   "log",
		Short: "Log a time entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLog(message, duration, client, project, service, noBillable)
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Note for the time entry (required)")
	cmd.Flags().StringVarP(&duration, "duration", "d", "", "Duration (e.g. 2h, 30m, 1h30m)")
	cmd.Flags().IntVar(&client, "client", 0, "Client ID (overrides .freshtime.json)")
	cmd.Flags().IntVar(&project, "project", 0, "Project ID (overrides .freshtime.json)")
	cmd.Flags().IntVar(&service, "service", 0, "Service ID (overrides .freshtime.json)")
	cmd.Flags().BoolVar(&noBillable, "no-billable", false, "Mark as non-billable")
	cmd.MarkFlagRequired("message")
	cmd.MarkFlagRequired("duration")

	return cmd
}

func runLog(message, duration string, clientID, projectID, serviceID int, noBillable bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Load project config for defaults
	pc, _ := config.LoadProjectConfigFromCwd()
	if pc != nil {
		if clientID == 0 {
			clientID = pc.ClientID
		}
		if projectID == 0 {
			projectID = pc.ProjectID
		}
		if serviceID == 0 {
			serviceID = pc.ServiceID
		}
	}

	if clientID == 0 {
		return fmt.Errorf("no client specified. Use --client or run `freshtime init` to create .freshtime.json")
	}

	seconds, err := parseDuration(duration)
	if err != nil {
		return err
	}

	http := api.NewClient(cfg)
	entry, err := api.CreateTimeEntry(http, cfg.BusinessID, api.CreateTimeEntryRequest{
		ClientID:  clientID,
		ProjectID: projectID,
		ServiceID: serviceID,
		Duration:  seconds,
		Note:      message,
		Billable:  !noBillable,
		StartedAt: time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	})
	if err != nil {
		return fmt.Errorf("failed to create time entry: %w", err)
	}

	hours := float64(seconds) / 3600
	fmt.Printf("Logged %.2fh: %s (entry #%d)\n", hours, message, entry.ID)
	return nil
}

// parseDuration parses a human-friendly duration string like "2h", "30m", "1h30m".
func parseDuration(s string) (int, error) {
	re := regexp.MustCompile(`^(?:(\d+)h)?(?:(\d+)m)?$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil || (matches[1] == "" && matches[2] == "") {
		return 0, fmt.Errorf("invalid duration %q (expected format: 2h, 30m, 1h30m)", s)
	}

	var seconds int
	if matches[1] != "" {
		h, _ := strconv.Atoi(matches[1])
		seconds += h * 3600
	}
	if matches[2] != "" {
		m, _ := strconv.Atoi(matches[2])
		seconds += m * 60
	}
	return seconds, nil
}
