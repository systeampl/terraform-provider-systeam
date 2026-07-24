package escalation_policy

import "github.com/hashicorp/terraform-plugin-framework/types"

type EscalationPolicyModel struct {
	ID             types.Int64           `tfsdk:"id"`
	Name           types.String          `tfsdk:"name"`
	OrganizationID types.Int64           `tfsdk:"organization_id"`
	IsActive       types.Bool            `tfsdk:"is_active"`
	Steps          []EscalationStepModel `tfsdk:"step"`
}

type EscalationStepModel struct {
	StepOrder        types.Int64  `tfsdk:"step_order"`
	DelayMinutes     types.Int64  `tfsdk:"delay_minutes"`
	TargetType       types.String `tfsdk:"target_type"`
	TargetUserID     types.Int64  `tfsdk:"target_user_id"`
	TargetScheduleID types.Int64  `tfsdk:"target_schedule_id"`
	TargetChannelID  types.Int64  `tfsdk:"target_channel_id"`
}
