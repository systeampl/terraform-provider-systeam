package notification_channel

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource                = &notificationChannelResource{}
	_ resource.ResourceWithConfigure   = &notificationChannelResource{}
	_ resource.ResourceWithImportState = &notificationChannelResource{}
)

type notificationChannelResource struct {
	client *client.Client
}

func NewResource() resource.Resource {
	return &notificationChannelResource{}
}

func (r *notificationChannelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_channel"
}

func (r *notificationChannelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a notification channel.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Notification channel ID.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Notification channel name.",
			},
			"channel_type": schema.StringAttribute{
				Required:    true,
				Description: "Channel type.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"email", "slack", "discord", "webhook", "telegram",
						"pagerduty", "msteams", "opsgenie", "jira", "whatsapp", "sms",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config": schema.MapAttribute{
				Required:    true,
				Description: "Channel configuration as key-value pairs. Keys and values depend on the channel_type.",
				ElementType: types.StringType,
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the channel is active.",
			},
			"organization_id": schema.Int64Attribute{
				Optional:    true,
				Description: "Organization ID that owns this channel.",
			},
		},
	}
}

func (r *notificationChannelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *notificationChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NotificationChannelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configMap := make(map[string]string)
	resp.Diagnostics.Append(plan.Config.ElementsAs(ctx, &configMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiConfig := make(map[string]any, len(configMap))
	for k, v := range configMap {
		apiConfig[k] = v
	}

	createReq := client.NotificationChannelCreateRequest{
		Name:        plan.Name.ValueString(),
		ChannelType: plan.ChannelType.ValueString(),
		Config:      apiConfig,
		IsActive:    plan.IsActive.ValueBool(),
	}

	if !plan.OrganizationID.IsNull() && !plan.OrganizationID.IsUnknown() {
		v := int(plan.OrganizationID.ValueInt64())
		createReq.OrganizationID = &v
	}

	channel, err := r.client.CreateNotificationChannel(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating notification channel", err.Error())
		return
	}

	resp.Diagnostics.Append(mapChannelToState(ctx, channel, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *notificationChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NotificationChannelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	channel, err := r.client.GetNotificationChannel(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading notification channel", err.Error())
		return
	}

	resp.Diagnostics.Append(mapChannelToState(ctx, channel, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *notificationChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NotificationChannelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configMap := make(map[string]string)
	resp.Diagnostics.Append(plan.Config.ElementsAs(ctx, &configMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiConfig := make(map[string]any, len(configMap))
	for k, v := range configMap {
		apiConfig[k] = v
	}

	name := plan.Name.ValueString()
	isActive := plan.IsActive.ValueBool()

	updateReq := client.NotificationChannelUpdateRequest{
		Name:     &name,
		Config:   &apiConfig,
		IsActive: &isActive,
	}

	channel, err := r.client.UpdateNotificationChannel(ctx, int(plan.ID.ValueInt64()), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating notification channel", err.Error())
		return
	}

	resp.Diagnostics.Append(mapChannelToState(ctx, channel, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *notificationChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NotificationChannelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotificationChannel(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Error deleting notification channel", err.Error())
	}
}

func (r *notificationChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected integer notification channel ID, got: %s", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func mapChannelToState(ctx context.Context, channel *client.NotificationChannel, state *NotificationChannelModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.Int64Value(int64(channel.ID))
	state.Name = types.StringValue(channel.Name)
	state.ChannelType = types.StringValue(channel.ChannelType)
	state.IsActive = types.BoolValue(channel.IsActive)

	if channel.OrganizationID != nil {
		state.OrganizationID = types.Int64Value(int64(*channel.OrganizationID))
	} else {
		state.OrganizationID = types.Int64Null()
	}

	configMap := make(map[string]string, len(channel.Config))
	for k, v := range channel.Config {
		configMap[k] = fmt.Sprintf("%v", v)
	}
	tfMap, d := types.MapValueFrom(ctx, types.StringType, configMap)
	diags.Append(d...)
	state.Config = tfMap

	return diags
}
