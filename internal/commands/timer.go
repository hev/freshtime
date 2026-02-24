package commands

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/hev/freshtime/internal/api"
	"github.com/hev/freshtime/internal/config"
)

// TimerState persists a running timer.
type TimerState struct {
	StartedAt time.Time `json:"started_at"`
	Note      string    `json:"note"`
	ClientID  int       `json:"client_id"`
	ProjectID int       `json:"project_id,omitempty"`
	ServiceID int       `json:"service_id,omitempty"`
	Billable  bool      `json:"billable"`
}

func timerPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "freshtime", "timer.json")
}

func loadTimer() (*TimerState, error) {
	data, err := os.ReadFile(timerPath())
	if err != nil {
		return nil, fmt.Errorf("no timer running")
	}
	var ts TimerState
	if err := json.Unmarshal(data, &ts); err != nil {
		return nil, fmt.Errorf("corrupt timer state: %w", err)
	}
	return &ts, nil
}

func saveTimer(ts *TimerState) error {
	data, err := json.MarshalIndent(ts, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(timerPath(), data, 0o644)
}

func clearTimer() error {
	return os.Remove(timerPath())
}

// StartCmd returns the start command.
func StartCmd() *cobra.Command {
	var (
		message    string
		client     int
		project    int
		service    int
		noBillable bool
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a time tracking timer",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(message, client, project, service, noBillable)
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Note for the time entry")
	cmd.Flags().IntVar(&client, "client", 0, "Client ID (overrides .freshtime.json)")
	cmd.Flags().IntVar(&project, "project", 0, "Project ID (overrides .freshtime.json)")
	cmd.Flags().IntVar(&service, "service", 0, "Service ID (overrides .freshtime.json)")
	cmd.Flags().BoolVar(&noBillable, "no-billable", false, "Mark as non-billable")

	return cmd
}

// StopCmd returns the stop command.
func StopCmd() *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the running timer and log the time entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop(message)
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Override the note set at start")

	return cmd
}

// StatusCmd returns the status command for checking timer state.
func TimerStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the current timer status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTimerStatus()
		},
	}
}

func runStart(message string, clientID, projectID, serviceID int, noBillable bool) error {
	// Check for existing timer
	if existing, _ := loadTimer(); existing != nil {
		elapsed := time.Since(existing.StartedAt)
		return fmt.Errorf("timer already running (started %s ago, note: %q). Run `freshtime stop` first",
			formatElapsed(elapsed), existing.Note)
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

	ts := &TimerState{
		StartedAt: time.Now(),
		Note:      message,
		ClientID:  clientID,
		ProjectID: projectID,
		ServiceID: serviceID,
		Billable:  !noBillable,
	}
	if err := saveTimer(ts); err != nil {
		return fmt.Errorf("failed to save timer: %w", err)
	}

	fmt.Printf("Timer started")
	if message != "" {
		fmt.Printf(": %s", message)
	}
	fmt.Println()
	return nil
}

func runStop(messageOverride string) error {
	ts, err := loadTimer()
	if err != nil {
		return err
	}

	elapsed := time.Since(ts.StartedAt)
	seconds := int(math.Round(elapsed.Seconds()))
	if seconds < 60 {
		seconds = 60 // minimum 1 minute
	}

	note := ts.Note
	if messageOverride != "" {
		note = messageOverride
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	http := api.NewClient(cfg)
	entry, err := api.CreateTimeEntry(http, cfg.BusinessID, api.CreateTimeEntryRequest{
		ClientID:  ts.ClientID,
		ProjectID: ts.ProjectID,
		ServiceID: ts.ServiceID,
		Duration:  seconds,
		Note:      note,
		Billable:  ts.Billable,
		StartedAt: ts.StartedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
	if err != nil {
		return fmt.Errorf("failed to create time entry: %w", err)
	}

	if err := clearTimer(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to clear timer state: %v\n", err)
	}

	hours := float64(seconds) / 3600
	fmt.Printf("Stopped. Logged %.2fh: %s (entry #%d)\n", hours, note, entry.ID)
	return nil
}

func runTimerStatus() error {
	ts, err := loadTimer()
	if err != nil {
		fmt.Println("No timer running.")
		return nil
	}

	elapsed := time.Since(ts.StartedAt)
	fmt.Printf("Timer running: %s\n", formatElapsed(elapsed))
	if ts.Note != "" {
		fmt.Printf("Note: %s\n", ts.Note)
	}
	fmt.Printf("Client: %d\n", ts.ClientID)
	if ts.ProjectID != 0 {
		fmt.Printf("Project: %d\n", ts.ProjectID)
	}
	return nil
}

func formatElapsed(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
