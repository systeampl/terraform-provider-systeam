package client

import (
	"context"
	"fmt"
	"net/http"
)

type Organization struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

func (c *Client) GetOrganizationBySlug(ctx context.Context, slug string) (*Organization, error) {
	var org Organization
	err := c.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/organizations/by-slug/%s", slug), nil, &org)
	if err != nil {
		return nil, err
	}
	return &org, nil
}
