package client

import (
	"context"
	"fmt"
	"net/http"
)

type OnCallParticipant struct {
	ID       int `json:"id"`
	UserID   int `json:"user_id"`
	Position int `json:"position"`
}

type OnCallSchedule struct {
	ID             int                 `json:"id"`
	Name           string              `json:"name"`
	OrganizationID int                 `json:"organization_id"`
	Timezone       string              `json:"timezone"`
	RotationType   string              `json:"rotation_type"`
	IsActive       bool                `json:"is_active"`
	Participants   []OnCallParticipant `json:"participants"`
}

type OnCallScheduleCreateRequest struct {
	Name         string              `json:"name"`
	Timezone     string              `json:"timezone"`
	RotationType string              `json:"rotation_type"`
	IsActive     bool                `json:"is_active"`
	Participants []OnCallParticipant `json:"participants,omitempty"`
}

type OnCallScheduleUpdateRequest struct {
	Name         *string             `json:"name,omitempty"`
	Timezone     *string             `json:"timezone,omitempty"`
	RotationType *string             `json:"rotation_type,omitempty"`
	IsActive     *bool               `json:"is_active,omitempty"`
	Participants []OnCallParticipant `json:"participants,omitempty"`
}

func (c *Client) CreateOnCallSchedule(ctx context.Context, orgID int, req OnCallScheduleCreateRequest) (*OnCallSchedule, error) {
	var schedule OnCallSchedule
	err := c.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/api/organizations/%d/oncall-schedules", orgID), req, &schedule)
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (c *Client) GetOnCallSchedule(ctx context.Context, orgID, scheduleID int) (*OnCallSchedule, error) {
	var schedule OnCallSchedule
	err := c.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/organizations/%d/oncall-schedules/%d", orgID, scheduleID), nil, &schedule)
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (c *Client) UpdateOnCallSchedule(ctx context.Context, orgID, scheduleID int, req OnCallScheduleUpdateRequest) (*OnCallSchedule, error) {
	var schedule OnCallSchedule
	err := c.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/api/organizations/%d/oncall-schedules/%d", orgID, scheduleID), req, &schedule)
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (c *Client) DeleteOnCallSchedule(ctx context.Context, orgID, scheduleID int) error {
	return c.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/organizations/%d/oncall-schedules/%d", orgID, scheduleID), nil, nil)
}
