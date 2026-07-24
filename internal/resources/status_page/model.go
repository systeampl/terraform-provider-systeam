package status_page

import "github.com/hashicorp/terraform-plugin-framework/types"

type StatusPageModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Slug           types.String `tfsdk:"slug"`
	Description    types.String `tfsdk:"description"`
	IsPublic       types.Bool   `tfsdk:"is_public"`
	CustomDomain   types.String `tfsdk:"custom_domain"`
	LogoURL        types.String `tfsdk:"logo_url"`
	CheckIDs       types.List   `tfsdk:"check_ids"`
	OrganizationID types.Int64  `tfsdk:"organization_id"`
	IsActive       types.Bool   `tfsdk:"is_active"`
}
