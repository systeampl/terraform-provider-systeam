package playbook

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pawel-cygal/terraform-provider-systeam/internal/client"
)

// The playbook action/trigger vocabularies mirror the backend
// (PlaybookStepSchema / PlaybookCreate). Keep in lockstep with the API.
var validActionTypes = []string{
	"add_responders", "set_priority", "set_severity", "setup_war_room",
	"create_jira_ticket", "auto_ack", "auto_resolve", "update_status_page",
	"notify_subscribers", "run_escalation", "outbound_webhook",
	"create_slack_channel", "add_links", "title_enrichment",
	"send_custom_notification", "manual_step",
}

var validTriggerTypes = []string{
	"check_status_change", "inbound_incident_created", "inbound_status_change", "manual",
}

var (
	_ resource.Resource                = &playbookResource{}
	_ resource.ResourceWithConfigure   = &playbookResource{}
	_ resource.ResourceWithImportState = &playbookResource{}
)

type playbookResource struct {
	client *client.Client
}

func NewResource() resource.Resource {
	return &playbookResource{}
}

func (r *playbookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_playbook"
}

func (r *playbookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a multi-stage incident-response playbook (spike.sh-style automation): a " +
			"trigger plus an ordered list of steps, each running one of the supported actions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:      true,
				Description:   "The unique identifier of the playbook.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.Int64Attribute{
				Required:      true,
				Description:   "The organization this playbook belongs to.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The playbook name (1-200 chars).",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "An optional description.",
			},
			"trigger_type": schema.StringAttribute{
				Required:    true,
				Description: "What fires the playbook: " + strings.Join(validTriggerTypes, ", ") + ".",
				Validators:  []validator.String{stringvalidator.OneOf(validTriggerTypes...)},
			},
			"trigger_conditions": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Computed:    true,
				Description: "JSON object of extra conditions the trigger must match.",
			},
			"service_id": schema.Int64Attribute{
				Optional:    true,
				Description: "Optionally scope the playbook to a service.",
			},
			"suppress_default_notifications": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Suppress the product's default notifications while this playbook runs.",
			},
			"is_active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the playbook is active.",
			},
		},
		Blocks: map[string]schema.Block{
			"step": schema.ListNestedBlock{
				Description: "Ordered steps executed when the playbook runs.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"step_order": schema.Int64Attribute{
							Required:    true,
							Description: "Position of this step in the run (1-based).",
						},
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Human-readable step name.",
						},
						"action_type": schema.StringAttribute{
							Required:    true,
							Description: "The action to run. One of: " + strings.Join(validActionTypes, ", ") + ".",
							Validators:  []validator.String{stringvalidator.OneOf(validActionTypes...)},
						},
						"config": schema.StringAttribute{
							CustomType:  jsontypes.NormalizedType{},
							Optional:    true,
							Computed:    true,
							Description: "Action-specific configuration as a JSON object.",
						},
						"conditions": schema.StringAttribute{
							CustomType:  jsontypes.NormalizedType{},
							Optional:    true,
							Description: "Optional JSON array of conditions gating this step.",
						},
						"parallel_group": schema.StringAttribute{
							Optional:    true,
							Description: "Steps sharing a parallel_group run concurrently.",
						},
						"is_manual": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "Whether the step waits for manual completion.",
						},
						"timeout_seconds": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Default:     int64default.StaticInt64(30),
							Description: "Step timeout in seconds (1-120).",
						},
					},
				},
			},
		},
	}
}

func (r *playbookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *playbookResource) buildRequest(plan *PlaybookModel) client.PlaybookWriteRequest {
	steps := make([]client.PlaybookStep, len(plan.Steps))
	for i, s := range plan.Steps {
		steps[i] = client.PlaybookStep{
			StepOrder:      int(s.StepOrder.ValueInt64()),
			Name:           s.Name.ValueString(),
			ActionType:     s.ActionType.ValueString(),
			Config:         rawFromNormalized(s.Config),
			Conditions:     rawFromNormalized(s.Conditions),
			ParallelGroup:  optionalString(s.ParallelGroup),
			IsManual:       s.IsManual.ValueBool(),
			TimeoutSeconds: int(s.TimeoutSeconds.ValueInt64()),
		}
	}
	return client.PlaybookWriteRequest{
		Name:                         plan.Name.ValueString(),
		Description:                  plan.Description.ValueString(),
		TriggerType:                  plan.TriggerType.ValueString(),
		TriggerConditions:            rawFromNormalized(plan.TriggerConditions),
		ServiceID:                    optionalInt(plan.ServiceID),
		SuppressDefaultNotifications: plan.SuppressDefaultNotifications.ValueBool(),
		Steps:                        steps,
	}
}

func (r *playbookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PlaybookModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pb, err := r.client.CreatePlaybook(ctx, int(plan.OrganizationID.ValueInt64()), r.buildRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating playbook", err.Error())
		return
	}

	resp.Diagnostics.Append(mapPlaybookToState(pb, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *playbookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PlaybookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pb, err := r.client.GetPlaybook(ctx, int(state.OrganizationID.ValueInt64()), int(state.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading playbook", err.Error())
		return
	}

	resp.Diagnostics.Append(mapPlaybookToState(pb, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *playbookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PlaybookModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pb, err := r.client.UpdatePlaybook(ctx, int(plan.OrganizationID.ValueInt64()), int(plan.ID.ValueInt64()), r.buildRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating playbook", err.Error())
		return
	}

	resp.Diagnostics.Append(mapPlaybookToState(pb, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *playbookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PlaybookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeletePlaybook(ctx, int(state.OrganizationID.ValueInt64()), int(state.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Error deleting playbook", err.Error())
	}
}

func (r *playbookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected import ID format: org_id:playbook_id")
		return
	}
	orgID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid org_id in import ID", err.Error())
		return
	}
	playbookID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid playbook_id in import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), playbookID)...)
}

func mapPlaybookToState(pb *client.Playbook, m *PlaybookModel) diag.Diagnostics {
	m.ID = types.Int64Value(int64(pb.ID))
	m.OrganizationID = types.Int64Value(int64(pb.OrganizationID))
	m.Name = types.StringValue(pb.Name)
	m.Description = types.StringValue(pb.Description)
	m.TriggerType = types.StringValue(pb.TriggerType)
	m.TriggerConditions = normalizedFromRaw(pb.TriggerConditions)
	m.ServiceID = nullableInt(pb.ServiceID)
	m.SuppressDefaultNotifications = types.BoolValue(pb.SuppressDefaultNotifications)
	m.IsActive = types.BoolValue(pb.IsActive)

	steps := make([]PlaybookStepModel, len(pb.Steps))
	for i, s := range pb.Steps {
		steps[i] = PlaybookStepModel{
			StepOrder:      types.Int64Value(int64(s.StepOrder)),
			Name:           types.StringValue(s.Name),
			ActionType:     types.StringValue(s.ActionType),
			Config:         normalizedFromRaw(s.Config),
			Conditions:     normalizedFromRaw(s.Conditions),
			ParallelGroup:  stringOrNull(s.ParallelGroup),
			IsManual:       types.BoolValue(s.IsManual),
			TimeoutSeconds: types.Int64Value(int64(s.TimeoutSeconds)),
		}
	}
	m.Steps = steps
	return nil
}

func rawFromNormalized(n jsontypes.Normalized) json.RawMessage {
	if n.IsNull() || n.IsUnknown() {
		return nil
	}
	return json.RawMessage(n.ValueString())
}

func normalizedFromRaw(raw json.RawMessage) jsontypes.Normalized {
	if len(raw) == 0 || string(raw) == "null" {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(raw))
}

func optionalString(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func stringOrNull(p *string) types.String {
	if p == nil {
		return types.StringNull()
	}
	return types.StringValue(*p)
}

func optionalInt(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := int(v.ValueInt64())
	return &i
}

func nullableInt(p *int) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*p))
}
