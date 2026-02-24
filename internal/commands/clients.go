package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hev/freshtime/internal/api"
	"github.com/hev/freshtime/internal/config"
)

// ClientsCmd returns the clients command.
func ClientsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clients",
		Short: "List clients with their IDs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClients()
		},
	}
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
