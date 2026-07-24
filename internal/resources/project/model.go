package project

import "github.com/hashicorp/terraform-plugin-framework/types"

type ProjectModel struct {
	ID                   types.Int64  `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	OrganizationID       types.Int64  `tfsdk:"organization_id"`
	AccessControlEnabled types.Bool   `tfsdk:"access_control_enabled"`
}
