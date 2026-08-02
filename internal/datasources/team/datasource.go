package team

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	syschecks "github.com/systeampl/syschecks-go"
)

var (
	_ datasource.DataSource              = &teamDataSource{}
	_ datasource.DataSourceWithConfigure = &teamDataSource{}
)

type teamDataSource struct {
	sdk *syschecks.Client
}

func NewDataSource() datasource.DataSource {
	return &teamDataSource{}
}

func (d *teamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (d *teamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing team by name within an organization.",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.Int64Attribute{
				Required:    true,
				Description: "The organization the team belongs to.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the team to look up.",
			},
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier of the team.",
			},
			"slug": schema.StringAttribute{
				Computed:    true,
				Description: "The team slug.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The team description.",
			},
			"is_active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the team is active.",
			},
		},
	}
}

func (d *teamDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *teamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg teamDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teams, err := d.sdk.Teams.ListTeams(ctx, int(cfg.OrganizationID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error listing teams", err.Error())
		return
	}

	name := cfg.Name.ValueString()
	for i := range *teams {
		t := &(*teams)[i]
		if t.Name != name {
			continue
		}
		cfg.ID = types.Int64Value(int64(t.Id))
		cfg.Slug = types.StringValue(t.Slug)
		cfg.IsActive = types.BoolValue(t.IsActive)
		if t.Description != nil {
			cfg.Description = types.StringValue(*t.Description)
		} else {
			cfg.Description = types.StringNull()
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
		return
	}

	resp.Diagnostics.AddError("Team not found", fmt.Sprintf("No team named %q in organization %d", name, cfg.OrganizationID.ValueInt64()))
}
