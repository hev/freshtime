package api

import (
	"encoding/json"
	"fmt"
)

// Service represents a FreshBooks service.
type Service struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ListServices fetches all services for a business and returns a map of service ID to name.
func ListServices(c *HttpClient, businessID int) (map[int]string, error) {
	path := fmt.Sprintf("/comments/business/%d/services", businessID)
	raw, err := c.GetPaginated(path, "services", nil)
	if err != nil {
		return nil, err
	}

	result := make(map[int]string, len(raw))
	for _, r := range raw {
		var s Service
		if err := json.Unmarshal(r, &s); err != nil {
			continue
		}
		result[s.ID] = s.Name
	}
	return result, nil
}
