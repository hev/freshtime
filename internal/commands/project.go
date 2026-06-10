package commands

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/hev/freshtime/internal/api"
	"github.com/hev/freshtime/internal/config"
)

// ProjectCmd returns the project command group.
func ProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage FreshBooks projects",
	}
	cmd.AddCommand(projectCreateCmd())
	cmd.AddCommand(projectListCmd())
	return cmd
}

func projectCreateCmd() *cobra.Command {
	var (
		name        string
		client      string
		projectType string
		rate        string
		fixedPrice  string
		budget      string
		description string
		due         string
		writeInit   bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project for a client (e.g. from a signed SOW)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectCreate(name, client, projectType, rate, fixedPrice, budget, description, due, writeInit)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Project name (required)")
	cmd.Flags().StringVar(&client, "client", "", "Client name (substring, case-insensitive) or numeric ID (required)")
	cmd.Flags().StringVar(&projectType, "type", "hourly", "Billing type: hourly or fixed")
	cmd.Flags().StringVar(&rate, "rate", "", "Hourly rate for hourly projects (e.g. 250)")
	cmd.Flags().StringVar(&fixedPrice, "fixed-price", "", "Total price for fixed projects (e.g. 15000)")
	cmd.Flags().StringVar(&budget, "budget", "", "Budgeted hours (e.g. 100h)")
	cmd.Flags().StringVar(&description, "description", "", "Project description (e.g. SOW scope summary)")
	cmd.Flags().StringVar(&due, "due", "", "Due date YYYY-MM-DD")
	cmd.Flags().BoolVar(&writeInit, "init", false, "Write .freshtime.json in the current directory pointing at the new project")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("client")

	return cmd
}

// buildProjectRequest validates flags and assembles the API request.
func buildProjectRequest(name string, clientID int, projectType, rate, fixedPrice, budget, description, due string) (api.CreateProjectRequest, error) {
	req := api.CreateProjectRequest{
		Title:       name,
		ClientID:    clientID,
		Description: description,
	}

	switch projectType {
	case "hourly":
		req.ProjectType = "hourly_rate"
		if fixedPrice != "" {
			return req, fmt.Errorf("--fixed-price only applies to --type fixed")
		}
		if rate != "" {
			req.BillingMethod = "project_rate"
			req.Rate = rate
		}
	case "fixed":
		req.ProjectType = "fixed_price"
		if rate != "" {
			return req, fmt.Errorf("--rate only applies to --type hourly")
		}
		if fixedPrice == "" {
			return req, fmt.Errorf("--type fixed requires --fixed-price")
		}
		req.FixedPrice = fixedPrice
	default:
		return req, fmt.Errorf("invalid --type %q (expected hourly or fixed)", projectType)
	}

	if budget != "" {
		seconds, err := parseDuration(budget)
		if err != nil {
			return req, fmt.Errorf("invalid --budget: %w", err)
		}
		req.Budget = seconds
	}

	if due != "" {
		if _, err := time.Parse("2006-01-02", due); err != nil {
			return req, fmt.Errorf("invalid --due %q (expected YYYY-MM-DD)", due)
		}
		req.DueDate = due
	}

	return req, nil
}

func runProjectCreate(name, client, projectType, rate, fixedPrice, budget, description, due string, writeInit bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	http := api.NewClient(cfg)
	clientNames, err := api.ListClients(http, cfg.AccountID)
	if err != nil {
		return fmt.Errorf("failed to list clients: %w", err)
	}
	clientID, clientName, err := resolveClient(clientNames, client)
	if err != nil {
		return err
	}
	if clientID == 0 {
		return fmt.Errorf("--client is required")
	}

	req, err := buildProjectRequest(name, clientID, projectType, rate, fixedPrice, budget, description, due)
	if err != nil {
		return err
	}

	project, err := api.CreateProject(http, cfg.BusinessID, req)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	fmt.Printf("Created project %q for %s (project #%d)\n", project.Title, clientName, project.ID)

	if writeInit {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		pc, _ := config.LoadProjectConfig(cwd)
		if pc == nil {
			pc = &config.ProjectConfig{}
		}
		pc.ClientID = clientID
		pc.ProjectID = project.ID
		if err := config.SaveProjectConfig(cwd, pc); err != nil {
			return fmt.Errorf("failed to write %s: %w", config.ProjectConfigFile, err)
		}
		fmt.Printf("Wrote %s (client %d, project %d)\n", config.ProjectConfigFile, clientID, project.ID)
	}

	return nil
}

func projectListCmd() *cobra.Command {
	var client string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects, optionally filtered by client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectList(client)
		},
	}

	cmd.Flags().StringVar(&client, "client", "", "Filter by client name (substring, case-insensitive) or numeric ID")

	return cmd
}

func runProjectList(client string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	http := api.NewClient(cfg)
	clientID := 0
	if client != "" {
		clientNames, err := api.ListClients(http, cfg.AccountID)
		if err != nil {
			return fmt.Errorf("failed to list clients: %w", err)
		}
		clientID, _, err = resolveClient(clientNames, client)
		if err != nil {
			return err
		}
	}

	projects, err := api.ListProjects(http, cfg.BusinessID, clientID)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return nil
	}

	type entry struct {
		id   int
		name string
	}
	var sorted []entry
	for id, name := range projects {
		sorted = append(sorted, entry{id, name})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].name < sorted[j].name
	})

	const idWidth = 10
	fmt.Printf("%-*s%s\n", idWidth, "ID", "Title")
	fmt.Println(strings.Repeat("─", 40))
	for _, e := range sorted {
		fmt.Printf("%-*d%s\n", idWidth, e.id, e.name)
	}
	return nil
}
