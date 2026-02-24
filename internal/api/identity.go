package api

import "fmt"

// Identity holds the account and business IDs for a FreshBooks user.
type Identity struct {
	AccountID  string
	BusinessID int
}

type meResponse struct {
	Response struct {
		ID                  int `json:"id"`
		BusinessMemberships []struct {
			Business struct {
				ID        int    `json:"id"`
				AccountID string `json:"account_id"`
			} `json:"business"`
		} `json:"business_memberships"`
	} `json:"response"`
}

// GetIdentity fetches the current user's identity from the FreshBooks API.
func GetIdentity(c *HttpClient) (*Identity, error) {
	var data meResponse
	if err := c.Get("/auth/api/v1/users/me", nil, &data); err != nil {
		return nil, err
	}

	memberships := data.Response.BusinessMemberships
	if len(memberships) == 0 {
		return nil, fmt.Errorf("no business memberships found on this account")
	}

	first := memberships[0]
	return &Identity{
		AccountID:  first.Business.AccountID,
		BusinessID: first.Business.ID,
	}, nil
}
