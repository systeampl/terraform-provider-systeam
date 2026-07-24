package shim

import (
	"github.com/hashicorp/terraform-plugin-framework/provider"
	internal "github.com/pawel-cygal/terraform-provider-systeam/internal/provider"
)

// New returns a new instance of the SysTeam provider.
// This package exists to re-export the provider constructor from internal/provider
// so that external modules (like the Pulumi bridge) can access it.
func New() provider.Provider {
	return internal.New()
}
