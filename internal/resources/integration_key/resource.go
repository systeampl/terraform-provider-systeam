package integration_key

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/systeampl/terraform-provider-systeam/internal/client"
)

var (
	_ resource.Resource                = &integrationKeyResource{}
	_ resource.ResourceWithConfigure   = &integrationKeyResource{}
	_ resource.ResourceWithImportState = &integrationKeyResource{}
)

type integrationKeyResource struct {
	client *client.Client
}

func NewResource() resource.Resource {
	return &integrationKeyResource{}
}

func (r *integrationKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_key"
}

func (r *integrationKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// The API has no update endpoint for integration keys, so every input is
	// RequiresReplace: a change destroys the key and mints a new one (a new
	// secret token). The token is returned only once, at creation.
	resp.Schema = schema.Schema{
		Description: "Manages an inbound-events integration key. External systems (Alertmanager, " +
			"Grafana, Prometheus, PagerDuty-format senders) authenticate with this key to raise " +
			"incidents that route to an escalation policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:      true,
				Description:   "The unique identifier of the integration key.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.Int64Attribute{
				Required:      true,
				Description:   "The ID of the organization this key belongs to.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:      true,
				Description:   "A human-readable name for the key.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"escalation_policy_id": schema.Int64Attribute{
				Required:      true,
				Description:   "The escalation policy that inbound events on this key are routed to.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"grouping_type": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Default:       stringdefault.StaticString("none"),
				Description:   "How incoming alerts are grouped into incidents: none, time_window, service, or intelligent.",
				Validators:    []validator.String{stringvalidator.OneOf("none", "time_window", "service", "intelligent")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"grouping_window_seconds": schema.Int64Attribute{
				Optional:      true,
				Computed:      true,
				Default:       int64default.StaticInt64(300),
				Description:   "The grouping window in seconds (used when grouping_type is not 'none').",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"token_prefix": schema.StringAttribute{
				Computed:      true,
				Description:   "The non-secret prefix of the key, shown in the UI to identify it.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"token": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The full secret routing key. Returned ONLY at creation; put it in your " +
					"Alertmanager/Grafana webhook URL. It is not readable afterwards, so it is preserved " +
					"in state from the create.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"is_active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the key is active (revoked keys are inactive).",
			},
		},
	}
}

func (r *integrationKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *integrationKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IntegrationKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := int(plan.OrganizationID.ValueInt64())
	key, err := r.client.CreateIntegrationKey(ctx, orgID, client.IntegrationKeyCreateRequest{
		Name:                  plan.Name.ValueString(),
		EscalationPolicyID:    int(plan.EscalationPolicyID.ValueInt64()),
		GroupingType:          plan.GroupingType.ValueString(),
		GroupingWindowSeconds: int(plan.GroupingWindowSeconds.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating integration key", err.Error())
		return
	}

	// The raw token is only here, in the create response — capture it into state.
	plan.Token = types.StringValue(key.Token)
	mapKeyToState(key, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *integrationKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IntegrationKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := int(state.OrganizationID.ValueInt64())
	keyID := int(state.ID.ValueInt64())

	key, err := r.client.GetIntegrationKey(ctx, orgID, keyID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading integration key", err.Error())
		return
	}
	if key == nil {
		// Revoked/deleted out of band — drop it so a plan recreates it.
		resp.State.RemoveResource(ctx)
		return
	}

	// The token is never returned by list/read; keep whatever the create stored.
	mapKeyToState(key, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *integrationKeyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Every attribute is RequiresReplace, so Terraform never calls Update. If it
	// somehow does, fail loudly rather than silently drift.
	resp.Diagnostics.AddError(
		"Integration keys cannot be updated in place",
		"The API has no update endpoint for integration keys; changes replace the key. "+
			"This is a provider bug if you see it — please report it.",
	)
}

func (r *integrationKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IntegrationKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeleteIntegrationKey(ctx, int(state.OrganizationID.ValueInt64()), int(state.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Error deleting integration key", err.Error())
	}
}

func (r *integrationKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected import ID format: org_id:key_id")
		return
	}
	orgID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid org_id in import ID", err.Error())
		return
	}
	keyID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid key_id in import ID", err.Error())
		return
	}
	// The token can't be recovered on import (shown once at creation); it stays null.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), keyID)...)
}

func mapKeyToState(key *client.IntegrationKey, m *IntegrationKeyModel) {
	m.ID = types.Int64Value(int64(key.ID))
	m.Name = types.StringValue(key.Name)
	m.EscalationPolicyID = types.Int64Value(int64(key.EscalationPolicyID))
	m.GroupingType = types.StringValue(key.GroupingType)
	m.GroupingWindowSeconds = types.Int64Value(int64(key.GroupingWindowSeconds))
	m.TokenPrefix = types.StringValue(key.TokenPrefix)
	m.IsActive = types.BoolValue(key.IsActive)
}
