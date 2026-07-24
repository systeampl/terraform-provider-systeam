package playbook

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PlaybookModel struct {
	ID                           types.Int64          `tfsdk:"id"`
	OrganizationID               types.Int64          `tfsdk:"organization_id"`
	Name                         types.String         `tfsdk:"name"`
	Description                  types.String         `tfsdk:"description"`
	TriggerType                  types.String         `tfsdk:"trigger_type"`
	TriggerConditions            jsontypes.Normalized `tfsdk:"trigger_conditions"`
	ServiceID                    types.Int64          `tfsdk:"service_id"`
	SuppressDefaultNotifications types.Bool           `tfsdk:"suppress_default_notifications"`
	Steps                        []PlaybookStepModel  `tfsdk:"step"`
	IsActive                     types.Bool           `tfsdk:"is_active"`
}

type PlaybookStepModel struct {
	StepOrder      types.Int64          `tfsdk:"step_order"`
	Name           types.String         `tfsdk:"name"`
	ActionType     types.String         `tfsdk:"action_type"`
	Config         jsontypes.Normalized `tfsdk:"config"`
	Conditions     jsontypes.Normalized `tfsdk:"conditions"`
	ParallelGroup  types.String         `tfsdk:"parallel_group"`
	IsManual       types.Bool           `tfsdk:"is_manual"`
	TimeoutSeconds types.Int64          `tfsdk:"timeout_seconds"`
}
