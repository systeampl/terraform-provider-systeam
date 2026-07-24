package client

import (
	"context"
	"fmt"
	"net/http"
)

// LifecycleWatch tracks the end-of-life / lifecycle of an external resource
// (e.g. an AI model) and notifies at thresholds. The API upserts by the natural
// key (vendor + resource_type + platform + resource_id); there is no GET-by-id,
// so reads list and match.
type LifecycleWatch struct {
	ID           int     `json:"id"`
	Vendor       string  `json:"vendor"`
	Platform     *string `json:"platform"`
	ResourceID   *string `json:"resource_id"`
	ResourceType string  `json:"resource_type"`
	NotifyOnNew  bool    `json:"notify_on_new"`
	Notify90d    bool    `json:"notify_90d"`
	Notify30d    bool    `json:"notify_30d"`
	Notify7d     bool    `json:"notify_7d"`
}

type LifecycleWatchUpsertRequest struct {
	Vendor       string  `json:"vendor"`
	Platform     *string `json:"platform,omitempty"`
	ResourceID   *string `json:"resource_id,omitempty"`
	ResourceType string  `json:"resource_type,omitempty"`
	NotifyOnNew  bool    `json:"notify_on_new"`
	Notify90d    bool    `json:"notify_90d"`
	Notify30d    bool    `json:"notify_30d"`
	Notify7d     bool    `json:"notify_7d"`
	ChannelIDs   []int   `json:"channel_ids,omitempty"`
}

// UpsertLifecycleWatch creates or updates a watch and returns its id (the only
// field the upsert endpoint echoes).
func (c *Client) UpsertLifecycleWatch(ctx context.Context, orgID int, req LifecycleWatchUpsertRequest) (int, error) {
	var resp struct {
		ID int `json:"id"`
	}
	err := c.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/api/organizations/%d/lifecycle-watches", orgID), req, &resp)
	if err != nil {
		return 0, err
	}
	return resp.ID, nil
}

// GetLifecycleWatch reads a watch by listing the org's watches and matching the
// id. Returns (nil, nil) when gone.
func (c *Client) GetLifecycleWatch(ctx context.Context, orgID, watchID int) (*LifecycleWatch, error) {
	var watches []LifecycleWatch
	err := c.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/organizations/%d/lifecycle-watches", orgID), nil, &watches)
	if err != nil {
		return nil, err
	}
	for i := range watches {
		if watches[i].ID == watchID {
			return &watches[i], nil
		}
	}
	return nil, nil
}

func (c *Client) DeleteLifecycleWatch(ctx context.Context, orgID, watchID int) error {
	return c.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/organizations/%d/lifecycle-watches/%d", orgID, watchID), nil, nil)
}
