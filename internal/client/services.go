package client

import (
	"context"
	"fmt"
	"net/http"
)

// Service is an org-scoped catalog entry that incidents/checks attach to.
type Service struct {
	ID                   int                 `json:"id"`
	OrganizationID       int                 `json:"organization_id"`
	Name                 string              `json:"name"`
	Slug                 string              `json:"slug"`
	Description          string              `json:"description"`
	RepoURL              string              `json:"repo_url"`
	DocsURL              string              `json:"docs_url"`
	OwnerTeamID          *int                `json:"owner_team_id"`
	EscalationPolicyID   *int                `json:"escalation_policy_id"`
	Tier                 string              `json:"tier"`
	NotificationChannels []ServiceChannelRef `json:"notification_channels"`
	IsActive             bool                `json:"is_active"`
}

// ServiceChannelRef is the response shape for an attached channel. Requests send
// bare ids (notification_channel_ids); responses return objects — use
// NotificationChannelIDs to bridge the two.
type ServiceChannelRef struct {
	ID int `json:"id"`
}

// NotificationChannelIDs extracts the attached channel ids from a read response.
func (s *Service) NotificationChannelIDs() []int {
	ids := make([]int, 0, len(s.NotificationChannels))
	for _, c := range s.NotificationChannels {
		ids = append(ids, c.ID)
	}
	return ids
}

type ServiceCreateRequest struct {
	Name                   string `json:"name"`
	Description            string `json:"description,omitempty"`
	RepoURL                string `json:"repo_url,omitempty"`
	DocsURL                string `json:"docs_url,omitempty"`
	OwnerTeamID            *int   `json:"owner_team_id,omitempty"`
	EscalationPolicyID     *int   `json:"escalation_policy_id,omitempty"`
	Tier                   string `json:"tier,omitempty"`
	NotificationChannelIDs []int  `json:"notification_channel_ids,omitempty"`
}

// ServiceUpdateRequest carries the mutable fields. Name and tier are always sent;
// the rest use omitempty semantics matching the backend's "apply when present".
type ServiceUpdateRequest struct {
	Name                   string `json:"name"`
	Description            string `json:"description,omitempty"`
	RepoURL                string `json:"repo_url,omitempty"`
	DocsURL                string `json:"docs_url,omitempty"`
	OwnerTeamID            *int   `json:"owner_team_id,omitempty"`
	EscalationPolicyID     *int   `json:"escalation_policy_id,omitempty"`
	Tier                   string `json:"tier,omitempty"`
	NotificationChannelIDs []int  `json:"notification_channel_ids,omitempty"`
}

func (c *Client) CreateService(ctx context.Context, orgID int, req ServiceCreateRequest) (*Service, error) {
	var svc Service
	err := c.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/api/organizations/%d/services", orgID), req, &svc)
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

func (c *Client) GetService(ctx context.Context, orgID, serviceID int) (*Service, error) {
	var svc Service
	err := c.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/organizations/%d/services/%d", orgID, serviceID), nil, &svc)
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

func (c *Client) UpdateService(ctx context.Context, orgID, serviceID int, req ServiceUpdateRequest) (*Service, error) {
	var svc Service
	err := c.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/api/organizations/%d/services/%d", orgID, serviceID), req, &svc)
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

func (c *Client) DeleteService(ctx context.Context, orgID, serviceID int) error {
	return c.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/organizations/%d/services/%d", orgID, serviceID), nil, nil)
}
