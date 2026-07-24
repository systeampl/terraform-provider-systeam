package service

import "github.com/hashicorp/terraform-plugin-framework/types"

type ServiceModel struct {
	ID                     types.Int64  `tfsdk:"id"`
	OrganizationID         types.Int64  `tfsdk:"organization_id"`
	Name                   types.String `tfsdk:"name"`
	Slug                   types.String `tfsdk:"slug"`
	Description            types.String `tfsdk:"description"`
	RepoURL                types.String `tfsdk:"repo_url"`
	DocsURL                types.String `tfsdk:"docs_url"`
	OwnerTeamID            types.Int64  `tfsdk:"owner_team_id"`
	EscalationPolicyID     types.Int64  `tfsdk:"escalation_policy_id"`
	Tier                   types.String `tfsdk:"tier"`
	NotificationChannelIDs types.Set    `tfsdk:"notification_channel_ids"`
	IsActive               types.Bool   `tfsdk:"is_active"`
}
