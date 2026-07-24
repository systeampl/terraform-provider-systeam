package check

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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

func TestRawToNormalized(t *testing.T) {
	if !rawToNormalized(nil).IsNull() {
		t.Error("nil raw → null")
	}
	if !rawToNormalized([]byte("null")).IsNull() {
		t.Error("JSON null → null")
	}
	if rawToNormalized([]byte(`{"a":1}`)).IsNull() {
		t.Error("non-empty JSON must not be null")
	}
}
