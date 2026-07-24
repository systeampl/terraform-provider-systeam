package lifecycle_watch

import "github.com/hashicorp/terraform-plugin-framework/types"

type LifecycleWatchModel struct {
	ID             types.Int64  `tfsdk:"id"`
	OrganizationID types.Int64  `tfsdk:"organization_id"`
	Vendor         types.String `tfsdk:"vendor"`
	ResourceType   types.String `tfsdk:"resource_type"`
	Platform       types.String `tfsdk:"platform"`
	ResourceID     types.String `tfsdk:"resource_id"`
	NotifyOnNew    types.Bool   `tfsdk:"notify_on_new"`
	Notify90d      types.Bool   `tfsdk:"notify_90d"`
	Notify30d      types.Bool   `tfsdk:"notify_30d"`
	Notify7d       types.Bool   `tfsdk:"notify_7d"`
	ChannelIDs     types.Set    `tfsdk:"channel_ids"`
}
