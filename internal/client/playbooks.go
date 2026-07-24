package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// PlaybookStep is one stage of an incident-automation playbook. config and
// conditions are action-specific JSON blobs kept opaque here (the provider
// carries them as normalized JSON).
type PlaybookStep struct {
	StepOrder      int             `json:"step_order"`
	Name           string          `json:"name"`
	ActionType     string          `json:"action_type"`
	Config         json.RawMessage `json:"config,omitempty"`
	Conditions     json.RawMessage `json:"conditions,omitempty"`
	ParallelGroup  *string         `json:"parallel_group,omitempty"`
	IsManual       bool            `json:"is_manual"`
	TimeoutSeconds int             `json:"timeout_seconds"`
}

// Playbook is a multi-stage incident-response automation (trigger -> ordered
// steps), modeled on spike.sh-style runbooks.
type Playbook struct {
	ID                           int             `json:"id"`
	OrganizationID               int             `json:"organization_id"`
	Name                         string          `json:"name"`
	Description                  string          `json:"description"`
	TriggerType                  string          `json:"trigger_type"`
	TriggerConditions            json.RawMessage `json:"trigger_conditions,omitempty"`
	ServiceID                    *int            `json:"service_id"`
	SuppressDefaultNotifications bool            `json:"suppress_default_notifications"`
	Steps                        []PlaybookStep  `json:"steps"`
	IsActive                     bool            `json:"is_active"`
}

// PlaybookWriteRequest is the create/update body (create ignores id/is_active).
type PlaybookWriteRequest struct {
	Name                         string          `json:"name"`
	Description                  string          `json:"description,omitempty"`
	TriggerType                  string          `json:"trigger_type"`
	TriggerConditions            json.RawMessage `json:"trigger_conditions,omitempty"`
	ServiceID                    *int            `json:"service_id,omitempty"`
	SuppressDefaultNotifications bool            `json:"suppress_default_notifications"`
	Steps                        []PlaybookStep  `json:"steps"`
	IsActive                     *bool           `json:"is_active,omitempty"`
}

func (c *Client) CreatePlaybook(ctx context.Context, orgID int, req PlaybookWriteRequest) (*Playbook, error) {
	var pb Playbook
	err := c.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/api/organizations/%d/playbooks", orgID), req, &pb)
	if err != nil {
		return nil, err
	}
	return &pb, nil
}

func (c *Client) GetPlaybook(ctx context.Context, orgID, playbookID int) (*Playbook, error) {
	var pb Playbook
	err := c.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/organizations/%d/playbooks/%d", orgID, playbookID), nil, &pb)
	if err != nil {
		return nil, err
	}
	return &pb, nil
}

func (c *Client) UpdatePlaybook(ctx context.Context, orgID, playbookID int, req PlaybookWriteRequest) (*Playbook, error) {
	var pb Playbook
	err := c.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/api/organizations/%d/playbooks/%d", orgID, playbookID), req, &pb)
	if err != nil {
		return nil, err
	}
	return &pb, nil
}

func (c *Client) DeletePlaybook(ctx context.Context, orgID, playbookID int) error {
	return c.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/organizations/%d/playbooks/%d", orgID, playbookID), nil, nil)
}
