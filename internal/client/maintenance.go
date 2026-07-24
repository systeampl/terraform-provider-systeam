package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type MaintenanceWindow struct {
	ID                int             `json:"id"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	StartTime         string          `json:"start_time"`
	EndTime           string          `json:"end_time"`
	Timezone          string          `json:"timezone"`
	OrganizationID    *int            `json:"organization_id"`
	CheckIDs          []int           `json:"check_ids"`
	ProjectIDs        []int           `json:"project_ids"`
	IsRecurring       bool            `json:"is_recurring"`
	RecurrencePattern json.RawMessage `json:"recurrence_pattern"`
	IsActive          bool            `json:"is_active"`
}

type MaintenanceWindowCreateRequest struct {
	Name              string          `json:"name"`
	Description       string          `json:"description,omitempty"`
	StartTime         string          `json:"start_time"`
	EndTime           string          `json:"end_time"`
	Timezone          string          `json:"timezone,omitempty"`
	OrganizationID    *int            `json:"organization_id,omitempty"`
	CheckIDs          []int           `json:"check_ids,omitempty"`
	ProjectIDs        []int           `json:"project_ids,omitempty"`
	IsRecurring       bool            `json:"is_recurring"`
	RecurrencePattern json.RawMessage `json:"recurrence_pattern,omitempty"`
	IsActive          bool            `json:"is_active"`
}

type MaintenanceWindowUpdateRequest struct {
	Name              *string         `json:"name,omitempty"`
	Description       *string         `json:"description,omitempty"`
	StartTime         *string         `json:"start_time,omitempty"`
	EndTime           *string         `json:"end_time,omitempty"`
	Timezone          *string         `json:"timezone,omitempty"`
	OrganizationID    *int            `json:"organization_id,omitempty"`
	CheckIDs          *[]int          `json:"check_ids,omitempty"`
	ProjectIDs        *[]int          `json:"project_ids,omitempty"`
	IsRecurring       *bool           `json:"is_recurring,omitempty"`
	RecurrencePattern json.RawMessage `json:"recurrence_pattern,omitempty"`
	IsActive          *bool           `json:"is_active,omitempty"`
}

func (c *Client) CreateMaintenanceWindow(ctx context.Context, req MaintenanceWindowCreateRequest) (*MaintenanceWindow, error) {
	var mw MaintenanceWindow
	err := c.DoRequest(ctx, http.MethodPost, "/api/maintenance", req, &mw)
	if err != nil {
		return nil, err
	}
	return &mw, nil
}

func (c *Client) GetMaintenanceWindow(ctx context.Context, id int) (*MaintenanceWindow, error) {
	var mw MaintenanceWindow
	err := c.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/maintenance/%d", id), nil, &mw)
	if err != nil {
		return nil, err
	}
	return &mw, nil
}

func (c *Client) UpdateMaintenanceWindow(ctx context.Context, id int, req MaintenanceWindowUpdateRequest) (*MaintenanceWindow, error) {
	var mw MaintenanceWindow
	err := c.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/api/maintenance/%d", id), req, &mw)
	if err != nil {
		return nil, err
	}
	return &mw, nil
}

func (c *Client) DeleteMaintenanceWindow(ctx context.Context, id int) error {
	return c.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/maintenance/%d", id), nil, nil)
}
