package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/hev/freshtime/internal/commands"
)

func main() {
	root := &cobra.Command{
		Use:     "freshtime",
		Short:   "FreshBooks weekly time summary CLI",
		Version: "1.0.0",
	}

	root.AddCommand(commands.SetupCmd())
	root.AddCommand(commands.WeeklyCmd())
	root.AddCommand(commands.ClientsCmd())
	root.AddCommand(commands.InvoiceCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
