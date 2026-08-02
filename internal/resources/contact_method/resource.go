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
	syschecks "github.com/systeampl/syschecks-go"
	"github.com/systeampl/syschecks-go/models"
	"github.com/systeampl/terraform-provider-systeam/internal/sdkutil"
)

var (
	_ resource.Resource                = &contactMethodResource{}
	_ resource.ResourceWithConfigure   = &contactMethodResource{}
	_ resource.ResourceWithImportState = &contactMethodResource{}
)

type contactMethodResource struct {
	sdk *syschecks.Client
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

func (r *contactMethodResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ContactMethodModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cm, err := r.sdk.ContactMethods.CreateContactMethod(ctx, models.CreateContactMethodJSONRequestBody{
		Kind:  plan.Kind.ValueString(),
		Value: sdkutil.StrPtr(plan.Value),
		Label: sdkutil.StrPtr(plan.Label),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating contact method", err.Error())
		return
	}

	// The create endpoint ignores enabled (it always starts enabled=true) and
	// applies no server-side label defaulting we can rely on. If the plan asks
	// for enabled=false, reconcile it with the update endpoint so create honors
	// the desired state.
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() && plan.Enabled.ValueBool() != cm.Enabled {
		enabled := plan.Enabled.ValueBool()
		updated, uerr := r.sdk.ContactMethods.UpdateContactMethod(ctx, cm.Id, models.UpdateContactMethodJSONRequestBody{
			Label:   sdkutil.StrPtr(plan.Label),
			Enabled: &enabled,
		})
		if uerr != nil {
			resp.Diagnostics.AddError("Error setting contact method enabled state", uerr.Error())
			return
		}
		cm = updated
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

	// No GET-by-id endpoint: list and find by id.
	methods, err := r.sdk.ContactMethods.ListContactMethods(ctx)
	if err != nil {
		if sdkutil.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading contact method", err.Error())
		return
	}

	id := int(state.ID.ValueInt64())
	var cm *models.ContactMethodResponse
	if methods != nil {
		for i := range *methods {
			if (*methods)[i].Id == id {
				cm = &(*methods)[i]
				break
			}
		}
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
	cm, err := r.sdk.ContactMethods.UpdateContactMethod(ctx, int(plan.ID.ValueInt64()), models.UpdateContactMethodJSONRequestBody{
		Label:   sdkutil.StrPtr(plan.Label),
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
	if _, err := r.sdk.ContactMethods.DeleteContactMethod(ctx, int(state.ID.ValueInt64())); err != nil {
		if sdkutil.IsNotFound(err) {
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

func mapContactMethodToState(cm *models.ContactMethodResponse, m *ContactMethodModel) {
	m.ID = types.Int64Value(int64(cm.Id))
	m.Kind = types.StringValue(cm.Kind)
	// The API masks value in every response (e.g. "e•••@example.com"), so it can
	// never round-trip. value is RequiresReplace, so the plan/prior state already
	// holds the authoritative secret — leave m.Value untouched rather than
	// clobbering it with the mask (which would break plan consistency).
	m.Label = sdkutil.Str(cm.Label)
	m.Enabled = types.BoolValue(cm.Enabled)
	m.Verified = types.BoolValue(cm.Verified)
}
