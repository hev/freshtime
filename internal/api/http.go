package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const BaseURL = "https://api.freshbooks.com"

// ApiError represents a non-2xx response from the FreshBooks API.
type ApiError struct {
	Status     int
	StatusText string
	Body       string
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("API error %d %s: %s", e.Status, e.StatusText, e.Body)
}

// AuthError represents a 401 Unauthorized response.
type AuthError struct {
	ApiError
}

// HttpClient wraps authenticated requests to the FreshBooks API.
type HttpClient struct {
	token     string
	client    *http.Client
	onRefresh func() (string, error) // returns new token
	retried   bool
}

// NewHttpClient creates an HttpClient with the given bearer token.
func NewHttpClient(token string) *HttpClient {
	return &HttpClient{
		token:  token,
		client: &http.Client{},
	}
}

// SetRefreshFunc sets a callback used to refresh the access token on 401.
func (c *HttpClient) SetRefreshFunc(fn func() (string, error)) {
	c.onRefresh = fn
}

// Get performs an authenticated GET request and decodes the JSON response into dest.
func (c *HttpClient) Get(path string, params map[string]string, dest any) error {
	u, err := url.Parse(BaseURL + path)
	if err != nil {
		return err
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	return c.doJSON(req, dest)
}

// Post performs an authenticated POST request.
func (c *HttpClient) Post(path string, body any, dest any) error {
	return c.mutate("POST", path, body, dest)
}

// Put performs an authenticated PUT request.
func (c *HttpClient) Put(path string, body any, dest any) error {
	return c.mutate("PUT", path, body, dest)
}

func (c *HttpClient) mutate(method, path string, body any, dest any) error {
	u := BaseURL + path
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, u, bytes.NewReader(data))
	if err != nil {
		return err
	}
	return c.doJSON(req, dest)
}

func (c *HttpClient) doJSON(req *http.Request, dest any) error {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == 401 && !c.retried && c.onRefresh != nil {
		c.retried = true
		newToken, refreshErr := c.onRefresh()
		if refreshErr != nil {
			return &AuthError{ApiError{401, "Unauthorized", "Session expired. Run `freshtime setup` to re-authenticate."}}
		}
		c.token = newToken
		return c.doJSON(req, dest)
	}

	if resp.StatusCode == 401 {
		return &AuthError{ApiError{401, "Unauthorized", string(respBody)}}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &ApiError{resp.StatusCode, resp.Status, string(respBody)}
	}

	if dest != nil {
		return json.Unmarshal(respBody, dest)
	}
	return nil
}

// GetPaginated fetches all pages for a paginated endpoint.
// resultKey is the JSON key containing the array of results.
func (c *HttpClient) GetPaginated(path, resultKey string, params map[string]string) ([]json.RawMessage, error) {
	var allResults []json.RawMessage
	page := 1
	totalPages := 1

	for page <= totalPages {
		p := make(map[string]string)
		for k, v := range params {
			p[k] = v
		}
		p["page"] = fmt.Sprintf("%d", page)
		p["per_page"] = "100"

		var raw map[string]json.RawMessage
		if err := c.Get(path, p, &raw); err != nil {
			return nil, err
		}

		items, pages := extractPage(raw, resultKey)
		allResults = append(allResults, items...)
		if pages > 0 {
			totalPages = pages
		}
		page++
	}

	return allResults, nil
}

// extractPage handles the two FreshBooks response shapes:
// - Timetracking: { [key]: [...], meta: { pages: N } }
// - Accounting:   { response: { result: { [key]: [...], pages: N } } }
func extractPage(data map[string]json.RawMessage, key string) ([]json.RawMessage, int) {
	// Try top-level (timetracking)
	if itemsRaw, ok := data[key]; ok {
		var items []json.RawMessage
		json.Unmarshal(itemsRaw, &items)

		pages := 1
		if metaRaw, ok := data["meta"]; ok {
			var meta struct {
				Pages int `json:"pages"`
			}
			json.Unmarshal(metaRaw, &meta)
			pages = meta.Pages
		}
		return items, pages
	}

	// Try nested (accounting)
	if respRaw, ok := data["response"]; ok {
		var resp map[string]json.RawMessage
		json.Unmarshal(respRaw, &resp)

		if resultRaw, ok := resp["result"]; ok {
			var result map[string]json.RawMessage
			json.Unmarshal(resultRaw, &result)

			if itemsRaw, ok := result[key]; ok {
				var items []json.RawMessage
				json.Unmarshal(itemsRaw, &items)

				pages := 1
				if pagesRaw, ok := result["pages"]; ok {
					json.Unmarshal(pagesRaw, &pages)
				}
				return items, pages
			}
		}
	}

	return nil, 1
}
