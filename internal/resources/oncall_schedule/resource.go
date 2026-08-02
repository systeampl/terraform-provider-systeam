package oncall_schedule

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	syschecks "github.com/systeampl/syschecks-go"
	"github.com/systeampl/syschecks-go/models"
	"github.com/systeampl/terraform-provider-systeam/internal/sdkutil"
)

var (
	_ resource.Resource                = &oncallScheduleResource{}
	_ resource.ResourceWithConfigure   = &oncallScheduleResource{}
	_ resource.ResourceWithImportState = &oncallScheduleResource{}
)

type oncallScheduleResource struct {
	sdk *syschecks.Client
}

func NewResource() resource.Resource {
	return &oncallScheduleResource{}
}

func (r *oncallScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oncall_schedule"
}

func (r *oncallScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an on-call schedule with participants.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier of the on-call schedule.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the on-call schedule.",
			},
			"organization_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the organization this schedule belongs to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"timezone": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("UTC"),
				Description: "Timezone for the schedule (e.g. 'Europe/Warsaw', 'UTC').",
			},
			"rotation_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("WEEKLY"),
				Description: "Rotation type: DAILY or WEEKLY.",
				Validators: []validator.String{
					stringvalidator.OneOf("DAILY", "WEEKLY"),
				},
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the on-call schedule is active.",
			},
			"participant": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Participants in the on-call rotation.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_id": schema.Int64Attribute{
							Required:    true,
							Description: "ID of the participant user.",
						},
						"position": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Default:     int64default.StaticInt64(0),
							Description: "Position in the rotation order.",
						},
					},
				},
			},
		},
	}
}

func (r *oncallScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *oncallScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OnCallScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timezone := plan.Timezone.ValueString()
	rotationType := plan.RotationType.ValueString()
	schedule, err := r.sdk.Oncall.CreateSchedule(ctx, int(plan.OrganizationID.ValueInt64()), models.CreateScheduleJSONRequestBody{
		Name:         plan.Name.ValueString(),
		Timezone:     &timezone,
		RotationType: &rotationType,
		Participants: participantsToSDK(plan.Participants),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating on-call schedule", err.Error())
		return
	}

	mapScheduleToState(schedule, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *oncallScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OnCallScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schedule, err := r.sdk.Oncall.GetSchedule(ctx, int(state.OrganizationID.ValueInt64()), int(state.ID.ValueInt64()))
	if err != nil {
		if sdkutil.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading on-call schedule", err.Error())
		return
	}

	mapScheduleToState(schedule, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *oncallScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OnCallScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	timezone := plan.Timezone.ValueString()
	rotationType := plan.RotationType.ValueString()
	isActive := plan.IsActive.ValueBool()
	schedule, err := r.sdk.Oncall.UpdateSchedule(ctx, int(plan.OrganizationID.ValueInt64()), int(plan.ID.ValueInt64()), models.UpdateScheduleJSONRequestBody{
		Name:         &name,
		Timezone:     &timezone,
		RotationType: &rotationType,
		IsActive:     &isActive,
		Participants: participantsToSDK(plan.Participants),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating on-call schedule", err.Error())
		return
	}

	mapScheduleToState(schedule, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *oncallScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OnCallScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.sdk.Oncall.DeleteSchedule(ctx, int(state.OrganizationID.ValueInt64()), int(state.ID.ValueInt64())); err != nil {
		if sdkutil.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting on-call schedule", err.Error())
	}
}

func (r *oncallScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected import ID format: org_id:schedule_id",
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

	scheduleID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid schedule_id in import ID",
			fmt.Sprintf("Could not parse schedule_id '%s': %s", parts[1], err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), scheduleID)...)
}

func participantsToSDK(participants []OnCallParticipantModel) *[]models.ParticipantSchema {
	if len(participants) == 0 {
		return nil
	}
	out := make([]models.ParticipantSchema, len(participants))
	for i, p := range participants {
		pos := int(p.Position.ValueInt64())
		out[i] = models.ParticipantSchema{
			UserId:   int(p.UserID.ValueInt64()),
			Position: &pos,
		}
	}
	return &out
}

func mapScheduleToState(schedule *models.OncallScheduleResponse, state *OnCallScheduleModel) {
	state.ID = types.Int64Value(int64(schedule.Id))
	state.Name = types.StringValue(schedule.Name)
	state.OrganizationID = types.Int64Value(int64(schedule.OrganizationId))
	state.Timezone = types.StringValue(schedule.Timezone)
	state.RotationType = types.StringValue(schedule.RotationType)
	state.IsActive = types.BoolValue(schedule.IsActive)

	if schedule.Participants != nil && len(*schedule.Participants) > 0 {
		participants := make([]OnCallParticipantModel, len(*schedule.Participants))
		for i, p := range *schedule.Participants {
			participants[i] = OnCallParticipantModel{
				UserID:   types.Int64Value(int64(p.UserId)),
				Position: types.Int64Value(int64(p.Position)),
			}
		}
		state.Participants = participants
	} else {
		state.Participants = nil
	}
}
