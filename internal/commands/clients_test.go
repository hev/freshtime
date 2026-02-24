package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hev/freshtime/internal/api"
)

func TestListClientsFormatting(t *testing.T) {
	t.Run("formats clients as a map", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]any{
				"response": map[string]any{
					"result": map[string]any{
						"clients": []map[string]any{
							{"id": 123, "organization": "Acme Corp", "fname": "", "lname": ""},
							{"id": 456, "organization": "Widget Inc", "fname": "", "lname": ""},
						},
						"pages": 1,
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		origBase := api.BaseURL
		api.BaseURL = srv.URL
		defer func() { api.BaseURL = origBase }()

		c := api.NewHttpClient("test-token")
		clients, err := api.ListClients(c, "abc123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(clients) != 2 {
			t.Fatalf("expected 2 clients, got %d", len(clients))
		}
		if clients[123] != "Acme Corp" {
			t.Errorf("client 123 = %q, want %q", clients[123], "Acme Corp")
		}
		if clients[456] != "Widget Inc" {
			t.Errorf("client 456 = %q, want %q", clients[456], "Widget Inc")
		}
	})

	t.Run("empty client list", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]any{
				"response": map[string]any{
					"result": map[string]any{
						"clients": []map[string]any{},
						"pages":   1,
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		origBase := api.BaseURL
		api.BaseURL = srv.URL
		defer func() { api.BaseURL = origBase }()

		c := api.NewHttpClient("test-token")
		clients, err := api.ListClients(c, "abc123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(clients) != 0 {
			t.Errorf("expected 0 clients, got %d", len(clients))
		}
	})

	t.Run("uses fname/lname when organization is empty", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]any{
				"response": map[string]any{
					"result": map[string]any{
						"clients": []map[string]any{
							{"id": 789, "organization": "", "fname": "John", "lname": "Doe"},
						},
						"pages": 1,
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		origBase := api.BaseURL
		api.BaseURL = srv.URL
		defer func() { api.BaseURL = origBase }()

		c := api.NewHttpClient("test-token")
		clients, err := api.ListClients(c, "abc123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		name := clients[789]
		if !strings.Contains(name, "John") || !strings.Contains(name, "Doe") {
			t.Errorf("client 789 = %q, want name containing John Doe", name)
		}
	})
}
