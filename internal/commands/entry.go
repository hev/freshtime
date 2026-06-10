package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/hev/freshtime/internal/api"
	"github.com/hev/freshtime/internal/config"
)

// EditCmd returns the edit command for fixing up an existing time entry.
func EditCmd() *cobra.Command {
	var (
		message  string
		duration string
		date     string
		billable bool
	)

	cmd := &cobra.Command{
		Use:   "edit <entry-id>",
		Short: "Edit a time entry's note, duration, date, or billable flag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid entry ID %q", args[0])
			}
			billablePtr := (*bool)(nil)
			if cmd.Flags().Changed("billable") {
				billablePtr = &billable
			}
			return runEdit(id, message, duration, date, billablePtr, cmd.Flags().Changed("message"))
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "New note for the entry")
	cmd.Flags().StringVarP(&duration, "duration", "d", "", "New duration (e.g. 2h, 30m, 1h30m)")
	cmd.Flags().StringVar(&date, "date", "", "New date YYYY-MM-DD")
	cmd.Flags().BoolVar(&billable, "billable", true, "Set billable (use --billable=false for non-billable)")

	return cmd
}

func runEdit(entryID int, message, duration, date string, billable *bool, messageSet bool) error {
	if !messageSet && duration == "" && date == "" && billable == nil {
		return fmt.Errorf("nothing to change. Use -m, -d, --date, or --billable")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	http := api.NewClient(cfg)
	entry, err := api.GetTimeEntry(http, cfg.BusinessID, entryID)
	if err != nil {
		return fmt.Errorf("failed to fetch entry #%d: %w", entryID, err)
	}

	req := api.CreateTimeEntryRequest{
		ClientID:  entry.ClientID,
		ProjectID: entry.ProjectID,
		ServiceID: entry.ServiceID,
		Duration:  entry.Duration,
		Note:      entry.Note,
		Billable:  entry.Billable,
		StartedAt: entry.StartedAt,
	}

	if messageSet {
		req.Note = message
	}
	if duration != "" {
		seconds, err := parseDuration(duration)
		if err != nil {
			return err
		}
		req.Duration = seconds
	}
	if date != "" {
		startedAt, err := entryStartedAt(date, time.Now())
		if err != nil {
			return err
		}
		req.StartedAt = startedAt
	}
	if billable != nil {
		req.Billable = *billable
	}

	updated, err := api.UpdateTimeEntry(http, cfg.BusinessID, entryID, req)
	if err != nil {
		return fmt.Errorf("failed to update entry #%d: %w", entryID, err)
	}

	fmt.Printf("Updated entry #%d: %.2fh %q\n", updated.ID, float64(updated.Duration)/3600, updated.Note)
	return nil
}

// DeleteCmd returns the delete command for removing a time entry.
func DeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <entry-id>",
		Short: "Delete a time entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid entry ID %q", args[0])
			}
			return runDelete(id)
		},
	}
}

func runDelete(entryID int) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	http := api.NewClient(cfg)
	if err := api.DeleteTimeEntry(http, cfg.BusinessID, entryID); err != nil {
		return fmt.Errorf("failed to delete entry #%d: %w", entryID, err)
	}

	fmt.Printf("Deleted entry #%d\n", entryID)
	return nil
}
