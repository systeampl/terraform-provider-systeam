package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/systeampl/terraform-provider-systeam/internal/client"
)

var (
	_ resource.Resource                = &serviceResource{}
	_ resource.ResourceWithConfigure   = &serviceResource{}
	_ resource.ResourceWithImportState = &serviceResource{}
)

type serviceResource struct {
	client *client.Client
}

func NewResource() resource.Resource {
	return &serviceResource{}
}

func (r *serviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (r *serviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a service: an org-scoped catalog entry that incidents and checks attach to, " +
			"with an owning team, an escalation policy, and notification channels.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:      true,
				Description:   "The unique identifier of the service.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.Int64Attribute{
				Required:      true,
				Description:   "The ID of the organization this service belongs to.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The service's display name.",
			},
			"slug": schema.StringAttribute{
				Computed:      true,
				Description:   "URL-safe identifier, generated from the name.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "An optional description.",
			},
			"repo_url": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Link to the service's source repository.",
			},
			"docs_url": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Link to the service's documentation.",
			},
			"owner_team_id": schema.Int64Attribute{
				Optional:    true,
				Description: "The team that owns this service.",
			},
			"escalation_policy_id": schema.Int64Attribute{
				Optional:    true,
				Description: "The escalation policy incidents on this service route to.",
			},
			"tier": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("P3"),
				Description: "Criticality tier: P1 (highest) through P4.",
				Validators:  []validator.String{stringvalidator.OneOf("P1", "P2", "P3", "P4")},
			},
			"notification_channel_ids": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.Int64Type,
				Description: "Notification channels attached to this service.",
			},
			"is_active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the service is active.",
			},
		},
	}
}

func (r *serviceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServiceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	channelIDs, diags := setToIntSlice(ctx, plan.NotificationChannelIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc, err := r.client.CreateService(ctx, int(plan.OrganizationID.ValueInt64()), client.ServiceCreateRequest{
		Name:                   plan.Name.ValueString(),
		Description:            plan.Description.ValueString(),
		RepoURL:                plan.RepoURL.ValueString(),
		DocsURL:                plan.DocsURL.ValueString(),
		OwnerTeamID:            optionalInt(plan.OwnerTeamID),
		EscalationPolicyID:     optionalInt(plan.EscalationPolicyID),
		Tier:                   plan.Tier.ValueString(),
		NotificationChannelIDs: channelIDs,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating service", err.Error())
		return
	}

	resp.Diagnostics.Append(mapServiceToState(ctx, svc, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc, err := r.client.GetService(ctx, int(state.OrganizationID.ValueInt64()), int(state.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading service", err.Error())
		return
	}

	resp.Diagnostics.Append(mapServiceToState(ctx, svc, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ServiceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	channelIDs, diags := setToIntSlice(ctx, plan.NotificationChannelIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc, err := r.client.UpdateService(ctx, int(plan.OrganizationID.ValueInt64()), int(plan.ID.ValueInt64()), client.ServiceUpdateRequest{
		Name:                   plan.Name.ValueString(),
		Description:            plan.Description.ValueString(),
		RepoURL:                plan.RepoURL.ValueString(),
		DocsURL:                plan.DocsURL.ValueString(),
		OwnerTeamID:            optionalInt(plan.OwnerTeamID),
		EscalationPolicyID:     optionalInt(plan.EscalationPolicyID),
		Tier:                   plan.Tier.ValueString(),
		NotificationChannelIDs: channelIDs,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating service", err.Error())
		return
	}

	resp.Diagnostics.Append(mapServiceToState(ctx, svc, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeleteService(ctx, int(state.OrganizationID.ValueInt64()), int(state.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Error deleting service", err.Error())
	}
}

func (r *serviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected import ID format: org_id:service_id")
		return
	}
	orgID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid org_id in import ID", err.Error())
		return
	}
	serviceID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid service_id in import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), serviceID)...)
}

// optionalInt turns a nullable model int into a request pointer (nil when unset).
func optionalInt(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := int(v.ValueInt64())
	return &i
}

func setToIntSlice(ctx context.Context, set types.Set) ([]int, diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return nil, nil
	}
	var ids []int64
	diags := set.ElementsAs(ctx, &ids, false)
	out := make([]int, len(ids))
	for i, v := range ids {
		out[i] = int(v)
	}
	return out, diags
}

func mapServiceToState(ctx context.Context, svc *client.Service, m *ServiceModel) diag.Diagnostics {
	m.ID = types.Int64Value(int64(svc.ID))
	m.OrganizationID = types.Int64Value(int64(svc.OrganizationID))
	m.Name = types.StringValue(svc.Name)
	m.Slug = types.StringValue(svc.Slug)
	m.Description = types.StringValue(svc.Description)
	m.RepoURL = types.StringValue(svc.RepoURL)
	m.DocsURL = types.StringValue(svc.DocsURL)
	m.OwnerTeamID = nullableInt(svc.OwnerTeamID)
	m.EscalationPolicyID = nullableInt(svc.EscalationPolicyID)
	m.Tier = types.StringValue(svc.Tier)
	m.IsActive = types.BoolValue(svc.IsActive)

	ids := svc.NotificationChannelIDs()
	elems := make([]int64, len(ids))
	for i, v := range ids {
		elems[i] = int64(v)
	}
	set, diags := types.SetValueFrom(ctx, types.Int64Type, elems)
	m.NotificationChannelIDs = set
	return diags
}

func nullableInt(p *int) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*p))
}
