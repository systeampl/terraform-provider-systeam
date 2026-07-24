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
	"github.com/systeampl/terraform-provider-systeam/internal/client"
)

var (
	_ resource.Resource                = &maintenanceWindowResource{}
	_ resource.ResourceWithConfigure   = &maintenanceWindowResource{}
	_ resource.ResourceWithImportState = &maintenanceWindowResource{}
)

type maintenanceWindowResource struct {
	client *client.Client
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
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *maintenanceWindowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MaintenanceWindowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkIDs := extractIntList(ctx, plan.CheckIDs)
	projectIDs := extractIntList(ctx, plan.ProjectIDs)

	createReq := client.MaintenanceWindowCreateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		StartTime:   plan.StartTime.ValueString(),
		EndTime:     plan.EndTime.ValueString(),
		Timezone:    plan.Timezone.ValueString(),
		CheckIDs:    checkIDs,
		ProjectIDs:  projectIDs,
		IsRecurring: plan.IsRecurring.ValueBool(),
		IsActive:    plan.IsActive.ValueBool(),
	}
	if !plan.OrganizationID.IsNull() && !plan.OrganizationID.IsUnknown() {
		v := int(plan.OrganizationID.ValueInt64())
		createReq.OrganizationID = &v
	}
	createReq.RecurrencePattern = normToRaw(plan.RecurrencePattern)

	mw, err := r.client.CreateMaintenanceWindow(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating maintenance window", err.Error())
		return
	}

	mapMaintenanceWindowToState(ctx, mw, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *maintenanceWindowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MaintenanceWindowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mw, err := r.client.GetMaintenanceWindow(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading maintenance window", err.Error())
		return
	}

	mapMaintenanceWindowToState(ctx, mw, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *maintenanceWindowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MaintenanceWindowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	desc := plan.Description.ValueString()
	startTime := plan.StartTime.ValueString()
	endTime := plan.EndTime.ValueString()
	tz := plan.Timezone.ValueString()
	isRecurring := plan.IsRecurring.ValueBool()
	isActive := plan.IsActive.ValueBool()
	checkIDs := extractIntList(ctx, plan.CheckIDs)
	projectIDs := extractIntList(ctx, plan.ProjectIDs)

	updateReq := client.MaintenanceWindowUpdateRequest{
		Name:        &name,
		Description: &desc,
		StartTime:   &startTime,
		EndTime:     &endTime,
		Timezone:    &tz,
		CheckIDs:    &checkIDs,
		ProjectIDs:  &projectIDs,
		IsRecurring: &isRecurring,
		IsActive:    &isActive,
	}
	if !plan.OrganizationID.IsNull() && !plan.OrganizationID.IsUnknown() {
		v := int(plan.OrganizationID.ValueInt64())
		updateReq.OrganizationID = &v
	}
	updateReq.RecurrencePattern = normToRaw(plan.RecurrencePattern)

	mw, err := r.client.UpdateMaintenanceWindow(ctx, int(plan.ID.ValueInt64()), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating maintenance window", err.Error())
		return
	}

	mapMaintenanceWindowToState(ctx, mw, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *maintenanceWindowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MaintenanceWindowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMaintenanceWindow(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
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

func mapMaintenanceWindowToState(ctx context.Context, mw *client.MaintenanceWindow, state *MaintenanceWindowModel) {
	state.ID = types.Int64Value(int64(mw.ID))
	state.Name = types.StringValue(mw.Name)
	state.Description = types.StringValue(mw.Description)
	state.StartTime = types.StringValue(mw.StartTime)
	state.EndTime = types.StringValue(mw.EndTime)
	state.Timezone = types.StringValue(mw.Timezone)
	state.IsRecurring = types.BoolValue(mw.IsRecurring)
	state.IsActive = types.BoolValue(mw.IsActive)
	if mw.OrganizationID != nil {
		state.OrganizationID = types.Int64Value(int64(*mw.OrganizationID))
	} else {
		state.OrganizationID = types.Int64Null()
	}
	state.RecurrencePattern = rawToNormalized(mw.RecurrencePattern)

	checkIDValues := make([]attr.Value, len(mw.CheckIDs))
	for i, id := range mw.CheckIDs {
		checkIDValues[i] = types.Int64Value(int64(id))
	}
	state.CheckIDs, _ = types.ListValue(types.Int64Type, checkIDValues)

	projectIDValues := make([]attr.Value, len(mw.ProjectIDs))
	for i, id := range mw.ProjectIDs {
		projectIDValues[i] = types.Int64Value(int64(id))
	}
	state.ProjectIDs, _ = types.ListValue(types.Int64Type, projectIDValues)
}

func normToRaw(n jsontypes.Normalized) json.RawMessage {
	if n.IsNull() || n.IsUnknown() {
		return nil
	}
	return json.RawMessage(n.ValueString())
}

func rawToNormalized(raw json.RawMessage) jsontypes.Normalized {
	s := string(raw)
	if len(raw) == 0 || s == "null" {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(s)
}
