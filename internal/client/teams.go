package client

import (
	"context"
	"fmt"
	"net/http"
)

// Team is an org-scoped grouping used for ownership/routing. Projects, on-call
// schedules, escalation policies and integration keys can belong to a team.
type Team struct {
	ID             int    `json:"id"`
	OrganizationID int    `json:"organization_id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Description    string `json:"description"`
	IsActive       bool   `json:"is_active"`
	MemberCount    int    `json:"member_count"`
}

// TeamCreateRequest creates a team. Slug is optional — the server slugifies the
// name when it is empty. Slug cannot be changed afterwards.
type TeamCreateRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
}

// TeamUpdateRequest updates the mutable fields. The API applies a field only
// when it is present, so both are always sent to make intent explicit.
type TeamUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *Client) CreateTeam(ctx context.Context, orgID int, req TeamCreateRequest) (*Team, error) {
	var team Team
	err := c.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/api/organizations/%d/teams", orgID), req, &team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (c *Client) GetTeam(ctx context.Context, orgID, teamID int) (*Team, error) {
	var team Team
	err := c.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/organizations/%d/teams/%d", orgID, teamID), nil, &team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (c *Client) UpdateTeam(ctx context.Context, orgID, teamID int, req TeamUpdateRequest) (*Team, error) {
	var team Team
	err := c.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/api/organizations/%d/teams/%d", orgID, teamID), req, &team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (c *Client) DeleteTeam(ctx context.Context, orgID, teamID int) error {
	return c.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/organizations/%d/teams/%d", orgID, teamID), nil, nil)
}
