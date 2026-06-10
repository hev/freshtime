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
		Short:   "FreshBooks time tracking CLI",
		Version: "1.1.0",
	}

	root.AddCommand(commands.SetupCmd())
	root.AddCommand(commands.WeeklyCmd())
	root.AddCommand(commands.DailyCmd())
	root.AddCommand(commands.ClientsCmd())
	root.AddCommand(commands.ProjectCmd())
	root.AddCommand(commands.InvoiceCmd())
	root.AddCommand(commands.InitCmd())
	root.AddCommand(commands.LogCmd())
	root.AddCommand(commands.ListCmd())
	root.AddCommand(commands.EditCmd())
	root.AddCommand(commands.DeleteCmd())
	root.AddCommand(commands.StartCmd())
	root.AddCommand(commands.StopCmd())
	root.AddCommand(commands.TimerStatusCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
