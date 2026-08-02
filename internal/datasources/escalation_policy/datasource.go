package escalation_policy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	syschecks "github.com/systeampl/syschecks-go"
)

var (
	_ datasource.DataSource              = &escalationPolicyDataSource{}
	_ datasource.DataSourceWithConfigure = &escalationPolicyDataSource{}
)

type escalationPolicyDataSource struct {
	sdk *syschecks.Client
}

func NewDataSource() datasource.DataSource {
	return &escalationPolicyDataSource{}
}

func (d *escalationPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_escalation_policy"
}

func (d *escalationPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing escalation policy by name within an organization.",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.Int64Attribute{
				Required:    true,
				Description: "The organization the escalation policy belongs to.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the escalation policy to look up.",
			},
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier of the escalation policy.",
			},
			"is_active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the escalation policy is active.",
			},
			"team_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The team the escalation policy belongs to.",
			},
		},
	}
}

func (d *escalationPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	sdk, ok := req.ProviderData.(*syschecks.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *syschecks.Client, got: %T", req.ProviderData))
		return
	}
	d.sdk = sdk
}

func (d *escalationPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg escalationPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policies, err := d.sdk.Oncall.ListPolicies(ctx, int(cfg.OrganizationID.ValueInt64()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error listing escalation policies", err.Error())
		return
	}

	name := cfg.Name.ValueString()
	for i := range *policies {
		p := &(*policies)[i]
		if p.Name != name {
			continue
		}
		cfg.ID = types.Int64Value(int64(p.Id))
		cfg.IsActive = types.BoolValue(p.IsActive)
		if p.TeamId != nil {
			cfg.TeamID = types.Int64Value(int64(*p.TeamId))
		} else {
			cfg.TeamID = types.Int64Null()
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
		return
	}

	resp.Diagnostics.AddError("Escalation policy not found", fmt.Sprintf("No escalation policy named %q in organization %d", name, cfg.OrganizationID.ValueInt64()))
}
