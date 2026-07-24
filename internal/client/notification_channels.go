package client

import (
	"context"
	"fmt"
	"net/http"
)

type NotificationChannel struct {
	ID             int            `json:"id"`
	Name           string         `json:"name"`
	ChannelType    string         `json:"channel_type"`
	Config         map[string]any `json:"config"`
	IsActive       bool           `json:"is_active"`
	OrganizationID *int           `json:"organization_id"`
}

type NotificationChannelCreateRequest struct {
	Name           string         `json:"name"`
	ChannelType    string         `json:"channel_type"`
	Config         map[string]any `json:"config"`
	IsActive       bool           `json:"is_active"`
	OrganizationID *int           `json:"organization_id,omitempty"`
}

type NotificationChannelUpdateRequest struct {
	Name     *string         `json:"name,omitempty"`
	Config   *map[string]any `json:"config,omitempty"`
	IsActive *bool           `json:"is_active,omitempty"`
}

func (c *Client) CreateNotificationChannel(ctx context.Context, req NotificationChannelCreateRequest) (*NotificationChannel, error) {
	var channel NotificationChannel
	err := c.DoRequest(ctx, http.MethodPost, "/api/notification-channels", req, &channel)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (c *Client) GetNotificationChannel(ctx context.Context, id int) (*NotificationChannel, error) {
	var channel NotificationChannel
	err := c.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/notification-channels/%d", id), nil, &channel)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (c *Client) UpdateNotificationChannel(ctx context.Context, id int, req NotificationChannelUpdateRequest) (*NotificationChannel, error) {
	var channel NotificationChannel
	err := c.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/api/notification-channels/%d", id), req, &channel)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (c *Client) DeleteNotificationChannel(ctx context.Context, id int) error {
	return c.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/notification-channels/%d", id), nil, nil)
}
