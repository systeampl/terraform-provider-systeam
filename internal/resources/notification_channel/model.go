package notification_channel

import "github.com/hashicorp/terraform-plugin-framework/types"

type NotificationChannelModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ChannelType    types.String `tfsdk:"channel_type"`
	Config         types.Map    `tfsdk:"config"`
	IsActive       types.Bool   `tfsdk:"is_active"`
	OrganizationID types.Int64  `tfsdk:"organization_id"`
}
