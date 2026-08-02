package service

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	syschecks "github.com/systeampl/syschecks-go"
)

var (
	_ datasource.DataSource              = &serviceDataSource{}
	_ datasource.DataSourceWithConfigure = &serviceDataSource{}
)

type serviceDataSource struct {
	sdk *syschecks.Client
}

func NewDataSource() datasource.DataSource {
	return &serviceDataSource{}
}

func (d *serviceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (d *serviceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing service by name within an organization.",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.Int64Attribute{
				Required:    true,
				Description: "The organization the service belongs to.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the service to look up.",
			},
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier of the service.",
			},
			"slug": schema.StringAttribute{
				Computed:    true,
				Description: "The service slug.",
			},
			"tier": schema.StringAttribute{
				Computed:    true,
				Description: "The service tier.",
			},
			"is_active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the service is active.",
			},
		},
	}
}

func (d *serviceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg serviceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	services, err := d.sdk.Services.ListServices(ctx, int(cfg.OrganizationID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error listing services", err.Error())
		return
	}

	name := cfg.Name.ValueString()
	for i := range *services {
		s := &(*services)[i]
		if s.Name != name {
			continue
		}
		cfg.ID = types.Int64Value(int64(s.Id))
		cfg.Slug = types.StringValue(s.Slug)
		cfg.Tier = types.StringValue(s.Tier)
		cfg.IsActive = types.BoolValue(s.IsActive)
		resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
		return
	}

	resp.Diagnostics.AddError("Service not found", fmt.Sprintf("No service named %q in organization %d", name, cfg.OrganizationID.ValueInt64()))
}
