package client

import (
	"context"
	"net/http"
)

// AgentRegistrationToken is a one-time, short-lived (1h) credential used to
// enroll a new private or geo agent. The backend exposes only a create endpoint:
// there is no list/get/delete — the token is write-only and consumed on first
// use (the agent trades it for a permanent api_key via POST /api/agents/register).
// Modeling note: like a generated password, the token is captured into Terraform
// state at create and never read back.
type AgentRegistrationToken struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	Mode      string `json:"mode"`
}

type AgentRegistrationTokenCreateRequest struct {
	Name           string `json:"name"`
	Mode           string `json:"mode,omitempty"`
	OrganizationID int    `json:"organization_id"`
}

// CreateAgentRegistrationToken mints an enrollment token for a private agent.
func (c *Client) CreateAgentRegistrationToken(ctx context.Context, req AgentRegistrationTokenCreateRequest) (*AgentRegistrationToken, error) {
	var tok AgentRegistrationToken
	err := c.DoRequest(ctx, http.MethodPost, "/api/agents/registration-token", req, &tok)
	if err != nil {
		return nil, err
	}
	return &tok, nil
}
