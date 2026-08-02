package notification_channel

import "github.com/hashicorp/terraform-plugin-framework/types"

type notificationChannelDataSourceModel struct {
	OrganizationID types.Int64  `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	ID             types.Int64  `tfsdk:"id"`
	ChannelType    types.String `tfsdk:"channel_type"`
	IsActive       types.Bool   `tfsdk:"is_active"`
}
