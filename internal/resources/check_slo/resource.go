package check_slo

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	syschecks "github.com/systeampl/syschecks-go"
	"github.com/systeampl/syschecks-go/models"
	"github.com/systeampl/terraform-provider-systeam/internal/sdkutil"
)

var (
	_ resource.Resource                = &checkSLOResource{}
	_ resource.ResourceWithConfigure   = &checkSLOResource{}
	_ resource.ResourceWithImportState = &checkSLOResource{}
)

type checkSLOResource struct {
	sdk *syschecks.Client
}

func NewResource() resource.Resource {
	return &checkSLOResource{}
}

func (r *checkSLOResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_check_slo"
}

func (r *checkSLOResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an SLO (Service Level Objective) for a check. One SLO per check.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier of the SLO.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"check_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the check this SLO belongs to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"sli_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("availability"),
				Description: "The SLI type. Currently only 'availability' is supported.",
			},
			"target_percentage": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(99.9),
				Description: "The SLO target percentage (90.0 - 100.0).",
			},
			"window_days": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(30),
				Description: "The SLO window in days (7, 30, or 90).",
			},
			"latency_threshold_ms": schema.Int64Attribute{
				Optional:    true,
				Description: "Latency threshold in milliseconds for latency-based SLIs.",
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the SLO is active.",
			},
			"notify_on_budget_warn": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Send notification when error budget reaches warning threshold.",
			},
			"budget_warn_pct": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(20.0),
				Description: "Error budget warning threshold percentage.",
			},
			"notify_on_budget_exhausted": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Send notification when error budget is exhausted.",
			},
			"burn_rate_alert_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether burn rate alerts are enabled.",
			},
			"burn_rate_threshold": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(14.4),
				Description: "Burn rate threshold for alerts.",
			},
			"burn_rate_window_minutes": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
				Description: "Burn rate evaluation window in minutes.",
			},
		},
	}
}

func (r *checkSLOResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	sdk, ok := req.ProviderData.(*syschecks.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *syschecks.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.sdk = sdk
}

func (r *checkSLOResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CheckSLOModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sliType := plan.SLIType.ValueString()
	targetPct := plan.TargetPercentage.ValueFloat64()
	windowDays := int(plan.WindowDays.ValueInt64())
	notifyWarn := plan.NotifyOnBudgetWarn.ValueBool()
	budgetWarnPct := plan.BudgetWarnPct.ValueFloat64()
	notifyExhausted := plan.NotifyOnBudgetExhausted.ValueBool()
	burnRateEnabled := plan.BurnRateAlertEnabled.ValueBool()
	burnRateThreshold := plan.BurnRateThreshold.ValueFloat64()
	burnRateWindow := int(plan.BurnRateWindowMinutes.ValueInt64())

	createReq := models.CreateCheckSloJSONRequestBody{
		SliType:                 &sliType,
		TargetPercentage:        &targetPct,
		WindowDays:              &windowDays,
		NotifyOnBudgetWarn:      &notifyWarn,
		BudgetWarnPct:           &budgetWarnPct,
		NotifyOnBudgetExhausted: &notifyExhausted,
		BurnRateAlertEnabled:    &burnRateEnabled,
		BurnRateThreshold:       &burnRateThreshold,
		BurnRateWindowMinutes:   &burnRateWindow,
	}

	if !plan.LatencyThresholdMs.IsNull() && !plan.LatencyThresholdMs.IsUnknown() {
		v := int(plan.LatencyThresholdMs.ValueInt64())
		createReq.LatencyThresholdMs = &v
	}

	slo, err := r.sdk.Checks.CreateCheckSlo(ctx, int(plan.CheckID.ValueInt64()), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating check SLO", err.Error())
		return
	}

	mapCheckSLOToState(slo, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *checkSLOResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CheckSLOModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.sdk.Checks.GetCheckSlo(ctx, int(state.CheckID.ValueInt64()))
	if err != nil {
		if sdkutil.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading check SLO", err.Error())
		return
	}
	if status.Slo == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	mapCheckSLOToState(status.Slo, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *checkSLOResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CheckSLOModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetPct := plan.TargetPercentage.ValueFloat64()
	windowDays := int(plan.WindowDays.ValueInt64())
	isActive := plan.IsActive.ValueBool()
	notifyWarn := plan.NotifyOnBudgetWarn.ValueBool()
	budgetWarnPct := plan.BudgetWarnPct.ValueFloat64()
	notifyExhausted := plan.NotifyOnBudgetExhausted.ValueBool()
	burnRateEnabled := plan.BurnRateAlertEnabled.ValueBool()
	burnRateThreshold := plan.BurnRateThreshold.ValueFloat64()
	burnRateWindow := int(plan.BurnRateWindowMinutes.ValueInt64())

	updateReq := models.UpdateCheckSloJSONRequestBody{
		TargetPercentage:        &targetPct,
		WindowDays:              &windowDays,
		IsActive:                &isActive,
		NotifyOnBudgetWarn:      &notifyWarn,
		BudgetWarnPct:           &budgetWarnPct,
		NotifyOnBudgetExhausted: &notifyExhausted,
		BurnRateAlertEnabled:    &burnRateEnabled,
		BurnRateThreshold:       &burnRateThreshold,
		BurnRateWindowMinutes:   &burnRateWindow,
	}

	if !plan.LatencyThresholdMs.IsNull() && !plan.LatencyThresholdMs.IsUnknown() {
		v := int(plan.LatencyThresholdMs.ValueInt64())
		updateReq.LatencyThresholdMs = &v
	}

	slo, err := r.sdk.Checks.UpdateCheckSlo(ctx, int(plan.CheckID.ValueInt64()), int(plan.ID.ValueInt64()), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating check SLO", err.Error())
		return
	}

	mapCheckSLOToState(slo, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *checkSLOResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CheckSLOModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.sdk.Checks.DeleteCheckSlo(ctx, int(state.CheckID.ValueInt64()), int(state.ID.ValueInt64())); err != nil {
		if sdkutil.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting check SLO", err.Error())
	}
}

func (r *checkSLOResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	checkID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected integer check ID, got: %s", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("check_id"), checkID)...)
}

func mapCheckSLOToState(slo *models.SLOResponse, state *CheckSLOModel) {
	state.ID = types.Int64Value(int64(slo.Id))
	state.CheckID = types.Int64Value(int64(slo.CheckId))
	state.SLIType = types.StringValue(slo.SliType)
	state.TargetPercentage = types.Float64Value(float64(slo.TargetPercentage))
	state.WindowDays = types.Int64Value(int64(slo.WindowDays))
	state.IsActive = types.BoolValue(slo.IsActive)
	state.NotifyOnBudgetWarn = types.BoolValue(slo.NotifyOnBudgetWarn)
	state.BudgetWarnPct = types.Float64Value(float64(slo.BudgetWarnPct))
	state.NotifyOnBudgetExhausted = types.BoolValue(slo.NotifyOnBudgetExhausted)
	state.BurnRateAlertEnabled = types.BoolValue(slo.BurnRateAlertEnabled)
	state.BurnRateThreshold = types.Float64Value(float64(slo.BurnRateThreshold))
	state.BurnRateWindowMinutes = types.Int64Value(int64(slo.BurnRateWindowMinutes))

	if slo.LatencyThresholdMs != nil {
		state.LatencyThresholdMs = types.Int64Value(int64(*slo.LatencyThresholdMs))
	} else {
		state.LatencyThresholdMs = types.Int64Null()
	}
}
