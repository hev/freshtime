package api

import (
	"encoding/json"
	"fmt"
)

// InvoiceLine represents a line item on an invoice.
type InvoiceLine struct {
	Type        int           `json:"type"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Qty         string        `json:"qty"`
	UnitCost    InvoiceAmount `json:"unit_cost"`
}

// InvoiceAmount holds a monetary amount with currency code.
type InvoiceAmount struct {
	Amount string `json:"amount"`
	Code   string `json:"code"`
}

// CreateInvoiceRequest is the payload for creating an invoice.
type CreateInvoiceRequest struct {
	Invoice InvoicePayload `json:"invoice"`
}

// InvoicePayload contains the invoice fields.
type InvoicePayload struct {
	CustomerID int           `json:"customerid"`
	CreateDate string        `json:"create_date"`
	Lines      []InvoiceLine `json:"lines"`
	Status     int           `json:"status"`
	Notes      string        `json:"notes,omitempty"`
}

// InvoiceResponse is the API response after creating an invoice.
type InvoiceResponse struct {
	InvoiceID     int           `json:"invoiceid"`
	InvoiceNumber string        `json:"invoice_number"`
	Amount        InvoiceAmount `json:"amount"`
	V3Status      string        `json:"v3_status"`
}

type createInvoiceResp struct {
	Response struct {
		Result struct {
			Invoice InvoiceResponse `json:"invoice"`
		} `json:"result"`
	} `json:"response"`
}

type shareLinkResp struct {
	Response struct {
		Result struct {
			ShareLink string `json:"share_link"`
		} `json:"result"`
	} `json:"response"`
}

// CreateInvoice creates a new invoice in FreshBooks.
func CreateInvoice(c *HttpClient, accountID string, req *CreateInvoiceRequest) (*InvoiceResponse, error) {
	path := fmt.Sprintf("/accounting/account/%s/invoices/invoices", accountID)
	var resp createInvoiceResp
	if err := c.Post(path, req, &resp); err != nil {
		return nil, err
	}
	return &resp.Response.Result.Invoice, nil
}

// GetShareLink fetches the share link for an invoice.
func GetShareLink(c *HttpClient, accountID string, invoiceID int) (string, error) {
	path := fmt.Sprintf("/accounting/account/%s/invoices/invoices/%d/share_link", accountID, invoiceID)
	var resp shareLinkResp
	if err := c.Get(path, nil, &resp); err != nil {
		return "", err
	}
	return resp.Response.Result.ShareLink, nil
}

// Invoice is an existing invoice, as listed by the accounting API.
type Invoice struct {
	InvoiceID     int           `json:"invoiceid"`
	InvoiceNumber string        `json:"invoice_number"`
	CustomerID    int           `json:"customerid"`
	Organization  string        `json:"organization"`
	CreateDate    string        `json:"create_date"`
	DueDate       string        `json:"due_date"`
	Amount        InvoiceAmount `json:"amount"`
	Outstanding   InvoiceAmount `json:"outstanding"`
	PaymentStatus string        `json:"payment_status"`
	V3Status      string        `json:"v3_status"`
}

// ListInvoices fetches invoices, filtered by any search params given
// (e.g. "search[invoice_number]").
func ListInvoices(c *HttpClient, accountID string, params map[string]string) ([]Invoice, error) {
	path := fmt.Sprintf("/accounting/account/%s/invoices/invoices", accountID)
	raw, err := c.GetPaginated(path, "invoices", params)
	if err != nil {
		return nil, err
	}
	invoices := make([]Invoice, 0, len(raw))
	for _, r := range raw {
		var inv Invoice
		if err := json.Unmarshal(r, &inv); err != nil {
			continue
		}
		invoices = append(invoices, inv)
	}
	return invoices, nil
}

// PaymentRequest records money received against an invoice.
type PaymentRequest struct {
	InvoiceID int           `json:"invoiceid"`
	Amount    InvoiceAmount `json:"amount"`
	Date      string        `json:"date"` // YYYY-MM-DD
	Type      string        `json:"type"` // e.g. "ACH", "Bank Transfer", "Check"
	Note      string        `json:"note,omitempty"`
}

// Payment is a recorded payment.
type Payment struct {
	PaymentID int           `json:"id"`
	InvoiceID int           `json:"invoiceid"`
	Amount    InvoiceAmount `json:"amount"`
	Date      string        `json:"date"`
	Type      string        `json:"type"`
}

// CreatePayment records a payment, which marks the invoice paid (or partial).
func CreatePayment(c *HttpClient, accountID string, req PaymentRequest) (*Payment, error) {
	path := fmt.Sprintf("/accounting/account/%s/payments/payments", accountID)
	body := map[string]any{"payment": req}
	var resp struct {
		Response struct {
			Result struct {
				Payment Payment `json:"payment"`
			} `json:"result"`
		} `json:"response"`
	}
	if err := c.Post(path, body, &resp); err != nil {
		return nil, err
	}
	return &resp.Response.Result.Payment, nil
}
