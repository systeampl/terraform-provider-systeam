package client

import (
	"context"
	"fmt"
	"net/http"
)

type CheckSLO struct {
	ID                      int     `json:"id"`
	CheckID                 int     `json:"check_id"`
	SLIType                 string  `json:"sli_type"`
	TargetPercentage        float64 `json:"target_percentage"`
	WindowDays              int     `json:"window_days"`
	LatencyThresholdMs      *int    `json:"latency_threshold_ms"`
	IsActive                bool    `json:"is_active"`
	NotifyOnBudgetWarn      bool    `json:"notify_on_budget_warn"`
	BudgetWarnPct           float64 `json:"budget_warn_pct"`
	NotifyOnBudgetExhausted bool    `json:"notify_on_budget_exhausted"`
	BurnRateAlertEnabled    bool    `json:"burn_rate_alert_enabled"`
	BurnRateThreshold       float64 `json:"burn_rate_threshold"`
	BurnRateWindowMinutes   int     `json:"burn_rate_window_minutes"`
}

type CheckSLOCreateRequest struct {
	SLIType                 string  `json:"sli_type"`
	TargetPercentage        float64 `json:"target_percentage"`
	WindowDays              int     `json:"window_days"`
	LatencyThresholdMs      *int    `json:"latency_threshold_ms,omitempty"`
	NotifyOnBudgetWarn      bool    `json:"notify_on_budget_warn"`
	BudgetWarnPct           float64 `json:"budget_warn_pct"`
	NotifyOnBudgetExhausted bool    `json:"notify_on_budget_exhausted"`
	BurnRateAlertEnabled    bool    `json:"burn_rate_alert_enabled"`
	BurnRateThreshold       float64 `json:"burn_rate_threshold"`
	BurnRateWindowMinutes   int     `json:"burn_rate_window_minutes"`
}

type CheckSLOUpdateRequest struct {
	TargetPercentage        *float64 `json:"target_percentage,omitempty"`
	WindowDays              *int     `json:"window_days,omitempty"`
	LatencyThresholdMs      *int     `json:"latency_threshold_ms,omitempty"`
	IsActive                *bool    `json:"is_active,omitempty"`
	NotifyOnBudgetWarn      *bool    `json:"notify_on_budget_warn,omitempty"`
	BudgetWarnPct           *float64 `json:"budget_warn_pct,omitempty"`
	NotifyOnBudgetExhausted *bool    `json:"notify_on_budget_exhausted,omitempty"`
	BurnRateAlertEnabled    *bool    `json:"burn_rate_alert_enabled,omitempty"`
	BurnRateThreshold       *float64 `json:"burn_rate_threshold,omitempty"`
	BurnRateWindowMinutes   *int     `json:"burn_rate_window_minutes,omitempty"`
}

type checkSLOGetResponse struct {
	CheckID int       `json:"check_id"`
	SLO     *CheckSLO `json:"slo"`
}

func (c *Client) CreateCheckSLO(ctx context.Context, checkID int, req CheckSLOCreateRequest) (*CheckSLO, error) {
	var slo CheckSLO
	err := c.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/api/checks/%d/slo", checkID), req, &slo)
	if err != nil {
		return nil, err
	}
	return &slo, nil
}

func (c *Client) GetCheckSLO(ctx context.Context, checkID int) (*CheckSLO, error) {
	var resp checkSLOGetResponse
	err := c.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/checks/%d/slo", checkID), nil, &resp)
	if err != nil {
		return nil, err
	}
	if resp.SLO == nil {
		return nil, &APIError{StatusCode: http.StatusNotFound, Message: "SLO not found"}
	}
	return resp.SLO, nil
}

func (c *Client) UpdateCheckSLO(ctx context.Context, checkID, sloID int, req CheckSLOUpdateRequest) (*CheckSLO, error) {
	var slo CheckSLO
	err := c.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/api/checks/%d/slo/%d", checkID, sloID), req, &slo)
	if err != nil {
		return nil, err
	}
	return &slo, nil
}

func (c *Client) DeleteCheckSLO(ctx context.Context, checkID, sloID int) error {
	return c.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/checks/%d/slo/%d", checkID, sloID), nil, nil)
}
