package team

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	syschecks "github.com/systeampl/syschecks-go"
	"github.com/systeampl/syschecks-go/models"
	"github.com/systeampl/terraform-provider-systeam/internal/sdkutil"
)

var (
	_ resource.Resource                = &teamResource{}
	_ resource.ResourceWithConfigure   = &teamResource{}
	_ resource.ResourceWithImportState = &teamResource{}
)

type teamResource struct {
	sdk *syschecks.Client
}

func NewResource() resource.Resource {
	return &teamResource{}
}

func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a team: an org-scoped grouping used for ownership and routing. Projects, " +
			"on-call schedules, escalation policies and integration keys can belong to a team.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:      true,
				Description:   "The unique identifier of the team.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.Int64Attribute{
				Required:      true,
				Description:   "The ID of the organization this team belongs to.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The team's display name.",
			},
			"slug": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "URL-safe identifier. Generated from the name when omitted; cannot be changed afterwards.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "An optional description of the team.",
			},
			"is_active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the team is active.",
			},
			"member_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of members in the team.",
			},
		},
	}
}

func (r *teamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TeamModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := r.sdk.Teams.CreateTeam(ctx, int(plan.OrganizationID.ValueInt64()), models.CreateTeamJSONRequestBody{
		Name:        plan.Name.ValueString(),
		Slug:        sdkutil.StrPtr(plan.Slug),
		Description: sdkutil.StrPtr(plan.Description),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating team", err.Error())
		return
	}

	applyTeam(&plan, team.Id, team.OrganizationId, team.Name, team.Slug, team.Description, team.IsActive, team.MemberCount)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TeamModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := r.sdk.Teams.GetTeam(ctx, int(state.OrganizationID.ValueInt64()), int(state.ID.ValueInt64()))
	if err != nil {
		if sdkutil.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading team", err.Error())
		return
	}

	applyTeam(&state, team.Id, team.OrganizationId, team.Name, team.Slug, team.Description, team.IsActive, team.MemberCount)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TeamModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	team, err := r.sdk.Teams.UpdateTeam(ctx, int(plan.OrganizationID.ValueInt64()), int(plan.ID.ValueInt64()), models.UpdateTeamJSONRequestBody{
		Name:        &name,
		Description: sdkutil.StrPtr(plan.Description),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating team", err.Error())
		return
	}

	applyTeam(&plan, team.Id, team.OrganizationId, team.Name, team.Slug, team.Description, team.IsActive, team.MemberCount)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TeamModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.sdk.Teams.DeleteTeam(ctx, int(state.OrganizationID.ValueInt64()), int(state.ID.ValueInt64())); err != nil {
		if sdkutil.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting team", err.Error())
	}
}

func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected import ID format: org_id:team_id")
		return
	}
	orgID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid org_id in import ID", err.Error())
		return
	}
	teamID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid team_id in import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), teamID)...)
}

func applyTeam(m *TeamModel, id, orgID int, name, slug string, description *string, isActive bool, memberCount int) {
	m.ID = types.Int64Value(int64(id))
	m.OrganizationID = types.Int64Value(int64(orgID))
	m.Name = types.StringValue(name)
	m.Slug = types.StringValue(slug)
	m.Description = sdkutil.Str(description)
	m.IsActive = types.BoolValue(isActive)
	m.MemberCount = types.Int64Value(int64(memberCount))
}
