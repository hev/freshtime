package api

import "fmt"

// InvoiceLine represents a line item on an invoice.
type InvoiceLine struct {
	Type        int            `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Qty         string         `json:"qty"`
	UnitCost    InvoiceAmount  `json:"unit_cost"`
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
