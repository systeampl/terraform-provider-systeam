package oncall_schedule

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	syschecks "github.com/systeampl/syschecks-go"
)

var (
	_ datasource.DataSource              = &oncallScheduleDataSource{}
	_ datasource.DataSourceWithConfigure = &oncallScheduleDataSource{}
)

type oncallScheduleDataSource struct {
	sdk *syschecks.Client
}

func NewDataSource() datasource.DataSource {
	return &oncallScheduleDataSource{}
}

func (d *oncallScheduleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oncall_schedule"
}

func (d *oncallScheduleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing on-call schedule by name within an organization.",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.Int64Attribute{
				Required:    true,
				Description: "The organization the on-call schedule belongs to.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the on-call schedule to look up.",
			},
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier of the on-call schedule.",
			},
			"timezone": schema.StringAttribute{
				Computed:    true,
				Description: "The timezone of the on-call schedule.",
			},
			"is_active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the on-call schedule is active.",
			},
			"team_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The identifier of the team the on-call schedule belongs to.",
			},
		},
	}
}

func (d *oncallScheduleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *oncallScheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg oncallScheduleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schedules, err := d.sdk.Oncall.ListSchedules(ctx, int(cfg.OrganizationID.ValueInt64()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error listing on-call schedules", err.Error())
		return
	}

	name := cfg.Name.ValueString()
	for i := range *schedules {
		s := &(*schedules)[i]
		if s.Name != name {
			continue
		}
		cfg.ID = types.Int64Value(int64(s.Id))
		cfg.Timezone = types.StringValue(s.Timezone)
		cfg.IsActive = types.BoolValue(s.IsActive)
		if s.TeamId != nil {
			cfg.TeamID = types.Int64Value(int64(*s.TeamId))
		} else {
			cfg.TeamID = types.Int64Null()
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
		return
	}

	resp.Diagnostics.AddError("On-call schedule not found", fmt.Sprintf("No on-call schedule named %q in organization %d", name, cfg.OrganizationID.ValueInt64()))
}
