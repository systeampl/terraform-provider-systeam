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
	syschecks "github.com/systeampl/syschecks-go"
	"github.com/systeampl/syschecks-go/models"
	"github.com/systeampl/terraform-provider-systeam/internal/sdkutil"
)

var (
	_ resource.Resource                = &serviceResource{}
	_ resource.ResourceWithConfigure   = &serviceResource{}
	_ resource.ResourceWithImportState = &serviceResource{}
)

type serviceResource struct {
	sdk *syschecks.Client
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

	svc, err := r.sdk.Services.CreateService(ctx, int(plan.OrganizationID.ValueInt64()), models.CreateServiceJSONRequestBody{
		Name:                   plan.Name.ValueString(),
		Description:            sdkutil.StrPtr(plan.Description),
		RepoUrl:                sdkutil.StrPtr(plan.RepoURL),
		DocsUrl:                sdkutil.StrPtr(plan.DocsURL),
		OwnerTeamId:            optionalInt(plan.OwnerTeamID),
		EscalationPolicyId:     optionalInt(plan.EscalationPolicyID),
		Tier:                   sdkutil.StrPtr(plan.Tier),
		NotificationChannelIds: intSlicePtr(channelIDs),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating service", err.Error())
		return
	}

	resp.Diagnostics.Append(applyService(ctx, &plan, svc.Id, svc.OrganizationId, svc.Name, svc.Slug,
		svc.Description, svc.RepoUrl, svc.DocsUrl, svc.OwnerTeamId, svc.EscalationPolicyId, svc.Tier,
		svc.IsActive, svc.NotificationChannels)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc, err := r.sdk.Services.GetService(ctx, int(state.OrganizationID.ValueInt64()), int(state.ID.ValueInt64()))
	if err != nil {
		if sdkutil.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading service", err.Error())
		return
	}

	resp.Diagnostics.Append(applyService(ctx, &state, svc.Id, svc.OrganizationId, svc.Name, svc.Slug,
		svc.Description, svc.RepoUrl, svc.DocsUrl, svc.OwnerTeamId, svc.EscalationPolicyId, svc.Tier,
		svc.IsActive, svc.NotificationChannels)...)
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

	svc, err := r.sdk.Services.UpdateService(ctx, int(plan.OrganizationID.ValueInt64()), int(plan.ID.ValueInt64()), models.UpdateServiceJSONRequestBody{
		Name:                   sdkutil.StrPtr(plan.Name),
		Description:            sdkutil.StrPtr(plan.Description),
		RepoUrl:                sdkutil.StrPtr(plan.RepoURL),
		DocsUrl:                sdkutil.StrPtr(plan.DocsURL),
		OwnerTeamId:            optionalInt(plan.OwnerTeamID),
		EscalationPolicyId:     optionalInt(plan.EscalationPolicyID),
		Tier:                   sdkutil.StrPtr(plan.Tier),
		NotificationChannelIds: intSlicePtr(channelIDs),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating service", err.Error())
		return
	}

	resp.Diagnostics.Append(applyService(ctx, &plan, svc.Id, svc.OrganizationId, svc.Name, svc.Slug,
		svc.Description, svc.RepoUrl, svc.DocsUrl, svc.OwnerTeamId, svc.EscalationPolicyId, svc.Tier,
		svc.IsActive, svc.NotificationChannels)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.sdk.Services.DeleteService(ctx, int(state.OrganizationID.ValueInt64()), int(state.ID.ValueInt64())); err != nil {
		if sdkutil.IsNotFound(err) {
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

// intSlicePtr turns the plan's channel ids into a request pointer, nil when the
// set is unset/empty — matching the previous omitempty semantics on the body.
func intSlicePtr(ids []int) *[]int {
	if len(ids) == 0 {
		return nil
	}
	return &ids
}

func applyService(ctx context.Context, m *ServiceModel,
	id, orgID int, name, slug string,
	description, repoURL, docsURL *string,
	ownerTeamID, escalationPolicyID *int, tier string, isActive bool,
	channels *[]models.ServiceNotificationChannelBrief,
) diag.Diagnostics {
	m.ID = types.Int64Value(int64(id))
	m.OrganizationID = types.Int64Value(int64(orgID))
	m.Name = types.StringValue(name)
	m.Slug = types.StringValue(slug)
	m.Description = strOrEmpty(description)
	m.RepoURL = strOrEmpty(repoURL)
	m.DocsURL = strOrEmpty(docsURL)
	m.OwnerTeamID = nullableInt(ownerTeamID)
	m.EscalationPolicyID = nullableInt(escalationPolicyID)
	m.Tier = types.StringValue(tier)
	m.IsActive = types.BoolValue(isActive)

	elems := make([]int64, 0)
	if channels != nil {
		for _, c := range *channels {
			elems = append(elems, int64(c.Id))
		}
	}
	set, diags := types.SetValueFrom(ctx, types.Int64Type, elems)
	m.NotificationChannelIDs = set
	return diags
}

// strOrEmpty maps a response *string to a framework String, empty when nil —
// preserving the previous always-non-null behavior for these computed fields.
func strOrEmpty(p *string) types.String {
	if p == nil {
		return types.StringValue("")
	}
	return types.StringValue(*p)
}

func nullableInt(p *int) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*p))
}
