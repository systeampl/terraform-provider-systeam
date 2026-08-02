package escalation_policy

import "github.com/hashicorp/terraform-plugin-framework/types"

type escalationPolicyDataSourceModel struct {
	OrganizationID types.Int64  `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	ID             types.Int64  `tfsdk:"id"`
	IsActive       types.Bool   `tfsdk:"is_active"`
	TeamID         types.Int64  `tfsdk:"team_id"`
}
