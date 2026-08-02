package project

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	syschecks "github.com/systeampl/syschecks-go"
)

var (
	_ datasource.DataSource              = &projectDataSource{}
	_ datasource.DataSourceWithConfigure = &projectDataSource{}
)

type projectDataSource struct {
	sdk *syschecks.Client
}

func NewDataSource() datasource.DataSource {
	return &projectDataSource{}
}

func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing project by name within an organization.",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.Int64Attribute{
				Required:    true,
				Description: "The organization the project belongs to.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the project to look up.",
			},
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier of the project.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The project description.",
			},
			"access_control_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether access control is enabled for the project.",
			},
		},
	}
}

func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg projectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projects, err := d.sdk.Organizations.ListOrganizationProjects(ctx, int(cfg.OrganizationID.ValueInt64()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error listing projects", err.Error())
		return
	}

	name := cfg.Name.ValueString()
	for i := range *projects {
		p := &(*projects)[i]
		if p.Name != name {
			continue
		}
		cfg.ID = types.Int64Value(int64(p.Id))
		if p.Description != nil {
			cfg.Description = types.StringValue(*p.Description)
		} else {
			cfg.Description = types.StringNull()
		}
		if p.AccessControlEnabled != nil {
			cfg.AccessControlEnabled = types.BoolValue(*p.AccessControlEnabled)
		} else {
			cfg.AccessControlEnabled = types.BoolNull()
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
		return
	}

	resp.Diagnostics.AddError("Project not found", fmt.Sprintf("No project named %q in organization %d", name, cfg.OrganizationID.ValueInt64()))
}
