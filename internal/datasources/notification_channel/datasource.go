package notification_channel

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	syschecks "github.com/systeampl/syschecks-go"
	"github.com/systeampl/syschecks-go/models"
)

var (
	_ datasource.DataSource              = &notificationChannelDataSource{}
	_ datasource.DataSourceWithConfigure = &notificationChannelDataSource{}
)

type notificationChannelDataSource struct {
	sdk *syschecks.Client
}

func NewDataSource() datasource.DataSource {
	return &notificationChannelDataSource{}
}

func (d *notificationChannelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_channel"
}

func (d *notificationChannelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing notification channel by name within an organization.",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.Int64Attribute{
				Required:    true,
				Description: "The organization the notification channel belongs to.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the notification channel to look up.",
			},
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier of the notification channel.",
			},
			"channel_type": schema.StringAttribute{
				Computed:    true,
				Description: "The notification channel type.",
			},
			"is_active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the notification channel is active.",
			},
		},
	}
}

func (d *notificationChannelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *notificationChannelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg notificationChannelDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgID := int(cfg.OrganizationID.ValueInt64())
	params := &models.ListChannelsParams{
		OrgId: &orgID,
	}

	channels, err := d.sdk.NotificationChannels.ListChannels(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Error listing notification channels", err.Error())
		return
	}

	name := cfg.Name.ValueString()
	for i := range *channels {
		ch := &(*channels)[i]
		if ch.Name != name {
			continue
		}
		cfg.ID = types.Int64Value(int64(ch.Id))
		cfg.ChannelType = types.StringValue(string(ch.ChannelType))
		cfg.IsActive = types.BoolValue(ch.IsActive)
		resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
		return
	}

	resp.Diagnostics.AddError("Notification channel not found", fmt.Sprintf("No notification channel named %q in organization %d", name, cfg.OrganizationID.ValueInt64()))
}
