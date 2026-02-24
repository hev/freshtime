package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hev/freshtime/internal/api"
	"github.com/hev/freshtime/internal/config"
)

// SetupCmd returns the setup command.
func SetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Configure your FreshBooks access token",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup()
		},
	}
}

func runSetup() error {
	fmt.Print("Enter your FreshBooks access token: ")
	var token string
	if _, err := fmt.Scanln(&token); err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}

	fmt.Println("Verifying token...")
	http := api.NewHttpClient(token)
	identity, err := api.GetIdentity(http)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	cfg := &config.Config{
		AccessToken: token,
		AccountID:   identity.AccountID,
		BusinessID:  identity.BusinessID,
	}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println("Setup complete.")
	fmt.Printf("  Account:  %s\n", identity.AccountID)
	fmt.Printf("  Business: %d\n", identity.BusinessID)
	fmt.Printf("  Config:   %s\n", config.Path())
	return nil
}
