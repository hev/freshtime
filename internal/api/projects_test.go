package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/projects/business/42/projects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body struct {
			Project CreateProjectRequest `json:"project"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if body.Project.Title != "SOW: Acme AI Transformation" {
			t.Errorf("unexpected title: %q", body.Project.Title)
		}
		if body.Project.ProjectType != "hourly_rate" {
			t.Errorf("unexpected project_type: %q", body.Project.ProjectType)
		}
		if body.Project.Budget != 360000 {
			t.Errorf("unexpected budget: %d", body.Project.Budget)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"project": map[string]any{"id": 777, "title": body.Project.Title, "client_id": body.Project.ClientID},
		})
	}))
	defer srv.Close()

	origBase := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = origBase }()

	c := NewHttpClient("test-token")
	project, err := CreateProject(c, 42, CreateProjectRequest{
		Title:         "SOW: Acme AI Transformation",
		ClientID:      9,
		ProjectType:   "hourly_rate",
		BillingMethod: "project_rate",
		Rate:          "250",
		Budget:        360000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.ID != 777 {
		t.Errorf("expected project ID 777, got %d", project.ID)
	}
	if project.ClientID != 9 {
		t.Errorf("expected client ID 9, got %d", project.ClientID)
	}
}

func TestCreateClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/accounting/account/abc123/users/clients" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body struct {
			Client CreateClientRequest `json:"client"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if body.Client.Organization != "Acme Inc" {
			t.Errorf("unexpected organization: %q", body.Client.Organization)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"response": map[string]any{
				"result": map[string]any{
					"client": map[string]any{"id": 55, "organization": "Acme Inc"},
				},
			},
		})
	}))
	defer srv.Close()

	origBase := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = origBase }()

	c := NewHttpClient("test-token")
	client, err := CreateClient(c, "abc123", CreateClientRequest{Organization: "Acme Inc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.ID != 55 {
		t.Errorf("expected client ID 55, got %d", client.ID)
	}
}

func TestListProjectsOmitsZeroClientFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("client_id") {
			t.Errorf("expected no client_id param, got %q", r.URL.Query().Get("client_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"projects": []map[string]any{{"id": 1, "title": "P1"}},
			"meta":     map[string]int{"pages": 1},
		})
	}))
	defer srv.Close()

	origBase := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = origBase }()

	c := NewHttpClient("test-token")
	projects, err := ListProjects(c, 42, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if projects[1] != "P1" {
		t.Errorf("expected project 1 = P1, got %v", projects)
	}
}
