package client

import (
	"context"
	"fmt"
	"net/http"
)

type Project struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	OrganizationID       int    `json:"organization_id"`
	AccessControlEnabled bool   `json:"access_control_enabled"`
}

type ProjectCreateRequest struct {
	Name                 string `json:"name"`
	Description          string `json:"description,omitempty"`
	OrganizationID       int    `json:"organization_id"`
	AccessControlEnabled bool   `json:"access_control_enabled"`
}

type ProjectUpdateRequest struct {
	Name                 *string `json:"name,omitempty"`
	Description          *string `json:"description,omitempty"`
	AccessControlEnabled *bool   `json:"access_control_enabled,omitempty"`
}

func (c *Client) CreateProject(ctx context.Context, orgID int, req ProjectCreateRequest) (*Project, error) {
	var project Project
	err := c.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/api/organizations/%d/projects", orgID), req, &project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (c *Client) GetProject(ctx context.Context, orgID, projectID int) (*Project, error) {
	var project Project
	err := c.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/organizations/%d/projects/%d", orgID, projectID), nil, &project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (c *Client) UpdateProject(ctx context.Context, orgID, projectID int, req ProjectUpdateRequest) (*Project, error) {
	var project Project
	err := c.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/api/organizations/%d/projects/%d", orgID, projectID), req, &project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (c *Client) DeleteProject(ctx context.Context, orgID, projectID int) error {
	return c.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/organizations/%d/projects/%d", orgID, projectID), nil, nil)
}
