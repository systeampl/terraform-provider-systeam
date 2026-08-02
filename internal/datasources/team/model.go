package team

import "github.com/hashicorp/terraform-plugin-framework/types"

type teamDataSourceModel struct {
	OrganizationID types.Int64  `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	ID             types.Int64  `tfsdk:"id"`
	Slug           types.String `tfsdk:"slug"`
	Description    types.String `tfsdk:"description"`
	IsActive       types.Bool   `tfsdk:"is_active"`
}
