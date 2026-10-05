package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/hev/freshtime/internal/api"
	"github.com/hev/freshtime/internal/config"
)

// invoiceOpenCmd lists invoices with money still outstanding.
func invoiceOpenCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "open",
		Short: "List invoices with an outstanding balance",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInvoiceOpen(asJSON)
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func runInvoiceOpen(asJSON bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	http := api.NewClient(cfg)
	invoices, err := api.ListInvoices(http, cfg.AccountID, nil)
	if err != nil {
		return fmt.Errorf("failed to list invoices: %w", err)
	}

	open := make([]api.Invoice, 0, len(invoices))
	for _, inv := range invoices {
		if hasOutstanding(inv) && inv.V3Status != "draft" {
			open = append(open, inv)
		}
	}
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(open)
	}

	fmt.Printf("%-10s %-11s %-11s %-24s %12s %12s  %s\n", "Number", "Date", "Due", "Client", "Amount", "Outstanding", "Status")
	fmt.Println(strings.Repeat("─", 96))
	for _, inv := range open {
		org := inv.Organization
		if len(org) > 24 {
			org = org[:24]
		}
		fmt.Printf("%-10s %-11s %-11s %-24s %12s %12s  %s\n", inv.InvoiceNumber, inv.CreateDate, inv.DueDate, org,
			inv.Amount.Amount, inv.Outstanding.Amount, inv.V3Status)
	}
	if len(open) == 0 {
		fmt.Println("No open invoices.")
	}
	return nil
}

func hasOutstanding(inv api.Invoice) bool {
	v, err := strconv.ParseFloat(inv.Outstanding.Amount, 64)
	return err == nil && v > 0
}

// invoicePaidCmd records a payment against an invoice.
func invoicePaidCmd() *cobra.Command {
	var amount, date, method, note string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "paid <invoice-number>",
		Short: "Mark an invoice paid by recording a payment (default: the full outstanding balance, today, ACH)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInvoicePaid(args[0], amount, date, method, note, dryRun)
		},
	}

	cmd.Flags().StringVar(&amount, "amount", "", "Amount received (default: the outstanding balance)")
	cmd.Flags().StringVar(&date, "date", "", "Date received YYYY-MM-DD (default: today)")
	cmd.Flags().StringVar(&method, "method", "ACH", `Payment type as FreshBooks names it: "ACH", "Bank Transfer", "Check", "Credit Card", "Cash"`)
	cmd.Flags().StringVar(&note, "note", "", "Note on the payment")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be recorded without recording it")

	return cmd
}

func runInvoicePaid(number, amount, date, method, note string, dryRun bool) error {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	} else if _, err := time.Parse("2006-01-02", date); err != nil {
		return fmt.Errorf("invalid --date %q (expected YYYY-MM-DD)", date)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	http := api.NewClient(cfg)

	invoices, err := api.ListInvoices(http, cfg.AccountID, map[string]string{"search[invoice_number]": number})
	if err != nil {
		return fmt.Errorf("failed to look up invoice: %w", err)
	}
	var inv *api.Invoice
	for i := range invoices {
		if invoices[i].InvoiceNumber == number {
			inv = &invoices[i]
			break
		}
	}
	if inv == nil {
		return fmt.Errorf("no invoice numbered %q (see `freshtime invoice open`)", number)
	}
	if !hasOutstanding(*inv) {
		return fmt.Errorf("invoice %s has nothing outstanding (status %s)", inv.InvoiceNumber, inv.V3Status)
	}

	if amount == "" {
		amount = inv.Outstanding.Amount
	}
	got, err := strconv.ParseFloat(amount, 64)
	if err != nil || got <= 0 {
		return fmt.Errorf("invalid --amount %q", amount)
	}
	owed, _ := strconv.ParseFloat(inv.Outstanding.Amount, 64)
	if got > owed+0.005 {
		return fmt.Errorf("--amount %s is more than the %s outstanding on invoice %s", amount, inv.Outstanding.Amount, inv.InvoiceNumber)
	}

	fmt.Printf("Invoice %s (%s): %s %s of %s outstanding, %s on %s\n", inv.InvoiceNumber, inv.Organization,
		amount, inv.Outstanding.Code, inv.Outstanding.Amount, method, date)
	if dryRun {
		fmt.Println("Dry run: nothing recorded.")
		return nil
	}

	p, err := api.CreatePayment(http, cfg.AccountID, api.PaymentRequest{
		InvoiceID: inv.InvoiceID,
		Amount:    api.InvoiceAmount{Amount: amount, Code: inv.Outstanding.Code},
		Date:      date,
		Type:      method,
		Note:      note,
	})
	if err != nil {
		return fmt.Errorf("failed to record payment: %w", err)
	}
	status := "paid"
	if got < owed-0.005 {
		status = "partially paid"
	}
	fmt.Printf("Recorded payment #%d; invoice %s is %s.\n", p.PaymentID, inv.InvoiceNumber, status)
	return nil
}
