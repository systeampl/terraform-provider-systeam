package agent_registration_token

import "github.com/hashicorp/terraform-plugin-framework/types"

type AgentRegistrationTokenModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.Int64  `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Mode           types.String `tfsdk:"mode"`
	Token          types.String `tfsdk:"token"`
	ExpiresAt      types.String `tfsdk:"expires_at"`
}
