package status_page

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pawel-cygal/terraform-provider-systeam/internal/client"
)

var (
	_ resource.Resource                = &statusPageResource{}
	_ resource.ResourceWithConfigure   = &statusPageResource{}
	_ resource.ResourceWithImportState = &statusPageResource{}
)

type statusPageResource struct {
	client *client.Client
}

func NewResource() resource.Resource {
	return &statusPageResource{}
}

func (r *statusPageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status_page"
}

func (r *statusPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a public status page.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Status page ID.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Status page name.",
			},
			"slug": schema.StringAttribute{
				Required:    true,
				Description: "URL-friendly identifier for the status page.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Status page description.",
			},
			"is_public": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the status page is publicly accessible.",
			},
			"custom_domain": schema.StringAttribute{
				Optional:    true,
				Description: "Custom domain for the status page.",
			},
			"logo_url": schema.StringAttribute{
				Optional:    true,
				Description: "URL of the logo to display on the status page.",
			},
			"check_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.Int64Type,
				Description: "List of check IDs to display on the status page.",
			},
			"organization_id": schema.Int64Attribute{
				Required:    true,
				Description: "Organization ID that owns this status page.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the status page is active.",
			},
		},
	}
}

func (r *statusPageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *statusPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StatusPageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var checkIDs []int
	resp.Diagnostics.Append(plan.CheckIDs.ElementsAs(ctx, &checkIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.StatusPageCreateRequest{
		Name:     plan.Name.ValueString(),
		Slug:     plan.Slug.ValueString(),
		IsPublic: plan.IsPublic.ValueBool(),
		CheckIDs: checkIDs,
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		createReq.Description = plan.Description.ValueString()
	}
	if !plan.CustomDomain.IsNull() && !plan.CustomDomain.IsUnknown() {
		createReq.CustomDomain = plan.CustomDomain.ValueString()
	}
	if !plan.LogoURL.IsNull() && !plan.LogoURL.IsUnknown() {
		createReq.LogoURL = plan.LogoURL.ValueString()
	}

	page, err := r.client.CreateStatusPage(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating status page", err.Error())
		return
	}

	mapStatusPageToState(ctx, page, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *statusPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StatusPageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, err := r.client.GetStatusPage(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading status page", err.Error())
		return
	}

	mapStatusPageToState(ctx, page, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *statusPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StatusPageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var checkIDs []int
	resp.Diagnostics.Append(plan.CheckIDs.ElementsAs(ctx, &checkIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	slug := plan.Slug.ValueString()
	isPublic := plan.IsPublic.ValueBool()
	isActive := plan.IsActive.ValueBool()

	updateReq := client.StatusPageUpdateRequest{
		Name:     &name,
		Slug:     &slug,
		IsPublic: &isPublic,
		IsActive: &isActive,
		CheckIDs: checkIDs,
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		updateReq.Description = &v
	}
	if !plan.CustomDomain.IsNull() && !plan.CustomDomain.IsUnknown() {
		v := plan.CustomDomain.ValueString()
		updateReq.CustomDomain = &v
	}
	if !plan.LogoURL.IsNull() && !plan.LogoURL.IsUnknown() {
		v := plan.LogoURL.ValueString()
		updateReq.LogoURL = &v
	}

	page, err := r.client.UpdateStatusPage(ctx, int(plan.ID.ValueInt64()), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating status page", err.Error())
		return
	}

	mapStatusPageToState(ctx, page, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *statusPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StatusPageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteStatusPage(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Error deleting status page", err.Error())
	}
}

func (r *statusPageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected integer status page ID, got: %s", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func mapStatusPageToState(ctx context.Context, page *client.StatusPage, state *StatusPageModel, diags *diag.Diagnostics) {
	state.ID = types.Int64Value(int64(page.ID))
	state.Name = types.StringValue(page.Name)
	state.Slug = types.StringValue(page.Slug)
	state.IsPublic = types.BoolValue(page.IsPublic)
	state.IsActive = types.BoolValue(page.IsActive)
	state.OrganizationID = types.Int64Value(int64(page.OrganizationID))

	if page.Description != "" {
		state.Description = types.StringValue(page.Description)
	} else if state.Description.IsNull() {
		state.Description = types.StringNull()
	}

	if page.CustomDomain != "" {
		state.CustomDomain = types.StringValue(page.CustomDomain)
	} else if state.CustomDomain.IsNull() {
		state.CustomDomain = types.StringNull()
	}

	if page.LogoURL != "" {
		state.LogoURL = types.StringValue(page.LogoURL)
	} else if state.LogoURL.IsNull() {
		state.LogoURL = types.StringNull()
	}

	checkIDValues := make([]types.Int64, len(page.CheckIDs))
	for i, id := range page.CheckIDs {
		checkIDValues[i] = types.Int64Value(int64(id))
	}
	listVal, d := types.ListValueFrom(ctx, types.Int64Type, checkIDValues)
	diags.Append(d...)
	state.CheckIDs = listVal
}
