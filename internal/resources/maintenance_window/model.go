package maintenance_window

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MaintenanceWindowModel struct {
	ID                types.Int64          `tfsdk:"id"`
	Name              types.String         `tfsdk:"name"`
	Description       types.String         `tfsdk:"description"`
	StartTime         types.String         `tfsdk:"start_time"`
	EndTime           types.String         `tfsdk:"end_time"`
	Timezone          types.String         `tfsdk:"timezone"`
	OrganizationID    types.Int64          `tfsdk:"organization_id"`
	CheckIDs          types.List           `tfsdk:"check_ids"`
	ProjectIDs        types.List           `tfsdk:"project_ids"`
	IsRecurring       types.Bool           `tfsdk:"is_recurring"`
	RecurrencePattern jsontypes.Normalized `tfsdk:"recurrence_pattern"`
	IsActive          types.Bool           `tfsdk:"is_active"`
}
