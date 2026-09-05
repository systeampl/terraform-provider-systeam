package check

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.ResourceWithValidateConfig = &checkResource{}

// selectorOnlyMatchTypes are the content_match_type values that only mean
// something against a selected element, so the API rejects them without a
// selector.
var selectorOnlyMatchTypes = map[string]bool{"equals": true, "exists": true, "not_exists": true}

// configIssue is one cross-field problem found in a check config, kept separate
// from diagnostics so the rules can be unit-tested without a tfsdk.Config.
type configIssue struct {
	Attribute string
	Summary   string
	Detail    string
}

// selectorContentMatchIssues applies the cross-field rules the API enforces for
// selector-scoped content matching. Value syntax (valid CSS, the 200-character
// selector limit) stays server-side so there is one source of truth for it.
func selectorContentMatchIssues(matchType, selector, extract, attribute string, selectorKnown bool) []configIssue {
	var issues []configIssue
	if extract == "attribute" && attribute == "" {
		issues = append(issues, configIssue{
			Attribute: "content_match_attribute",
			Summary:   "content_match_attribute is required",
			Detail:    "content_match_extract = \"attribute\" needs content_match_attribute to name the attribute to read.",
		})
	}

	// An unknown selector (an interpolated value) may still be set at apply
	// time, so only a known-empty one is a config error.
	if !selectorKnown || selector != "" {
		return issues
	}

	if selectorOnlyMatchTypes[matchType] {
		issues = append(issues, configIssue{
			Attribute: "content_match_type",
			Summary:   "content_match_type requires content_match_selector",
			Detail:    "content_match_type = \"" + matchType + "\" matches a selected element, so content_match_selector must be set.",
		})
	}
	for _, f := range []struct{ attr, value string }{
		{"content_match_extract", extract},
		{"content_match_attribute", attribute},
	} {
		if f.value != "" {
			issues = append(issues, configIssue{
				Attribute: f.attr,
				Summary:   f.attr + " requires content_match_selector",
				Detail:    f.attr + " describes what to read from the selected elements, so content_match_selector must be set.",
			})
		}
	}
	return issues
}

func (r *checkResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg CheckModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	issues := selectorContentMatchIssues(
		cfg.ContentMatchType.ValueString(),
		cfg.ContentMatchSelector.ValueString(),
		cfg.ContentMatchExtract.ValueString(),
		cfg.ContentMatchAttribute.ValueString(),
		!cfg.ContentMatchSelector.IsUnknown(),
	)
	for _, issue := range issues {
		resp.Diagnostics.AddAttributeError(path.Root(issue.Attribute), issue.Summary, issue.Detail)
	}
}
