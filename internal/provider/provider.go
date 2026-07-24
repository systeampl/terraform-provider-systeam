package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/systeampl/terraform-provider-systeam/internal/client"
	orgds "github.com/systeampl/terraform-provider-systeam/internal/datasources/organization"
	"github.com/systeampl/terraform-provider-systeam/internal/resources/agent_registration_token"
	checkresource "github.com/systeampl/terraform-provider-systeam/internal/resources/check"
	"github.com/systeampl/terraform-provider-systeam/internal/resources/check_slo"
	"github.com/systeampl/terraform-provider-systeam/internal/resources/contact_method"
	"github.com/systeampl/terraform-provider-systeam/internal/resources/escalation_policy"
	"github.com/systeampl/terraform-provider-systeam/internal/resources/integration_key"
	"github.com/systeampl/terraform-provider-systeam/internal/resources/lifecycle_watch"
	"github.com/systeampl/terraform-provider-systeam/internal/resources/maintenance_window"
	"github.com/systeampl/terraform-provider-systeam/internal/resources/notification_channel"
	"github.com/systeampl/terraform-provider-systeam/internal/resources/oncall_schedule"
	"github.com/systeampl/terraform-provider-systeam/internal/resources/playbook"
	"github.com/systeampl/terraform-provider-systeam/internal/resources/project"
	"github.com/systeampl/terraform-provider-systeam/internal/resources/service"
	"github.com/systeampl/terraform-provider-systeam/internal/resources/status_page"
	"github.com/systeampl/terraform-provider-systeam/internal/resources/team"
)

var _ provider.Provider = &systeamProvider{}

type systeamProvider struct{}

type systeamProviderModel struct {
	APIURL   types.String `tfsdk:"api_url"`
	APIToken types.String `tfsdk:"api_token"`
}

func New() provider.Provider {
	return &systeamProvider{}
}

func (p *systeamProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "systeam"
}

func (p *systeamProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Optional:    true,
				Description: "Base URL of the SysTeam Healthchecks API. Can also be set via SYSTEAM_API_URL environment variable.",
			},
			"api_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "PAT token for API authentication. Can also be set via SYSTEAM_API_TOKEN environment variable.",
			},
		},
	}
}

func (p *systeamProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config systeamProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiURL := os.Getenv("SYSTEAM_API_URL")
	if !config.APIURL.IsNull() && !config.APIURL.IsUnknown() {
		apiURL = config.APIURL.ValueString()
	}

	apiToken := os.Getenv("SYSTEAM_API_TOKEN")
	if !config.APIToken.IsNull() && !config.APIToken.IsUnknown() {
		apiToken = config.APIToken.ValueString()
	}

	if apiURL == "" {
		resp.Diagnostics.AddError(
			"Missing API URL",
			"The provider requires an API URL. Set it in the provider configuration block or via the SYSTEAM_API_URL environment variable.",
		)
	}

	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing API Token",
			"The provider requires an API token. Set it in the provider configuration block or via the SYSTEAM_API_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	c := client.NewClient(apiURL, apiToken)
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *systeamProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		project.NewResource,
		checkresource.NewResource,
		check_slo.NewResource,
		notification_channel.NewResource,
		maintenance_window.NewResource,
		escalation_policy.NewResource,
		oncall_schedule.NewResource,
		status_page.NewResource,
		integration_key.NewResource,
		agent_registration_token.NewResource,
		team.NewResource,
		service.NewResource,
		lifecycle_watch.NewResource,
		contact_method.NewResource,
		playbook.NewResource,
	}
}

func (p *systeamProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		orgds.NewDataSource,
	}
}
