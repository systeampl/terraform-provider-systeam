package oncall_schedule

import "github.com/hashicorp/terraform-plugin-framework/types"

type OnCallScheduleModel struct {
	ID             types.Int64              `tfsdk:"id"`
	Name           types.String             `tfsdk:"name"`
	OrganizationID types.Int64              `tfsdk:"organization_id"`
	Timezone       types.String             `tfsdk:"timezone"`
	RotationType   types.String             `tfsdk:"rotation_type"`
	IsActive       types.Bool               `tfsdk:"is_active"`
	Participants   []OnCallParticipantModel `tfsdk:"participant"`
}

type OnCallParticipantModel struct {
	UserID   types.Int64 `tfsdk:"user_id"`
	Position types.Int64 `tfsdk:"position"`
}
