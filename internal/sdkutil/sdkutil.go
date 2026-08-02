// Package sdkutil holds the small shared helpers resources use while they run on
// the official Go SDK (github.com/systeampl/syschecks-go).
package sdkutil

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/types"
	syschecks "github.com/systeampl/syschecks-go"
)

// IsNotFound reports whether err is an SDK error carrying a 404 status.
func IsNotFound(err error) bool {
	var e *syschecks.Error
	return errors.As(err, &e) && e.StatusCode == 404
}

// StrPtr converts a framework String to *string, nil when null or unknown.
func StrPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

// Str converts a *string to a framework String, null when nil.
func Str(p *string) types.String {
	if p == nil {
		return types.StringNull()
	}
	return types.StringValue(*p)
}

// IntPtr converts a framework Int64 to *int, nil when null or unknown.
func IntPtr(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := int(v.ValueInt64())
	return &i
}

// Int converts a *int to a framework Int64, null when nil.
func Int(p *int) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*p))
}
