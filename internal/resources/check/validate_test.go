package check

import "testing"

// The plan-time rules must agree with what the API enforces in
// validate_content_match_config — a config the provider accepts and the API
// then rejects turns a plan error into a failed apply.
func TestSelectorContentMatchIssues(t *testing.T) {
	cases := []struct {
		name                                 string
		matchType, selector, extract, attrib string
		selectorKnown                        bool
		wantAttrs                            []string
	}{
		{name: "plain contains needs nothing", matchType: "contains", selectorKnown: true},
		{name: "equals without selector", matchType: "equals", selectorKnown: true,
			wantAttrs: []string{"content_match_type"}},
		{name: "exists without selector", matchType: "exists", selectorKnown: true,
			wantAttrs: []string{"content_match_type"}},
		{name: "not_exists without selector", matchType: "not_exists", selectorKnown: true,
			wantAttrs: []string{"content_match_type"}},
		{name: "equals with selector is fine", matchType: "equals", selector: "h1", selectorKnown: true},
		{name: "extract without selector", matchType: "contains", extract: "text", selectorKnown: true,
			wantAttrs: []string{"content_match_extract"}},
		{name: "attribute without selector", matchType: "contains", attrib: "href", selectorKnown: true,
			wantAttrs: []string{"content_match_attribute"}},
		{name: "everything missing a selector at once", matchType: "equals", extract: "attribute", attrib: "href", selectorKnown: true,
			wantAttrs: []string{"content_match_type", "content_match_extract", "content_match_attribute"}},
		{name: "unknown selector defers to apply", matchType: "equals", extract: "text", selectorKnown: false},
		{name: "extract=attribute without an attribute name", matchType: "contains", selector: "h1", extract: "attribute", selectorKnown: true,
			wantAttrs: []string{"content_match_attribute"}},
		{name: "extract=attribute with a name is fine", matchType: "contains", selector: "h1", extract: "attribute", attrib: "href", selectorKnown: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := selectorContentMatchIssues(tc.matchType, tc.selector, tc.extract, tc.attrib, tc.selectorKnown)
			if len(got) != len(tc.wantAttrs) {
				t.Fatalf("got %d issues %v, want %d for %v", len(got), got, len(tc.wantAttrs), tc.wantAttrs)
			}
			for i, want := range tc.wantAttrs {
				if got[i].Attribute != want {
					t.Errorf("issue %d on %q, want %q", i, got[i].Attribute, want)
				}
			}
		})
	}
}
