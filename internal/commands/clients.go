package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hev/freshtime/internal/api"
	"github.com/hev/freshtime/internal/config"
)

// ClientsCmd returns the clients command. Running it bare lists clients;
// `clients create` adds a new one.
func ClientsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clients",
		Short: "List clients with their IDs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClients()
		},
	}
	cmd.AddCommand(clientsCreateCmd())
	return cmd
}

func clientsCreateCmd() *cobra.Command {
	var org, first, last, email string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClientsCreate(org, first, last, email)
		},
	}

	cmd.Flags().StringVar(&org, "org", "", "Organization name")
	cmd.Flags().StringVar(&first, "first", "", "Contact first name")
	cmd.Flags().StringVar(&last, "last", "", "Contact last name")
	cmd.Flags().StringVar(&email, "email", "", "Contact email")

	return cmd
}

func runClientsCreate(org, first, last, email string) error {
	if org == "" && first == "" && last == "" {
		return fmt.Errorf("provide at least --org or a contact name (--first/--last)")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	http := api.NewClient(cfg)
	client, err := api.CreateClient(http, cfg.AccountID, api.CreateClientRequest{
		Organization: org,
		FName:        first,
		LName:        last,
		Email:        email,
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	name := client.Organization
	if name == "" {
		name = strings.TrimSpace(client.FName + " " + client.LName)
	}
	fmt.Printf("Created client %q (client #%d)\n", name, client.ID)
	return nil
}

func runClients() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	http := api.NewClient(cfg)
	clients, err := api.ListClients(http, cfg.AccountID)
	if err != nil {
		return err
	}

	const idWidth = 8
	fmt.Printf("%-*s%s\n", idWidth, "ID", "Name")
	fmt.Println(strings.Repeat("─", 40))

	if len(clients) == 0 {
		fmt.Println("No clients found.")
		return nil
	}

	// Sort by name for consistent output
	type entry struct {
		id   int
		name string
	}
	var sorted []entry
	for id, name := range clients {
		sorted = append(sorted, entry{id, name})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].name < sorted[j].name
	})

	for _, e := range sorted {
		fmt.Printf("%-*d%s\n", idWidth, e.id, e.name)
	}
	return nil
}
