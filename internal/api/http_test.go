package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		if r.URL.Query().Get("foo") != "bar" {
			t.Errorf("expected query param foo=bar, got %s", r.URL.Query().Get("foo"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"hello": "world"})
	}))
	defer srv.Close()

	// Override BaseURL for test
	origBase := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = origBase }()

	c := NewHttpClient("test-token")
	var result map[string]string
	err := c.Get("/test", map[string]string{"foo": "bar"}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["hello"] != "world" {
		t.Errorf("expected hello=world, got %v", result)
	}
}

func TestPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["key"] != "value" {
			t.Errorf("expected key=value in body, got %v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "123"})
	}))
	defer srv.Close()

	origBase := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = origBase }()

	c := NewHttpClient("test-token")
	var result map[string]string
	err := c.Post("/create", map[string]string{"key": "value"}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["id"] != "123" {
		t.Errorf("expected id=123, got %v", result)
	}
}

func TestPut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"updated": "true"})
	}))
	defer srv.Close()

	origBase := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = origBase }()

	c := NewHttpClient("test-token")
	var result map[string]string
	err := c.Put("/update", map[string]string{"key": "value"}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["updated"] != "true" {
		t.Errorf("expected updated=true, got %v", result)
	}
}

func TestApiError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	origBase := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = origBase }()

	c := NewHttpClient("test-token")
	var result map[string]string
	err := c.Get("/fail", nil, &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*ApiError)
	if !ok {
		t.Fatalf("expected *ApiError, got %T", err)
	}
	if apiErr.Status != 500 {
		t.Errorf("expected status 500, got %d", apiErr.Status)
	}
}

func TestAuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte("unauthorized"))
	}))
	defer srv.Close()

	origBase := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = origBase }()

	c := NewHttpClient("bad-token")
	var result map[string]string
	err := c.Get("/auth", nil, &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	_, ok := err.(*AuthError)
	if !ok {
		t.Fatalf("expected *AuthError, got %T", err)
	}
}

func TestTokenRefreshOn401(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.Header.Get("Authorization") == "Bearer old-token" {
			w.WriteHeader(401)
			w.Write([]byte("expired"))
			return
		}
		if r.Header.Get("Authorization") == "Bearer new-token" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
			return
		}
		w.WriteHeader(401)
		w.Write([]byte("unknown token"))
	}))
	defer srv.Close()

	origBase := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = origBase }()

	c := NewHttpClient("old-token")
	c.SetRefreshFunc(func() (string, error) {
		return "new-token", nil
	})

	var result map[string]string
	err := c.Get("/protected", nil, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["ok"] != "true" {
		t.Errorf("expected ok=true, got %v", result)
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 calls (retry), got %d", callCount)
	}
}

func TestGetPaginatedTimetracking(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		resp := map[string]any{
			"entries": []map[string]string{{"id": "1"}},
			"meta":    map[string]int{"pages": 2},
		}
		if page >= 2 {
			resp["meta"] = map[string]int{"pages": 2}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	origBase := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = origBase }()

	c := NewHttpClient("test-token")
	results, err := c.GetPaginated("/entries", "entries", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results (2 pages), got %d", len(results))
	}
}

func TestGetPaginatedAccounting(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"response": map[string]any{
				"result": map[string]any{
					"clients": []map[string]string{{"id": "1"}, {"id": "2"}},
					"pages":   1,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	origBase := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = origBase }()

	c := NewHttpClient("test-token")
	results, err := c.GetPaginated("/clients", "clients", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestExtractPageTimetracking(t *testing.T) {
	data := map[string]json.RawMessage{
		"entries": json.RawMessage(`[{"id":"1"},{"id":"2"}]`),
		"meta":    json.RawMessage(`{"pages":3}`),
	}
	items, pages := extractPage(data, "entries")
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	if pages != 3 {
		t.Errorf("expected 3 pages, got %d", pages)
	}
}

func TestExtractPageAccounting(t *testing.T) {
	data := map[string]json.RawMessage{
		"response": json.RawMessage(`{"result":{"invoices":[{"id":"1"}],"pages":5}}`),
	}
	items, pages := extractPage(data, "invoices")
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
	if pages != 5 {
		t.Errorf("expected 5 pages, got %d", pages)
	}
}
