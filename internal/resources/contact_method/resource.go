package contact_method

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pawel-cygal/terraform-provider-systeam/internal/client"
)

var (
	_ resource.Resource                = &contactMethodResource{}
	_ resource.ResourceWithConfigure   = &contactMethodResource{}
	_ resource.ResourceWithImportState = &contactMethodResource{}
)

type contactMethodResource struct {
	client *client.Client
}

func NewResource() resource.Resource {
	return &contactMethodResource{}
}

func (r *contactMethodResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contact_method"
}

func (r *contactMethodResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a per-user contact method (phone/email/push) used by on-call escalation. " +
			"Tied to the provider token's user. Verification happens out of band (a code is sent to the " +
			"target) and cannot be automated — a freshly created method starts unverified.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:      true,
				Description:   "The unique identifier of the contact method.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"kind": schema.StringAttribute{
				Required:      true,
				Description:   "Contact channel: sms, voice, email, or push.",
				Validators:    []validator.String{stringvalidator.OneOf("sms", "voice", "email", "push")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"value": schema.StringAttribute{
				Optional:      true,
				Description:   "The phone number (sms/voice) or email address (email). Omit for push.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"label": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A human-readable label.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the method is enabled for notifications.",
			},
			"verified": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the method has been verified (read-only; verify out of band).",
			},
		},
	}
}

func (r *contactMethodResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *contactMethodResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ContactMethodModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cm, err := r.client.CreateContactMethod(ctx, client.ContactMethodCreateRequest{
		Kind:  plan.Kind.ValueString(),
		Value: optionalString(plan.Value),
		Label: optionalString(plan.Label),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating contact method", err.Error())
		return
	}

	mapContactMethodToState(cm, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *contactMethodResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ContactMethodModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cm, err := r.client.GetContactMethod(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading contact method", err.Error())
		return
	}
	if cm == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	mapContactMethodToState(cm, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *contactMethodResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ContactMethodModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled := plan.Enabled.ValueBool()
	cm, err := r.client.UpdateContactMethod(ctx, int(plan.ID.ValueInt64()), client.ContactMethodUpdateRequest{
		Label:   optionalString(plan.Label),
		Enabled: &enabled,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating contact method", err.Error())
		return
	}

	mapContactMethodToState(cm, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *contactMethodResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ContactMethodModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeleteContactMethod(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Error deleting contact method", err.Error())
	}
}

func (r *contactMethodResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected a numeric contact method id")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func mapContactMethodToState(cm *client.ContactMethod, m *ContactMethodModel) {
	m.ID = types.Int64Value(int64(cm.ID))
	m.Kind = types.StringValue(cm.Kind)
	m.Value = stringOrNull(cm.Value)
	m.Label = stringOrNull(cm.Label)
	m.Enabled = types.BoolValue(cm.Enabled)
	m.Verified = types.BoolValue(cm.Verified)
}

func optionalString(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func stringOrNull(p *string) types.String {
	if p == nil {
		return types.StringNull()
	}
	return types.StringValue(*p)
}
