package client

import (
	"context"
	"fmt"
	"net/http"
)

// ContactMethod is a per-user contact endpoint (phone/email/push) used by on-call
// escalation. It is user-scoped (tied to the caller's token), not org-scoped.
// Verification (turning verified=true) happens out of band via a code and is not
// manageable as code.
type ContactMethod struct {
	ID       int     `json:"id"`
	Kind     string  `json:"kind"`
	Value    *string `json:"value"`
	Label    *string `json:"label"`
	Verified bool    `json:"verified"`
	Enabled  bool    `json:"enabled"`
}

type ContactMethodCreateRequest struct {
	Kind  string  `json:"kind"`
	Value *string `json:"value,omitempty"`
	Label *string `json:"label,omitempty"`
}

type ContactMethodUpdateRequest struct {
	Label   *string `json:"label,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}

func (c *Client) CreateContactMethod(ctx context.Context, req ContactMethodCreateRequest) (*ContactMethod, error) {
	var cm ContactMethod
	err := c.DoRequest(ctx, http.MethodPost, "/api/profile/contact-methods", req, &cm)
	if err != nil {
		return nil, err
	}
	return &cm, nil
}

// GetContactMethod reads by listing (no GET-by-id endpoint). Returns (nil, nil)
// when gone.
func (c *Client) GetContactMethod(ctx context.Context, methodID int) (*ContactMethod, error) {
	var methods []ContactMethod
	err := c.DoRequest(ctx, http.MethodGet, "/api/profile/contact-methods", nil, &methods)
	if err != nil {
		return nil, err
	}
	for i := range methods {
		if methods[i].ID == methodID {
			return &methods[i], nil
		}
	}
	return nil, nil
}

func (c *Client) UpdateContactMethod(ctx context.Context, methodID int, req ContactMethodUpdateRequest) (*ContactMethod, error) {
	var cm ContactMethod
	err := c.DoRequest(ctx, http.MethodPatch, fmt.Sprintf("/api/profile/contact-methods/%d", methodID), req, &cm)
	if err != nil {
		return nil, err
	}
	return &cm, nil
}

func (c *Client) DeleteContactMethod(ctx context.Context, methodID int) error {
	return c.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/profile/contact-methods/%d", methodID), nil, nil)
}
