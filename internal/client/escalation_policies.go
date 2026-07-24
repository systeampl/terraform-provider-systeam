package client

import (
	"context"
	"fmt"
	"net/http"
)

type EscalationStep struct {
	StepOrder        int    `json:"step_order"`
	DelayMinutes     int    `json:"delay_minutes"`
	TargetType       string `json:"target_type"`
	TargetUserID     *int   `json:"target_user_id,omitempty"`
	TargetScheduleID *int   `json:"target_schedule_id,omitempty"`
	TargetChannelID  *int   `json:"target_channel_id,omitempty"`
}

type EscalationPolicy struct {
	ID             int              `json:"id"`
	Name           string           `json:"name"`
	OrganizationID int              `json:"organization_id"`
	IsActive       bool             `json:"is_active"`
	Steps          []EscalationStep `json:"steps"`
}

type EscalationPolicyCreateRequest struct {
	Name     string           `json:"name"`
	IsActive bool             `json:"is_active"`
	Steps    []EscalationStep `json:"steps,omitempty"`
}

type EscalationPolicyUpdateRequest struct {
	Name     *string          `json:"name,omitempty"`
	IsActive *bool            `json:"is_active,omitempty"`
	Steps    []EscalationStep `json:"steps,omitempty"`
}

func (c *Client) CreateEscalationPolicy(ctx context.Context, orgID int, req EscalationPolicyCreateRequest) (*EscalationPolicy, error) {
	var policy EscalationPolicy
	err := c.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/api/organizations/%d/escalation-policies", orgID), req, &policy)
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func (c *Client) GetEscalationPolicy(ctx context.Context, orgID, policyID int) (*EscalationPolicy, error) {
	var policy EscalationPolicy
	err := c.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/organizations/%d/escalation-policies/%d", orgID, policyID), nil, &policy)
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func (c *Client) UpdateEscalationPolicy(ctx context.Context, orgID, policyID int, req EscalationPolicyUpdateRequest) (*EscalationPolicy, error) {
	var policy EscalationPolicy
	err := c.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/api/organizations/%d/escalation-policies/%d", orgID, policyID), req, &policy)
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func (c *Client) DeleteEscalationPolicy(ctx context.Context, orgID, policyID int) error {
	return c.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/organizations/%d/escalation-policies/%d", orgID, policyID), nil, nil)
}
