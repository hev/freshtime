package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/hev/freshtime/internal/api"
	"github.com/hev/freshtime/internal/config"
)

// InvoiceCmd returns the invoice command.
func InvoiceCmd() *cobra.Command {
	var rate string
	var currency string
	var dryRun bool
	var notes string

	cmd := &cobra.Command{
		Use:   "invoice <client-id>",
		Short: "Create an invoice for all unbilled time entries for a client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientID, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid client ID: %w", err)
			}
			return runInvoice(clientID, rate, currency, dryRun, notes)
		},
	}

	cmd.Flags().StringVar(&rate, "rate", "", "Override the hourly rate for this run")
	cmd.Flags().StringVar(&currency, "currency", "", "Override currency code (default: config or USD)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be invoiced without creating it")
	cmd.Flags().StringVar(&notes, "notes", "", "Add notes to the invoice")

	return cmd
}

func buildInvoiceLines(entries []api.TimeEntry, rate, currency string) []api.InvoiceLine {
	lines := make([]api.InvoiceLine, 0, len(entries))
	for _, entry := range entries {
		name := entry.Note
		if name == "" {
			name = "Consulting"
		}
		desc := ""
		if parts := splitDateTime(entry.LocalStartedAt); parts != "" {
			desc = parts
		}
		lines = append(lines, api.InvoiceLine{
			Type:        0,
			Name:        name,
			Description: desc,
			Qty:         fmt.Sprintf("%.2f", float64(entry.Duration)/3600),
			UnitCost:    api.InvoiceAmount{Amount: rate, Code: currency},
		})
	}
	return lines
}

func splitDateTime(dt string) string {
	if len(dt) >= 10 {
		return dt[:10]
	}
	return dt
}

func runInvoice(clientID int, rate, currency string, dryRun bool, notes string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	http := api.NewHttpClient(cfg.AccessToken)

	entries, err := api.ListUnbilledEntries(http, cfg.BusinessID, clientID)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No unbilled time entries found for this client.")
		return nil
	}

	// Resolve rate
	if rate == "" {
		rate = cfg.ClientRates[strconv.Itoa(clientID)]
	}
	if rate == "" {
		return fmt.Errorf("no rate configured for client %d. Use --rate <amount> or set client_rates.%d in config", clientID, clientID)
	}

	// Resolve currency
	if currency == "" {
		currency = cfg.DefaultCurrency
	}
	if currency == "" {
		currency = "USD"
	}

	lines := buildInvoiceLines(entries, rate, currency)

	var totalSeconds int
	for _, e := range entries {
		totalSeconds += e.Duration
	}
	totalHours := float64(totalSeconds) / 3600
	rateFloat, _ := strconv.ParseFloat(rate, 64)
	totalAmount := totalHours * rateFloat

	if dryRun {
		fmt.Println("Dry run — no invoice created.")
		fmt.Println()
		fmt.Printf("Entries:  %d\n", len(entries))
		fmt.Printf("Hours:   %.2f\n", totalHours)
		fmt.Printf("Rate:    %s %s/hr\n", rate, currency)
		fmt.Printf("Total:   %.2f %s\n\n", totalAmount, currency)
		fmt.Println("Line items:")
		for _, line := range lines {
			fmt.Printf("  %s  %sh  %s\n", line.Description, line.Qty, line.Name)
		}
		return nil
	}

	today := time.Now().Format("2006-01-02")
	req := &api.CreateInvoiceRequest{
		Invoice: api.InvoicePayload{
			CustomerID: clientID,
			CreateDate: today,
			Lines:      lines,
			Status:     1,
			Notes:      notes,
		},
	}

	invoice, err := api.CreateInvoice(http, cfg.AccountID, req)
	if err != nil {
		return err
	}

	fmt.Printf("Invoice #%s created (draft).\n", invoice.InvoiceNumber)
	fmt.Printf("ID:      %d\n", invoice.InvoiceID)
	fmt.Printf("Entries: %d\n", len(entries))
	fmt.Printf("Hours:   %.2f\n", totalHours)
	fmt.Printf("Total:   %s %s\n", invoice.Amount.Amount, invoice.Amount.Code)

	shareLink, err := api.GetShareLink(http, cfg.AccountID, invoice.InvoiceID)
	if err != nil || shareLink == "" {
		fmt.Println("Link:    (share link unavailable — may need invoices:read scope)")
	} else {
		fmt.Printf("Link:    %s\n", shareLink)
	}

	if err := api.MarkEntriesAsBilled(http, cfg.BusinessID, entries); err != nil {
		fmt.Printf("Warning: Failed to mark entries as billed — %s\n", err)
	} else {
		fmt.Printf("Billed:  %d entries marked as billed\n", len(entries))
	}

	return nil
}
