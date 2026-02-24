package api

import (
	"encoding/json"
	"fmt"
)

// Project represents a FreshBooks project.
type Project struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// ListProjects fetches all projects for a given client and returns a map of project ID to title.
func ListProjects(c *HttpClient, businessID, clientID int) (map[int]string, error) {
	path := fmt.Sprintf("/projects/business/%d/projects", businessID)
	raw, err := c.GetPaginated(path, "projects", map[string]string{
		"client_id": fmt.Sprintf("%d", clientID),
	})
	if err != nil {
		return nil, err
	}

	result := make(map[int]string, len(raw))
	for _, r := range raw {
		var p Project
		if err := json.Unmarshal(r, &p); err != nil {
			continue
		}
		result[p.ID] = p.Title
	}
	return result, nil
}
