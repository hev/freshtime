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
