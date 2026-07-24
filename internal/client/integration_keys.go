package client

import (
	"context"
	"fmt"
	"net/http"
)

// IntegrationKey is an inbound-events routing key: external systems
// (Alertmanager, Grafana, Prometheus, PagerDuty-format senders) POST alerts to
// /api/events/... authenticated by this key, which routes them to an escalation
// policy. There is no update endpoint on the API, so any change replaces the key.
type IntegrationKey struct {
	ID                    int    `json:"id"`
	Name                  string `json:"name"`
	TokenPrefix           string `json:"token_prefix"`
	EscalationPolicyID    int    `json:"escalation_policy_id"`
	IsActive              bool   `json:"is_active"`
	GroupingType          string `json:"grouping_type"`
	GroupingWindowSeconds int    `json:"grouping_window_seconds"`
	// Token is the full secret routing key. The API returns it ONLY in the
	// create response ("shown once"); list/read never include it.
	Token string `json:"key,omitempty"`
}

type IntegrationKeyCreateRequest struct {
	Name                  string `json:"name"`
	EscalationPolicyID    int    `json:"escalation_policy_id"`
	GroupingType          string `json:"grouping_type,omitempty"`
	GroupingWindowSeconds int    `json:"grouping_window_seconds,omitempty"`
}

func (c *Client) CreateIntegrationKey(ctx context.Context, orgID int, req IntegrationKeyCreateRequest) (*IntegrationKey, error) {
	var key IntegrationKey
	err := c.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/api/organizations/%d/integration-keys", orgID), req, &key)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// GetIntegrationKey reads a key by listing the org's keys and matching the id —
// the API exposes no GET-by-id. Returns (nil, nil) when the key is gone, so the
// resource can drop it from state.
//
// Deletion on the backend is a soft revoke (is_active=false, revoked_at set) and
// the list endpoint still returns revoked keys. A revoked key is unusable — it
// authenticates nothing — so we treat it as gone here. That way a key revoked
// out-of-band (e.g. in the UI) is detected as drift and Terraform recreates it,
// instead of silently keeping a dead key in state.
func (c *Client) GetIntegrationKey(ctx context.Context, orgID, keyID int) (*IntegrationKey, error) {
	var keys []IntegrationKey
	err := c.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/organizations/%d/integration-keys", orgID), nil, &keys)
	if err != nil {
		return nil, err
	}
	for i := range keys {
		if keys[i].ID == keyID {
			if !keys[i].IsActive {
				return nil, nil
			}
			return &keys[i], nil
		}
	}
	return nil, nil
}

func (c *Client) DeleteIntegrationKey(ctx context.Context, orgID, keyID int) error {
	return c.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/organizations/%d/integration-keys/%d", orgID, keyID), nil, nil)
}
