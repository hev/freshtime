package api

import (
	"encoding/json"
	"fmt"
)

// Project represents a FreshBooks project.
type Project struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	ClientID int    `json:"client_id"`
}

// CreateProjectRequest holds the parameters for creating a project.
type CreateProjectRequest struct {
	Title         string `json:"title"`
	ClientID      int    `json:"client_id"`
	ProjectType   string `json:"project_type"`             // "hourly_rate" or "fixed_price"
	BillingMethod string `json:"billing_method,omitempty"` // e.g. "project_rate" for hourly projects
	Rate          string `json:"rate,omitempty"`           // hourly rate, decimal string
	FixedPrice    string `json:"fixed_price,omitempty"`    // total price for fixed_price projects
	Budget        int    `json:"budget,omitempty"`         // seconds
	Description   string `json:"description,omitempty"`
	DueDate       string `json:"due_date,omitempty"` // YYYY-MM-DD
}

// CreateProject creates a new project via the FreshBooks API.
func CreateProject(c *HttpClient, businessID int, req CreateProjectRequest) (*Project, error) {
	path := fmt.Sprintf("/projects/business/%d/projects", businessID)
	body := map[string]any{"project": req}
	var resp struct {
		Project Project `json:"project"`
	}
	if err := c.Post(path, body, &resp); err != nil {
		return nil, err
	}
	return &resp.Project, nil
}

// ListProjects fetches all projects, filtered to a client when clientID is
// non-zero, and returns a map of project ID to title.
func ListProjects(c *HttpClient, businessID, clientID int) (map[int]string, error) {
	path := fmt.Sprintf("/projects/business/%d/projects", businessID)
	params := map[string]string{}
	if clientID != 0 {
		params["client_id"] = fmt.Sprintf("%d", clientID)
	}
	raw, err := c.GetPaginated(path, "projects", params)
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
