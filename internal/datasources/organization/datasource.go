package organization

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	syschecks "github.com/systeampl/syschecks-go"
	"github.com/systeampl/terraform-provider-systeam/internal/sdkutil"
)

var (
	_ datasource.DataSource              = &organizationDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationDataSource{}
)

type organizationDataSource struct {
	sdk *syschecks.Client
}

func NewDataSource() datasource.DataSource {
	return &organizationDataSource{}
}

func (d *organizationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *organizationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a SysTeam Healthchecks organization by slug.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier of the organization.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the organization.",
			},
			"slug": schema.StringAttribute{
				Required:    true,
				Description: "The slug of the organization.",
			},
		},
	}
}

func (d *organizationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	sdk, ok := req.ProviderData.(*syschecks.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *syschecks.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.sdk = sdk
}

func (d *organizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config OrganizationModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := d.sdk.Organizations.GetOrganizationBySlug(ctx, config.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading organization", err.Error())
		return
	}

	config.ID = types.Int64Value(int64(org.Id))
	config.Name = types.StringValue(org.Name)
	config.Slug = sdkutil.Str(org.Slug)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
