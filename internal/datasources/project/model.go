package project

import "github.com/hashicorp/terraform-plugin-framework/types"

type projectDataSourceModel struct {
	OrganizationID       types.Int64  `tfsdk:"organization_id"`
	Name                 types.String `tfsdk:"name"`
	ID                   types.Int64  `tfsdk:"id"`
	Description          types.String `tfsdk:"description"`
	AccessControlEnabled types.Bool   `tfsdk:"access_control_enabled"`
}
