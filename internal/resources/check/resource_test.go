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
