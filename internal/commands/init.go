package commands

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hev/freshtime/internal/api"
	"github.com/hev/freshtime/internal/config"
)

// InitCmd returns the init command.
func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize .freshtime.json in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit()
		},
	}
}

func runInit() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	http := api.NewClient(cfg)
	reader := bufio.NewReader(os.Stdin)

	// Pick client
	clients, err := api.ListClients(http, cfg.AccountID)
	if err != nil {
		return fmt.Errorf("failed to list clients: %w", err)
	}
	clientID, err := pickFromMap(reader, "Client", clients)
	if err != nil {
		return err
	}

	// Pick project
	projects, err := api.ListProjects(http, cfg.BusinessID, clientID)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}
	var projectID int
	if len(projects) > 0 {
		projectID, err = pickFromMap(reader, "Project", projects)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("No projects found for this client, skipping.")
	}

	// Pick service
	services, err := api.ListServices(http, cfg.BusinessID)
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}
	var serviceID int
	if len(services) > 0 {
		serviceID, err = pickFromMap(reader, "Service", services)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("No services found, skipping.")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	pc := &config.ProjectConfig{
		ClientID:  clientID,
		ProjectID: projectID,
		ServiceID: serviceID,
	}
	if err := config.SaveProjectConfig(cwd, pc); err != nil {
		return fmt.Errorf("failed to write %s: %w", config.ProjectConfigFile, err)
	}

	fmt.Printf("Wrote %s\n", config.ProjectConfigFile)
	return nil
}

type mapEntry struct {
	id   int
	name string
}

func pickFromMap(reader *bufio.Reader, label string, items map[int]string) (int, error) {
	var entries []mapEntry
	for id, name := range items {
		entries = append(entries, mapEntry{id, name})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	fmt.Printf("\n%s:\n", label)
	for i, e := range entries {
		fmt.Printf("  %d) %s (ID: %d)\n", i+1, e.name, e.id)
	}

	for {
		fmt.Printf("Select %s [1-%d]: ", strings.ToLower(label), len(entries))
		input, err := reader.ReadString('\n')
		if err != nil {
			return 0, err
		}
		input = strings.TrimSpace(input)
		n, err := strconv.Atoi(input)
		if err != nil || n < 1 || n > len(entries) {
			fmt.Println("Invalid selection, try again.")
			continue
		}
		selected := entries[n-1]
		fmt.Printf("Selected: %s\n", selected.name)
		return selected.id, nil
	}
}
