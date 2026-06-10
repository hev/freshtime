package commands

import "testing"

func TestBuildProjectRequestHourly(t *testing.T) {
	req, err := buildProjectRequest("Acme SOW", 9, "hourly", "250", "", "100h", "scope", "2026-09-30")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.ProjectType != "hourly_rate" {
		t.Errorf("expected hourly_rate, got %q", req.ProjectType)
	}
	if req.BillingMethod != "project_rate" || req.Rate != "250" {
		t.Errorf("expected project_rate at 250, got %q/%q", req.BillingMethod, req.Rate)
	}
	if req.Budget != 100*3600 {
		t.Errorf("expected budget %d, got %d", 100*3600, req.Budget)
	}
	if req.DueDate != "2026-09-30" {
		t.Errorf("expected due date, got %q", req.DueDate)
	}
}

func TestBuildProjectRequestHourlyWithoutRate(t *testing.T) {
	req, err := buildProjectRequest("Acme SOW", 9, "hourly", "", "", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.BillingMethod != "" || req.Rate != "" {
		t.Errorf("expected no billing method without a rate, got %q/%q", req.BillingMethod, req.Rate)
	}
}

func TestBuildProjectRequestFixed(t *testing.T) {
	req, err := buildProjectRequest("Acme SOW", 9, "fixed", "", "15000", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.ProjectType != "fixed_price" || req.FixedPrice != "15000" {
		t.Errorf("unexpected fixed request: %+v", req)
	}
}

func TestBuildProjectRequestValidation(t *testing.T) {
	cases := []struct {
		name                                       string
		projectType, rate, fixedPrice, budget, due string
	}{
		{"fixed without price", "fixed", "", "", "", ""},
		{"fixed with rate", "fixed", "250", "15000", "", ""},
		{"hourly with fixed price", "hourly", "", "15000", "", ""},
		{"bad type", "retainer", "", "", "", ""},
		{"bad budget", "hourly", "", "", "lots", ""},
		{"bad due date", "hourly", "", "", "", "soon"},
	}
	for _, tc := range cases {
		if _, err := buildProjectRequest("X", 9, tc.projectType, tc.rate, tc.fixedPrice, tc.budget, "", tc.due); err == nil {
			t.Errorf("%s: expected error", tc.name)
		}
	}
}
