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
	syschecks "github.com/systeampl/syschecks-go"
	"github.com/systeampl/syschecks-go/models"
	"github.com/systeampl/terraform-provider-systeam/internal/sdkutil"
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
	sdk *syschecks.Client
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

func (r *playbookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PlaybookModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	suppress := plan.SuppressDefaultNotifications.ValueBool()
	pb, err := r.sdk.Playbooks.CreatePlaybook(ctx, int(plan.OrganizationID.ValueInt64()), models.CreatePlaybookJSONRequestBody{
		Name:                         plan.Name.ValueString(),
		Description:                  sdkutil.StrPtr(plan.Description),
		TriggerType:                  plan.TriggerType.ValueString(),
		TriggerConditions:            objectFromNormalized(plan.TriggerConditions),
		ServiceId:                    sdkutil.IntPtr(plan.ServiceID),
		SuppressDefaultNotifications: &suppress,
		Steps:                        stepsToSDK(plan.Steps),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating playbook", err.Error())
		return
	}

	resp.Diagnostics.Append(mapPlaybookToState(pb, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *playbookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PlaybookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pb, err := r.sdk.Playbooks.GetPlaybook(ctx, int(state.OrganizationID.ValueInt64()), int(state.ID.ValueInt64()))
	if err != nil {
		if sdkutil.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading playbook", err.Error())
		return
	}

	resp.Diagnostics.Append(mapPlaybookToState(pb, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *playbookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PlaybookModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	triggerType := plan.TriggerType.ValueString()
	suppress := plan.SuppressDefaultNotifications.ValueBool()
	pb, err := r.sdk.Playbooks.UpdatePlaybook(ctx, int(plan.OrganizationID.ValueInt64()), int(plan.ID.ValueInt64()), models.UpdatePlaybookJSONRequestBody{
		Name:                         &name,
		Description:                  sdkutil.StrPtr(plan.Description),
		TriggerType:                  &triggerType,
		TriggerConditions:            objectFromNormalized(plan.TriggerConditions),
		ServiceId:                    sdkutil.IntPtr(plan.ServiceID),
		SuppressDefaultNotifications: &suppress,
		Steps:                        stepsToSDK(plan.Steps),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating playbook", err.Error())
		return
	}

	resp.Diagnostics.Append(mapPlaybookToState(pb, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *playbookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PlaybookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.sdk.Playbooks.DeletePlaybook(ctx, int(state.OrganizationID.ValueInt64()), int(state.ID.ValueInt64())); err != nil {
		if sdkutil.IsNotFound(err) {
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

// stepResponse mirrors a single element of PlaybookResponse.Steps, which the SDK
// types as an opaque interface{}. config/conditions stay opaque JSON blobs so the
// provider round-trips them through jsontypes.Normalized exactly as before.
type stepResponse struct {
	StepOrder      int             `json:"step_order"`
	Name           string          `json:"name"`
	ActionType     string          `json:"action_type"`
	Config         json.RawMessage `json:"config"`
	Conditions     json.RawMessage `json:"conditions"`
	ParallelGroup  *string         `json:"parallel_group"`
	IsManual       bool            `json:"is_manual"`
	TimeoutSeconds int             `json:"timeout_seconds"`
}

// stepsToSDK converts the plan's nested step blocks into the SDK step schema. The
// config/conditions JSON blobs are decoded into the SDK's typed shapes.
func stepsToSDK(steps []PlaybookStepModel) *[]models.PlaybookStepSchema {
	out := make([]models.PlaybookStepSchema, len(steps))
	for i, s := range steps {
		isManual := s.IsManual.ValueBool()
		timeout := int(s.TimeoutSeconds.ValueInt64())
		out[i] = models.PlaybookStepSchema{
			StepOrder:      int(s.StepOrder.ValueInt64()),
			Name:           s.Name.ValueString(),
			ActionType:     s.ActionType.ValueString(),
			Config:         objectFromNormalized(s.Config),
			Conditions:     conditionsFromNormalized(s.Conditions),
			ParallelGroup:  sdkutil.StrPtr(s.ParallelGroup),
			IsManual:       &isManual,
			TimeoutSeconds: &timeout,
		}
	}
	return &out
}

func mapPlaybookToState(pb *models.PlaybookResponse, m *PlaybookModel) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.Int64Value(int64(pb.Id))
	m.OrganizationID = types.Int64Value(int64(pb.OrganizationId))
	m.Name = types.StringValue(pb.Name)
	if pb.Description != nil {
		m.Description = types.StringValue(*pb.Description)
	} else {
		m.Description = types.StringValue("")
	}
	m.TriggerType = types.StringValue(pb.TriggerType)
	m.TriggerConditions = normalizedFromRaw(rawFromMap(pb.TriggerConditions))
	m.ServiceID = sdkutil.Int(pb.ServiceId)
	m.SuppressDefaultNotifications = types.BoolValue(pb.SuppressDefaultNotifications)
	m.IsActive = types.BoolValue(pb.IsActive)

	var raw []stepResponse
	if len(pb.Steps) > 0 {
		b, err := json.Marshal(pb.Steps)
		if err != nil {
			diags.AddError("Error decoding playbook steps", err.Error())
			return diags
		}
		if err := json.Unmarshal(b, &raw); err != nil {
			diags.AddError("Error decoding playbook steps", err.Error())
			return diags
		}
	}

	steps := make([]PlaybookStepModel, len(raw))
	for i, s := range raw {
		steps[i] = PlaybookStepModel{
			StepOrder:      types.Int64Value(int64(s.StepOrder)),
			Name:           types.StringValue(s.Name),
			ActionType:     types.StringValue(s.ActionType),
			Config:         normalizedFromRaw(s.Config),
			Conditions:     normalizedFromRaw(s.Conditions),
			ParallelGroup:  sdkutil.Str(s.ParallelGroup),
			IsManual:       types.BoolValue(s.IsManual),
			TimeoutSeconds: types.Int64Value(int64(s.TimeoutSeconds)),
		}
	}
	m.Steps = steps
	return diags
}

// objectFromNormalized decodes a normalized JSON object into the SDK's map shape,
// nil when null or unknown.
func objectFromNormalized(n jsontypes.Normalized) *map[string]interface{} {
	if n.IsNull() || n.IsUnknown() {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(n.ValueString()), &m); err != nil {
		return nil
	}
	return &m
}

// conditionsFromNormalized decodes a normalized JSON array into the SDK's typed
// condition list, nil when null or unknown.
func conditionsFromNormalized(n jsontypes.Normalized) *[]models.PlaybookCondition {
	if n.IsNull() || n.IsUnknown() {
		return nil
	}
	var c []models.PlaybookCondition
	if err := json.Unmarshal([]byte(n.ValueString()), &c); err != nil {
		return nil
	}
	return &c
}

// rawFromMap re-marshals the SDK's map-typed trigger conditions back into a raw
// JSON blob so it can flow through jsontypes.Normalized unchanged.
func rawFromMap(m *map[string]interface{}) json.RawMessage {
	if m == nil {
		return nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return b
}

func normalizedFromRaw(raw json.RawMessage) jsontypes.Normalized {
	if len(raw) == 0 || string(raw) == "null" {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(string(raw))
}
