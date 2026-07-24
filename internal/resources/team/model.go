package team

import "github.com/hashicorp/terraform-plugin-framework/types"

type TeamModel struct {
	ID             types.Int64  `tfsdk:"id"`
	OrganizationID types.Int64  `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Slug           types.String `tfsdk:"slug"`
	Description    types.String `tfsdk:"description"`
	IsActive       types.Bool   `tfsdk:"is_active"`
	MemberCount    types.Int64  `tfsdk:"member_count"`
}
