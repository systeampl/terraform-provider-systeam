package contact_method

import "github.com/hashicorp/terraform-plugin-framework/types"

type ContactMethodModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Kind     types.String `tfsdk:"kind"`
	Value    types.String `tfsdk:"value"`
	Label    types.String `tfsdk:"label"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Verified types.Bool   `tfsdk:"verified"`
}
