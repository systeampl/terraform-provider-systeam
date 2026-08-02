package service

import "github.com/hashicorp/terraform-plugin-framework/types"

type serviceDataSourceModel struct {
	OrganizationID types.Int64  `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	ID             types.Int64  `tfsdk:"id"`
	Slug           types.String `tfsdk:"slug"`
	Tier           types.String `tfsdk:"tier"`
	IsActive       types.Bool   `tfsdk:"is_active"`
}
