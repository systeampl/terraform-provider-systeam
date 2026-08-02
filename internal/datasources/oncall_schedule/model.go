package oncall_schedule

import "github.com/hashicorp/terraform-plugin-framework/types"

type oncallScheduleDataSourceModel struct {
	OrganizationID types.Int64  `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	ID             types.Int64  `tfsdk:"id"`
	Timezone       types.String `tfsdk:"timezone"`
	IsActive       types.Bool   `tfsdk:"is_active"`
	TeamID         types.Int64  `tfsdk:"team_id"`
}
