package maintenance_window

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	syschecks "github.com/systeampl/syschecks-go"
	"github.com/systeampl/syschecks-go/models"
	"github.com/systeampl/terraform-provider-systeam/internal/sdkutil"
)

var (
	_ resource.Resource                = &maintenanceWindowResource{}
	_ resource.ResourceWithConfigure   = &maintenanceWindowResource{}
	_ resource.ResourceWithImportState = &maintenanceWindowResource{}
)

type maintenanceWindowResource struct {
	sdk *syschecks.Client
}

func NewResource() resource.Resource {
	return &maintenanceWindowResource{}
}

func (r *maintenanceWindowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_maintenance_window"
}

func (r *maintenanceWindowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	emptyIntList, _ := types.ListValue(types.Int64Type, []attr.Value{})

	resp.Schema = schema.Schema{
		Description: "Manages a maintenance window.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier of the maintenance window.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the maintenance window.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "A description of the maintenance window.",
			},
			"start_time": schema.StringAttribute{
				Required:    true,
				Description: "Start time in RFC3339 format.",
			},
			"end_time": schema.StringAttribute{
				Required:    true,
				Description: "End time in RFC3339 format.",
			},
			"timezone": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("UTC"),
				Description: "Timezone for the maintenance window.",
			},
			"organization_id": schema.Int64Attribute{
				Optional:    true,
				Description: "Organization that owns the window. REQUIRED for non-superadmin callers; omit only for a superadmin platform-global window.",
			},
			"check_ids": schema.ListAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(emptyIntList),
				Description: "List of check IDs included in the maintenance window.",
			},
			"project_ids": schema.ListAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(emptyIntList),
				Description: "List of project IDs included in the maintenance window.",
			},
			"is_recurring": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the maintenance window is recurring.",
			},
			"recurrence_pattern": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Computed:    true,
				Description: "Recurrence as JSON when is_recurring=true, e.g. jsonencode({type=\"weekly\",days=[0,3],time=\"02:00\",duration_minutes=120}). days: Mon=0..Sun=6.",
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the maintenance window is active.",
			},
		},
	}
}

func (r *maintenanceWindowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *maintenanceWindowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MaintenanceWindowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkIDs := extractIntList(ctx, plan.CheckIDs)
	projectIDs := extractIntList(ctx, plan.ProjectIDs)

	createReq := models.CreateMaintenanceWindowJSONRequestBody{
		Name:              plan.Name.ValueString(),
		Description:       sdkutil.StrPtr(plan.Description),
		StartTime:         plan.StartTime.ValueString(),
		EndTime:           plan.EndTime.ValueString(),
		Timezone:          sdkutil.StrPtr(plan.Timezone),
		CheckIds:          &checkIDs,
		ProjectIds:        &projectIDs,
		IsRecurring:       plan.IsRecurring.ValueBoolPointer(),
		RecurrencePattern: normToPattern(plan.RecurrencePattern),
	}
	if !plan.OrganizationID.IsNull() && !plan.OrganizationID.IsUnknown() {
		v := int(plan.OrganizationID.ValueInt64())
		createReq.OrganizationId = &v
	}

	mw, err := r.sdk.Maintenance.CreateMaintenanceWindow(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating maintenance window", err.Error())
		return
	}

	mapMaintenanceWindowToState(mw, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *maintenanceWindowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MaintenanceWindowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mw, err := r.sdk.Maintenance.GetMaintenanceWindow(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if sdkutil.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading maintenance window", err.Error())
		return
	}

	mapMaintenanceWindowToState(mw, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *maintenanceWindowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MaintenanceWindowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkIDs := extractIntList(ctx, plan.CheckIDs)
	projectIDs := extractIntList(ctx, plan.ProjectIDs)

	updateReq := models.UpdateMaintenanceWindowJSONRequestBody{
		Name:              plan.Name.ValueStringPointer(),
		Description:       sdkutil.StrPtr(plan.Description),
		StartTime:         plan.StartTime.ValueStringPointer(),
		EndTime:           plan.EndTime.ValueStringPointer(),
		Timezone:          sdkutil.StrPtr(plan.Timezone),
		CheckIds:          &checkIDs,
		ProjectIds:        &projectIDs,
		IsRecurring:       plan.IsRecurring.ValueBoolPointer(),
		IsActive:          plan.IsActive.ValueBoolPointer(),
		RecurrencePattern: normToPattern(plan.RecurrencePattern),
	}

	mw, err := r.sdk.Maintenance.UpdateMaintenanceWindow(ctx, int(plan.ID.ValueInt64()), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating maintenance window", err.Error())
		return
	}

	mapMaintenanceWindowToState(mw, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *maintenanceWindowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MaintenanceWindowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.sdk.Maintenance.DeleteMaintenanceWindow(ctx, int(state.ID.ValueInt64())); err != nil {
		if sdkutil.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting maintenance window", err.Error())
	}
}

func (r *maintenanceWindowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected integer maintenance window ID, got: %s", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func extractIntList(ctx context.Context, list types.List) []int {
	if list.IsNull() || list.IsUnknown() {
		return []int{}
	}
	elements := make([]types.Int64, 0, len(list.Elements()))
	list.ElementsAs(ctx, &elements, false)
	result := make([]int, len(elements))
	for i, e := range elements {
		result[i] = int(e.ValueInt64())
	}
	return result
}

func mapMaintenanceWindowToState(mw *models.MaintenanceWindowResponse, state *MaintenanceWindowModel) {
	state.ID = types.Int64Value(int64(mw.Id))
	state.Name = types.StringValue(mw.Name)
	state.Description = types.StringValue(derefStr(mw.Description))
	state.StartTime = types.StringValue(mw.StartTime)
	state.EndTime = types.StringValue(mw.EndTime)
	state.Timezone = types.StringValue(derefStr(mw.Timezone))
	state.IsRecurring = types.BoolValue(derefBool(mw.IsRecurring))
	state.IsActive = types.BoolValue(mw.IsActive)
	if mw.OrganizationId != nil {
		state.OrganizationID = types.Int64Value(int64(*mw.OrganizationId))
	} else {
		state.OrganizationID = types.Int64Null()
	}
	state.RecurrencePattern = patternToNormalized(mw.RecurrencePattern)

	state.CheckIDs = intListToState(mw.CheckIds)
	state.ProjectIDs = intListToState(mw.ProjectIds)
}

func intListToState(ids *[]int) types.List {
	var src []int
	if ids != nil {
		src = *ids
	}
	values := make([]attr.Value, len(src))
	for i, id := range src {
		values[i] = types.Int64Value(int64(id))
	}
	list, _ := types.ListValue(types.Int64Type, values)
	return list
}

func normToPattern(n jsontypes.Normalized) *map[string]interface{} {
	if n.IsNull() || n.IsUnknown() {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(n.ValueString()), &m); err != nil {
		return nil
	}
	return &m
}

func patternToNormalized(p *map[string]interface{}) jsontypes.Normalized {
	if p == nil {
		return jsontypes.NewNormalizedNull()
	}
	b, err := json.Marshal(*p)
	if err != nil {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(b))
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func derefBool(p *bool) bool {
	return p != nil && *p
}
