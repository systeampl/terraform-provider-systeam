package escalation_policy

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

var (
	_ resource.Resource                = &escalationPolicyResource{}
	_ resource.ResourceWithConfigure   = &escalationPolicyResource{}
	_ resource.ResourceWithImportState = &escalationPolicyResource{}
)

type escalationPolicyResource struct {
	client *client.Client
}

func NewResource() resource.Resource {
	return &escalationPolicyResource{}
}

func (r *escalationPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_escalation_policy"
}

func (r *escalationPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an escalation policy with ordered steps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier of the escalation policy.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the escalation policy.",
			},
			"organization_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the organization this policy belongs to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the escalation policy is active.",
			},
			"step": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Ordered escalation steps.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"step_order": schema.Int64Attribute{
							Required:    true,
							Description: "Order of this step in the escalation chain (1-based).",
						},
						"delay_minutes": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Default:     int64default.StaticInt64(0),
							Description: "Delay in minutes before escalating to this step.",
						},
						"target_type": schema.StringAttribute{
							Required:    true,
							Description: "Type of escalation target: user, schedule, or channel.",
							Validators: []validator.String{
								stringvalidator.OneOf("user", "schedule", "channel"),
							},
						},
						"target_user_id": schema.Int64Attribute{
							Optional:    true,
							Description: "ID of the target user (when target_type is 'user').",
						},
						"target_schedule_id": schema.Int64Attribute{
							Optional:    true,
							Description: "ID of the target on-call schedule (when target_type is 'schedule').",
						},
						"target_channel_id": schema.Int64Attribute{
							Optional:    true,
							Description: "ID of the target notification channel (when target_type is 'channel').",
						},
					},
				},
			},
		},
	}
}

func (r *escalationPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *escalationPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EscalationPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := int(plan.OrganizationID.ValueInt64())
	createReq := client.EscalationPolicyCreateRequest{
		Name:     plan.Name.ValueString(),
		IsActive: plan.IsActive.ValueBool(),
		Steps:    stepsFromModel(plan.Steps),
	}

	policy, err := r.client.CreateEscalationPolicy(ctx, orgID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating escalation policy", err.Error())
		return
	}

	mapPolicyToState(policy, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *escalationPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EscalationPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := int(state.OrganizationID.ValueInt64())
	policyID := int(state.ID.ValueInt64())

	policy, err := r.client.GetEscalationPolicy(ctx, orgID, policyID)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading escalation policy", err.Error())
		return
	}

	mapPolicyToState(policy, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *escalationPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EscalationPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := int(plan.OrganizationID.ValueInt64())
	policyID := int(plan.ID.ValueInt64())

	name := plan.Name.ValueString()
	isActive := plan.IsActive.ValueBool()
	updateReq := client.EscalationPolicyUpdateRequest{
		Name:     &name,
		IsActive: &isActive,
		Steps:    stepsFromModel(plan.Steps),
	}

	policy, err := r.client.UpdateEscalationPolicy(ctx, orgID, policyID, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating escalation policy", err.Error())
		return
	}

	mapPolicyToState(policy, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *escalationPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EscalationPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := int(state.OrganizationID.ValueInt64())
	policyID := int(state.ID.ValueInt64())

	err := r.client.DeleteEscalationPolicy(ctx, orgID, policyID)
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Error deleting escalation policy", err.Error())
	}
}

func (r *escalationPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected import ID format: org_id:policy_id",
		)
		return
	}

	orgID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid org_id in import ID",
			fmt.Sprintf("Could not parse org_id '%s': %s", parts[0], err),
		)
		return
	}

	policyID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid policy_id in import ID",
			fmt.Sprintf("Could not parse policy_id '%s': %s", parts[1], err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), policyID)...)
}

func stepsFromModel(steps []EscalationStepModel) []client.EscalationStep {
	if steps == nil {
		return nil
	}
	result := make([]client.EscalationStep, len(steps))
	for i, s := range steps {
		step := client.EscalationStep{
			StepOrder:    int(s.StepOrder.ValueInt64()),
			DelayMinutes: int(s.DelayMinutes.ValueInt64()),
			TargetType:   s.TargetType.ValueString(),
		}
		if !s.TargetUserID.IsNull() && !s.TargetUserID.IsUnknown() {
			v := int(s.TargetUserID.ValueInt64())
			step.TargetUserID = &v
		}
		if !s.TargetScheduleID.IsNull() && !s.TargetScheduleID.IsUnknown() {
			v := int(s.TargetScheduleID.ValueInt64())
			step.TargetScheduleID = &v
		}
		if !s.TargetChannelID.IsNull() && !s.TargetChannelID.IsUnknown() {
			v := int(s.TargetChannelID.ValueInt64())
			step.TargetChannelID = &v
		}
		result[i] = step
	}
	return result
}

func mapPolicyToState(policy *client.EscalationPolicy, state *EscalationPolicyModel) {
	state.ID = types.Int64Value(int64(policy.ID))
	state.Name = types.StringValue(policy.Name)
	state.OrganizationID = types.Int64Value(int64(policy.OrganizationID))
	state.IsActive = types.BoolValue(policy.IsActive)

	if len(policy.Steps) > 0 {
		steps := make([]EscalationStepModel, len(policy.Steps))
		for i, s := range policy.Steps {
			steps[i] = EscalationStepModel{
				StepOrder:    types.Int64Value(int64(s.StepOrder)),
				DelayMinutes: types.Int64Value(int64(s.DelayMinutes)),
				TargetType:   types.StringValue(s.TargetType),
			}
			if s.TargetUserID != nil {
				steps[i].TargetUserID = types.Int64Value(int64(*s.TargetUserID))
			} else {
				steps[i].TargetUserID = types.Int64Null()
			}
			if s.TargetScheduleID != nil {
				steps[i].TargetScheduleID = types.Int64Value(int64(*s.TargetScheduleID))
			} else {
				steps[i].TargetScheduleID = types.Int64Null()
			}
			if s.TargetChannelID != nil {
				steps[i].TargetChannelID = types.Int64Value(int64(*s.TargetChannelID))
			} else {
				steps[i].TargetChannelID = types.Int64Null()
			}
		}
		state.Steps = steps
	} else {
		state.Steps = nil
	}
}
