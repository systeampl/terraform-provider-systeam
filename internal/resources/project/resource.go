package project

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	syschecks "github.com/systeampl/syschecks-go"
	"github.com/systeampl/syschecks-go/models"
	"github.com/systeampl/terraform-provider-systeam/internal/sdkutil"
)

var (
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

type projectResource struct {
	sdk *syschecks.Client
}

func NewResource() resource.Resource {
	return &projectResource{}
}

func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a SysTeam Healthchecks project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier of the project.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the project.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A description of the project.",
			},
			"organization_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the organization this project belongs to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"access_control_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether access control is enabled for this project.",
			},
		},
	}
}

func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := int(plan.OrganizationID.ValueInt64())
	project, err := r.sdk.Organizations.CreateOrganizationProject(ctx, orgID, models.CreateOrganizationProjectJSONRequestBody{
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating project", err.Error())
		return
	}

	applyProject(&plan, project)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := int(state.OrganizationID.ValueInt64())
	projectID := int(state.ID.ValueInt64())

	project, err := r.sdk.Organizations.GetOrganizationProject(ctx, orgID, projectID)
	if err != nil {
		if sdkutil.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}

	applyProject(&state, project)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := int(plan.OrganizationID.ValueInt64())
	projectID := int(plan.ID.ValueInt64())

	name := plan.Name.ValueString()
	accessControl := plan.AccessControlEnabled.ValueBool()
	project, err := r.sdk.Organizations.UpdateOrganizationProject(ctx, orgID, projectID, models.UpdateOrganizationProjectJSONRequestBody{
		Name:                 &name,
		Description:          sdkutil.StrPtr(plan.Description),
		AccessControlEnabled: &accessControl,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}

	applyProject(&plan, project)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := int(state.OrganizationID.ValueInt64())
	projectID := int(state.ID.ValueInt64())

	if _, err := r.sdk.Organizations.DeleteOrganizationProject(ctx, orgID, projectID); err != nil {
		if sdkutil.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting project", err.Error())
	}
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected import ID format: org_id:project_id",
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

	projectID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid project_id in import ID",
			fmt.Sprintf("Could not parse project_id '%s': %s", parts[1], err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), projectID)...)
}

func applyProject(m *ProjectModel, p *models.ProjectResponse) {
	m.ID = types.Int64Value(int64(p.Id))
	m.Name = types.StringValue(p.Name)
	m.Description = sdkutil.Str(p.Description)
	if p.OrganizationId != nil {
		m.OrganizationID = types.Int64Value(int64(*p.OrganizationId))
	}
	if p.AccessControlEnabled != nil {
		m.AccessControlEnabled = types.BoolValue(*p.AccessControlEnabled)
	}
}
