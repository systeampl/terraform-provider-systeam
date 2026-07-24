package integration_key

import "github.com/hashicorp/terraform-plugin-framework/types"

type IntegrationKeyModel struct {
	ID                    types.Int64  `tfsdk:"id"`
	OrganizationID        types.Int64  `tfsdk:"organization_id"`
	Name                  types.String `tfsdk:"name"`
	EscalationPolicyID    types.Int64  `tfsdk:"escalation_policy_id"`
	GroupingType          types.String `tfsdk:"grouping_type"`
	GroupingWindowSeconds types.Int64  `tfsdk:"grouping_window_seconds"`
	TokenPrefix           types.String `tfsdk:"token_prefix"`
	Token                 types.String `tfsdk:"token"`
	IsActive              types.Bool   `tfsdk:"is_active"`
}
