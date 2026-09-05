package check

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/systeampl/syschecks-go/models"
)

// JSON-object fields use jsontypes.Normalized so formatting and key order don't
// produce spurious diffs — this is the contract Pulumi/Crossplane will inherit.
func TestJSONObjectSemanticEquality(t *testing.T) {
	a := jsontypes.NewNormalizedValue(`{"X-Api-Key":"abc","Accept":"application/json"}`)
	// Same data, different whitespace AND key order.
	b := jsontypes.NewNormalizedValue("{\n  \"Accept\": \"application/json\",\n  \"X-Api-Key\": \"abc\"\n}")

	eq, diags := a.StringSemanticEquals(context.Background(), b)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !eq {
		t.Error("JSON differing only in whitespace + key order must be semantically equal (no diff)")
	}

	// Genuinely different JSON must NOT be equal.
	c := jsontypes.NewNormalizedValue(`{"X-Api-Key":"different"}`)
	if eq2, _ := a.StringSemanticEquals(context.Background(), c); eq2 {
		t.Error("different JSON must be unequal")
	}
}

func TestMapToNormalized(t *testing.T) {
	if !mapToNormalized(nil).IsNull() {
		t.Error("nil map → null")
	}
	if m := (map[string]interface{}{"a": float64(1)}); mapToNormalized(&m).IsNull() {
		t.Error("non-empty map must not be null")
	}
}

func TestSliceToNormalized(t *testing.T) {
	if !sliceToNormalized(nil).IsNull() {
		t.Error("nil slice → null")
	}
	if s := ([]interface{}{"a"}); sliceToNormalized(&s).IsNull() {
		t.Error("non-empty slice must not be null")
	}
}

// TestContentChangeFieldsRoundTrip guards against the class of bug fixed in
// fix/create-check-missing-fields: a field accepted by the schema but dropped
// on read-back makes Terraform report "Provider produced inconsistent result
// after apply" because state no longer matches the applied config. Every
// field the API echoes back must be re-mapped in mapCheckToState.
func TestContentChangeFieldsRoundTrip(t *testing.T) {
	ctx := context.Background()

	enabled := true
	severity := "degraded"
	patterns := []string{`\d{4}-\d{2}-\d{2}`, `session_id=\w+`}
	geo := true
	check := &models.CheckResponse{
		ContentChangeEnabled:         &enabled,
		ContentChangeSeverity:        &severity,
		ContentIgnorePatterns:        &patterns,
		GeoContentConsistencyEnabled: &geo,
	}

	state := &CheckModel{}
	diags := mapCheckToState(ctx, check, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !state.ContentChangeEnabled.ValueBool() {
		t.Error("content_change_enabled: want true after read-back")
	}
	if got := state.ContentChangeSeverity.ValueString(); got != "degraded" {
		t.Errorf("content_change_severity: want %q, got %q", "degraded", got)
	}
	if !state.GeoContentConsistencyEnabled.ValueBool() {
		t.Error("geo_content_consistency_enabled: want true after read-back")
	}

	var gotPatterns []string
	if d := state.ContentIgnorePatterns.ElementsAs(ctx, &gotPatterns, false); d.HasError() {
		t.Fatalf("unexpected diagnostics reading back content_ignore_patterns: %v", d)
	}
	if len(gotPatterns) != len(patterns) {
		t.Fatalf("content_ignore_patterns: want %v, got %v", patterns, gotPatterns)
	}
	for i, p := range patterns {
		if gotPatterns[i] != p {
			t.Errorf("content_ignore_patterns[%d]: want %q, got %q", i, p, gotPatterns[i])
		}
	}

	// A plan that set content_ignore_patterns to the same list must round-trip
	// to an identical types.List value, or Terraform sees a post-apply diff.
	planList, d := types.ListValueFrom(ctx, types.StringType, patterns)
	if d.HasError() {
		t.Fatalf("unexpected diagnostics building plan list: %v", d)
	}
	if !planList.Equal(state.ContentIgnorePatterns) {
		t.Error("content_ignore_patterns: state after read-back does not equal the planned list")
	}
}

// TestContentIgnorePatternsNullRoundTrip mirrors dns_expected_ips: when the API
// returns no patterns (nil, the server-side default), state must come back
// null — not an empty non-null list, which would also produce a diff against a
// config that never set the attribute.
func TestContentIgnorePatternsNullRoundTrip(t *testing.T) {
	ctx := context.Background()

	check := &models.CheckResponse{ContentIgnorePatterns: nil}
	state := &CheckModel{}
	diags := mapCheckToState(ctx, check, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !state.ContentIgnorePatterns.IsNull() {
		t.Errorf("content_ignore_patterns: want null when API returns no patterns, got %v", state.ContentIgnorePatterns)
	}
}

// TestContentIgnorePatternsClearOnUpdate pins the provider's half of the
// contract the backend guarantees via
// test_api_clears_patterns_when_sent_an_empty_list: a config that explicitly
// sets content_ignore_patterns = [] must produce a request body containing
// "content_ignore_patterns":[] on Update, not an omitted field — otherwise a
// previously-set pattern can never be removed (the backend's `if x is not
// None` guard skips the clear).
func TestContentIgnorePatternsClearOnUpdate(t *testing.T) {
	ctx := context.Background()

	emptyList, d := types.ListValueFrom(ctx, types.StringType, []string{})
	if d.HasError() {
		t.Fatalf("unexpected diagnostics building empty list: %v", d)
	}
	cleared, d := listToStrs(ctx, emptyList)
	if d.HasError() {
		t.Fatalf("unexpected diagnostics from listToStrs: %v", d)
	}
	if cleared == nil {
		t.Fatal("listToStrs([]) must return a non-nil empty slice — a nil slice here " +
			"omits the field from the update payload and the backend never clears the patterns")
	}

	var updateReq models.UpdateCheckJSONRequestBody
	updateReq.ContentIgnorePatterns = &cleared
	body, err := json.Marshal(updateReq)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(body), `"content_ignore_patterns":[]`) {
		t.Errorf("update body must contain an explicit empty list, got: %s", body)
	}

	// And a null (never-configured) list must stay omitted.
	nullCleared, d := listToStrs(ctx, types.ListNull(types.StringType))
	if d.HasError() {
		t.Fatalf("unexpected diagnostics from listToStrs(null): %v", d)
	}
	if nullCleared != nil {
		t.Error("listToStrs(null) must return nil so the field is omitted from the payload")
	}
}

// TestSelectorContentMatchFieldsRoundTrip extends the read-back contract to the
// DOM-assertion fields: anything the API echoes must be re-mapped, or Terraform
// reports "inconsistent result after apply".
func TestSelectorContentMatchFieldsRoundTrip(t *testing.T) {
	ctx := context.Background()

	selector, extract, attribute, matchType := "link[rel=canonical]", "attribute", "href", "equals"
	state := &CheckModel{}
	diags := mapCheckToState(ctx, &models.CheckResponse{
		ContentMatchSelector:  &selector,
		ContentMatchExtract:   &extract,
		ContentMatchAttribute: &attribute,
		ContentMatchType:      &matchType,
	}, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	for name, got := range map[string]struct{ have, want string }{
		"content_match_selector":  {state.ContentMatchSelector.ValueString(), selector},
		"content_match_extract":   {state.ContentMatchExtract.ValueString(), extract},
		"content_match_attribute": {state.ContentMatchAttribute.ValueString(), attribute},
		"content_match_type":      {state.ContentMatchType.ValueString(), matchType},
	} {
		if got.have != got.want {
			t.Errorf("%s: want %q, got %q", name, got.want, got.have)
		}
	}
}

// TestContentMatchExtractOmittedWhenEmpty pins an asymmetry the other string
// fields do not have. The API rejects content_match_extract="" outright on
// create (CheckCreate's model validator only accepts text/attribute), while it
// treats "" on update as "clear this back to NULL". So the create payload must
// omit an unset extract, and the update payload must still carry the empty
// string. Sending "" on create would make every content_match_enabled check
// fail with a 422.
func TestContentMatchExtractOmittedWhenEmpty(t *testing.T) {
	var createReq models.CreateNewCheckJSONRequestBody
	setContentMatchExtract(&createReq.ContentMatchExtract, types.StringValue(""), false)
	body, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(body), "content_match_extract") {
		t.Errorf("create body must omit an empty content_match_extract, got: %s", body)
	}

	setContentMatchExtract(&createReq.ContentMatchExtract, types.StringValue("attribute"), false)
	body, _ = json.Marshal(createReq)
	if !strings.Contains(string(body), `"content_match_extract":"attribute"`) {
		t.Errorf("create body must carry a configured extract, got: %s", body)
	}

	var updateReq models.UpdateCheckJSONRequestBody
	setContentMatchExtract(&updateReq.ContentMatchExtract, types.StringValue(""), true)
	body, _ = json.Marshal(updateReq)
	if !strings.Contains(string(body), `"content_match_extract":""`) {
		t.Errorf("update body must send \"\" so a previously-set extract can be cleared, got: %s", body)
	}
}

// TestContentMatchTextClearsOnUpdate pins the same clearing contract for the
// provider: an emptied content_match_text must reach the API as "", which the
// backend normalizes to NULL. Omitting it (or sending null) would leave the
// previously saved text in place and produce a permanent plan diff.
func TestContentMatchTextClearsOnUpdate(t *testing.T) {
	var updateReq models.UpdateCheckJSONRequestBody
	updateReq.ContentMatchText = strPtr(types.StringValue("").ValueString())

	body, err := json.Marshal(updateReq)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(body), `"content_match_text":""`) {
		t.Errorf("update body must carry an explicit empty string, got: %s", body)
	}
}
