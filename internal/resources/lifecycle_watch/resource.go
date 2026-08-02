package lifecycle_watch

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	syschecks "github.com/systeampl/syschecks-go"
	"github.com/systeampl/syschecks-go/models"
	"github.com/systeampl/terraform-provider-systeam/internal/sdkutil"
)

var (
	_ resource.Resource                = &lifecycleWatchResource{}
	_ resource.ResourceWithConfigure   = &lifecycleWatchResource{}
	_ resource.ResourceWithImportState = &lifecycleWatchResource{}
)

type lifecycleWatchResource struct {
	sdk *syschecks.Client
}

func NewResource() resource.Resource {
	return &lifecycleWatchResource{}
}

func (r *lifecycleWatchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected import ID format: org_id:watch_id")
		return
	}
	orgID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid org_id in import ID", err.Error())
		return
	}
	watchID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid watch_id in import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), watchID)...)
}

func (r *lifecycleWatchResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lifecycle_watch"
}

func (r *lifecycleWatchResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// The API upserts by the natural key (vendor + resource_type + platform +
	// resource_id), so those are RequiresReplace: changing one addresses a
	// different watch. notify_* toggles are updated in place via the same upsert.
	resp.Schema = schema.Schema{
		Description: "Tracks the lifecycle / end-of-life of an external resource (e.g. an AI model, the " +
			"'Technology Watch' feature) and notifies at 90/30/7-day thresholds and on new releases.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:      true,
				Description:   "The unique identifier of the watch.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.Int64Attribute{
				Required:      true,
				Description:   "The organization this watch belongs to.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"vendor": schema.StringAttribute{
				Required:      true,
				Description:   "The vendor being watched (e.g. 'openai', 'anthropic').",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"resource_type": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Default:       stringdefault.StaticString("ai_model"),
				Description:   "The kind of resource being watched (default 'ai_model').",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"platform": schema.StringAttribute{
				Optional:      true,
				Description:   "Optional platform qualifier (part of the watch's identity).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"resource_id": schema.StringAttribute{
				Optional:      true,
				Description:   "Optional specific resource id to watch (part of the watch's identity).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"notify_on_new": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				Description: "Notify when a new release/version appears.",
			},
			"notify_90d": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				Description: "Notify 90 days before end-of-life.",
			},
			"notify_30d": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				Description: "Notify 30 days before end-of-life.",
			},
			"notify_7d": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				Description: "Notify 7 days before end-of-life.",
			},
			"channel_ids": schema.SetAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
				Description: "Notification channels to alert. Write-only: the API does not return these on read.",
			},
		},
	}
}

func (r *lifecycleWatchResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *lifecycleWatchResource) upsert(ctx context.Context, plan *LifecycleWatchModel) error {
	var channelIDs []int
	if !plan.ChannelIDs.IsNull() && !plan.ChannelIDs.IsUnknown() {
		var ids []int64
		plan.ChannelIDs.ElementsAs(ctx, &ids, false)
		for _, v := range ids {
			channelIDs = append(channelIDs, int(v))
		}
	}

	notifyOnNew := plan.NotifyOnNew.ValueBool()
	notify90d := plan.Notify90d.ValueBool()
	notify30d := plan.Notify30d.ValueBool()
	notify7d := plan.Notify7d.ValueBool()

	body := models.UpsertLifecycleWatchJSONRequestBody{
		Vendor:       plan.Vendor.ValueString(),
		Platform:     sdkutil.StrPtr(plan.Platform),
		ResourceId:   sdkutil.StrPtr(plan.ResourceID),
		ResourceType: sdkutil.StrPtr(plan.ResourceType),
		NotifyOnNew:  &notifyOnNew,
		Notify90d:    &notify90d,
		Notify30d:    &notify30d,
		Notify7d:     &notify7d,
	}
	if channelIDs != nil {
		body.ChannelIds = &channelIDs
	}

	raw, err := r.sdk.Lifecycle.UpsertLifecycleWatch(ctx, int(plan.OrganizationID.ValueInt64()), body)
	if err != nil {
		return err
	}
	var res struct {
		Id int `json:"id"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return err
	}
	plan.ID = types.Int64Value(int64(res.Id))
	return nil
}

func (r *lifecycleWatchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LifecycleWatchModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.upsert(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Error creating lifecycle watch", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *lifecycleWatchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LifecycleWatchModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	watch, err := r.sdk.Lifecycle.GetLifecycleWatch(ctx, int(state.OrganizationID.ValueInt64()), int(state.ID.ValueInt64()))
	if err != nil {
		if sdkutil.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading lifecycle watch", err.Error())
		return
	}

	// channel_ids is not returned by the API — leave state's value untouched.
	state.Vendor = sdkutil.Str(watch.Vendor)
	state.ResourceType = types.StringValue(watch.ResourceType)
	state.Platform = sdkutil.Str(watch.Platform)
	state.ResourceID = sdkutil.Str(watch.ResourceId)
	state.NotifyOnNew = types.BoolValue(watch.NotifyOnNew)
	state.Notify90d = types.BoolValue(watch.Notify90d)
	state.Notify30d = types.BoolValue(watch.Notify30d)
	state.Notify7d = types.BoolValue(watch.Notify7d)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *lifecycleWatchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LifecycleWatchModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.upsert(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Error updating lifecycle watch", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *lifecycleWatchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LifecycleWatchModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, err := r.sdk.Lifecycle.DeleteLifecycleWatch(ctx, int(state.OrganizationID.ValueInt64()), int(state.ID.ValueInt64())); err != nil {
		if sdkutil.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting lifecycle watch", err.Error())
	}
}
