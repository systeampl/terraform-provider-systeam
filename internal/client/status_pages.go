package client

import (
	"context"
	"fmt"
	"net/http"
)

type StatusPage struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Description    string `json:"description"`
	IsPublic       bool   `json:"is_public"`
	CustomDomain   string `json:"custom_domain"`
	LogoURL        string `json:"logo_url"`
	CheckIDs       []int  `json:"check_ids"`
	OrganizationID int    `json:"organization_id"`
	IsActive       bool   `json:"is_active"`
}

type StatusPageCreateRequest struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Description  string `json:"description,omitempty"`
	IsPublic     bool   `json:"is_public"`
	CustomDomain string `json:"custom_domain,omitempty"`
	LogoURL      string `json:"logo_url,omitempty"`
	CheckIDs     []int  `json:"check_ids"`
}

type StatusPageUpdateRequest struct {
	Name         *string `json:"name,omitempty"`
	Slug         *string `json:"slug,omitempty"`
	Description  *string `json:"description,omitempty"`
	IsPublic     *bool   `json:"is_public,omitempty"`
	CustomDomain *string `json:"custom_domain,omitempty"`
	LogoURL      *string `json:"logo_url,omitempty"`
	CheckIDs     []int   `json:"check_ids,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

func (c *Client) CreateStatusPage(ctx context.Context, req StatusPageCreateRequest) (*StatusPage, error) {
	var page StatusPage
	err := c.DoRequest(ctx, http.MethodPost, "/api/status-pages", req, &page)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func (c *Client) GetStatusPage(ctx context.Context, pageID int) (*StatusPage, error) {
	var page StatusPage
	err := c.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/status-pages/%d", pageID), nil, &page)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func (c *Client) UpdateStatusPage(ctx context.Context, pageID int, req StatusPageUpdateRequest) (*StatusPage, error) {
	var page StatusPage
	err := c.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/api/status-pages/%d", pageID), req, &page)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func (c *Client) DeleteStatusPage(ctx context.Context, pageID int) error {
	return c.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/status-pages/%d", pageID), nil, nil)
}
