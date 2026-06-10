package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ClientRecord represents a FreshBooks client.
type ClientRecord struct {
	ID           int    `json:"id"`
	Organization string `json:"organization"`
	FName        string `json:"fname"`
	LName        string `json:"lname"`
}

// CreateClientRequest holds the parameters for creating a client.
type CreateClientRequest struct {
	Organization string `json:"organization,omitempty"`
	FName        string `json:"fname,omitempty"`
	LName        string `json:"lname,omitempty"`
	Email        string `json:"email,omitempty"`
}

// CreateClient creates a new client via the FreshBooks accounting API.
func CreateClient(c *HttpClient, accountID string, req CreateClientRequest) (*ClientRecord, error) {
	path := fmt.Sprintf("/accounting/account/%s/users/clients", accountID)
	body := map[string]any{"client": req}
	var resp struct {
		Response struct {
			Result struct {
				Client ClientRecord `json:"client"`
			} `json:"result"`
		} `json:"response"`
	}
	if err := c.Post(path, body, &resp); err != nil {
		return nil, err
	}
	return &resp.Response.Result.Client, nil
}

// ListClients fetches all clients and returns a map of client ID to display name.
func ListClients(c *HttpClient, accountID string) (map[int]string, error) {
	path := fmt.Sprintf("/accounting/account/%s/users/clients", accountID)
	raw, err := c.GetPaginated(path, "clients", nil)
	if err != nil {
		return nil, err
	}

	result := make(map[int]string, len(raw))
	for _, r := range raw {
		var cr ClientRecord
		if err := json.Unmarshal(r, &cr); err != nil {
			continue
		}
		name := cr.Organization
		if name == "" {
			name = strings.TrimSpace(cr.FName + " " + cr.LName)
		}
		if name == "" {
			name = fmt.Sprintf("Client #%d", cr.ID)
		}
		result[cr.ID] = name
	}
	return result, nil
}
