package agent_registration_token

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource              = &agentRegistrationTokenResource{}
	_ resource.ResourceWithConfigure = &agentRegistrationTokenResource{}
)

type agentRegistrationTokenResource struct {
	client *client.Client
}

func NewResource() resource.Resource {
	return &agentRegistrationTokenResource{}
}

func (r *agentRegistrationTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_registration_token"
}

func (r *agentRegistrationTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// The backend only offers create — the token is a one-time, 1h, write-only
	// credential (no read/update/delete). It is therefore modeled like a
	// generated secret: minted once, held in state, and re-minted (replace) on
	// any input change. Feed `token` to the agent's `register` command (e.g. via
	// a local-exec/remote-exec provisioner, or hand it to Ansible/Salt).
	resp.Schema = schema.Schema{
		Description: "Mints a one-time token to enroll a private (or geo) agent. The agent trades " +
			"this token for a permanent key via `healthcheck-agent register`. Write-only and valid " +
			"for one hour; changing any input mints a new token.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Synthetic identifier ('<organization_id>:<name>').",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.Int64Attribute{
				Required:      true,
				Description:   "The organization the enrolled agent will belong to.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:      true,
				Description:   "Name for the agent that will be created when this token is used.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"mode": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Default:       stringdefault.StaticString("private"),
				Description:   "Agent mode: 'private' (customer host) or 'geo' (superadmin-only geo probe).",
				Validators:    []validator.String{stringvalidator.OneOf("private", "geo")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"token": schema.StringAttribute{
				Computed:      true,
				Sensitive:     true,
				Description:   "The one-time enrollment token. Pass to `healthcheck-agent register -token`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"expires_at": schema.StringAttribute{
				Computed:      true,
				Description:   "RFC3339 expiry (one hour after creation).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *agentRegistrationTokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *agentRegistrationTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AgentRegistrationTokenModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tok, err := r.client.CreateAgentRegistrationToken(ctx, client.AgentRegistrationTokenCreateRequest{
		Name:           plan.Name.ValueString(),
		Mode:           plan.Mode.ValueString(),
		OrganizationID: int(plan.OrganizationID.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating agent registration token", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d:%s", plan.OrganizationID.ValueInt64(), plan.Name.ValueString()))
	plan.Token = types.StringValue(tok.Token)
	plan.ExpiresAt = types.StringValue(tok.ExpiresAt)
	// The server echoes the mode; trust it so 'geo' resolutions stick.
	if tok.Mode != "" {
		plan.Mode = types.StringValue(tok.Mode)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *agentRegistrationTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// No read-back endpoint exists and the token is write-only. Keep state as-is;
	// the token may already be consumed or expired, but there is nothing to
	// reconcile against and re-minting on every plan would be worse.
	var state AgentRegistrationTokenModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *agentRegistrationTokenResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All inputs are RequiresReplace, so Terraform never calls Update.
	resp.Diagnostics.AddError(
		"Agent registration tokens cannot be updated",
		"Changing any input mints a new token (replace). This is a provider bug if you see it.",
	)
}

func (r *agentRegistrationTokenResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Nothing to delete server-side: the token expires after an hour and is
	// consumed on first use. Removing it from state is sufficient. Note this does
	// NOT delete an agent that already enrolled with the token — the backend has
	// no delete-agent endpoint yet.
}
