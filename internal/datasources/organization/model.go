package organization

import "github.com/hashicorp/terraform-plugin-framework/types"

type OrganizationModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Slug types.String `tfsdk:"slug"`
}
