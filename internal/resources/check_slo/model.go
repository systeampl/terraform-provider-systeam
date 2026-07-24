package check_slo

import "github.com/hashicorp/terraform-plugin-framework/types"

type CheckSLOModel struct {
	ID                      types.Int64   `tfsdk:"id"`
	CheckID                 types.Int64   `tfsdk:"check_id"`
	SLIType                 types.String  `tfsdk:"sli_type"`
	TargetPercentage        types.Float64 `tfsdk:"target_percentage"`
	WindowDays              types.Int64   `tfsdk:"window_days"`
	LatencyThresholdMs      types.Int64   `tfsdk:"latency_threshold_ms"`
	IsActive                types.Bool    `tfsdk:"is_active"`
	NotifyOnBudgetWarn      types.Bool    `tfsdk:"notify_on_budget_warn"`
	BudgetWarnPct           types.Float64 `tfsdk:"budget_warn_pct"`
	NotifyOnBudgetExhausted types.Bool    `tfsdk:"notify_on_budget_exhausted"`
	BurnRateAlertEnabled    types.Bool    `tfsdk:"burn_rate_alert_enabled"`
	BurnRateThreshold       types.Float64 `tfsdk:"burn_rate_threshold"`
	BurnRateWindowMinutes   types.Int64   `tfsdk:"burn_rate_window_minutes"`
}
