package provider

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// The parity gate keeps this provider from silently going stale: everything the
// Syschecks product offers as declarative, customer-managed config must be
// manageable here, and everything registered here must be a documented product
// resource.
//
// The authoritative list is NOT in this file — it lives in the repo-level
// canonical manifest iac-parity/product-surface.json, which EVERY provider
// (Terraform, Pulumi, Ansible, Salt, Crossplane) checks against so they stay in
// lockstep. This test reads the Terraform column of that manifest.

const manifestRelPath = "../../../iac-parity/product-surface.json"

type surfaceManifest struct {
	Providers []string          `json:"providers"`
	Resources []surfaceResource `json:"resources"`
}

type surfaceResource struct {
	Product   string            `json:"product"`
	UIArea    string            `json:"ui_area"`
	TypeNames map[string]string `json:"type_names"`
	Providers map[string]string `json:"providers"`
	Note      string            `json:"note"`
}

func loadManifest(t *testing.T) surfaceManifest {
	t.Helper()
	path, err := filepath.Abs(manifestRelPath)
	if err != nil {
		t.Fatalf("resolving manifest path: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading canonical surface manifest at %s: %v", path, err)
	}
	var m surfaceManifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	if len(m.Resources) == 0 {
		t.Fatalf("manifest %s has no resources", path)
	}
	return m
}

// registeredResourceTypes instantiates every resource the provider registers and
// asks each for its Terraform type name via Metadata — the same call Terraform
// core makes. This reflects the ACTUAL wired-up surface, not a hand-copied list.
func registeredResourceTypes(t *testing.T) map[string]bool {
	t.Helper()
	ctx := context.Background()
	p := New()
	types := map[string]bool{}
	for _, factory := range p.Resources(ctx) {
		r := factory()
		var resp resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "systeam"}, &resp)
		if resp.TypeName == "" {
			t.Fatalf("resource %T returned an empty TypeName from Metadata", r)
		}
		types[resp.TypeName] = true
	}
	return types
}

// TestUIParity is the forcing function. It reconciles the Terraform column of the
// canonical manifest with what the provider actually registers.
func TestUIParity(t *testing.T) {
	m := loadManifest(t)
	registered := registeredResourceTypes(t)

	// Terraform types the manifest claims are covered.
	claimedCovered := map[string]surfaceResource{}
	for _, r := range m.Resources {
		status := r.Providers["terraform"]
		tfType := r.TypeNames["terraform"]
		switch status {
		case "covered":
			if tfType == "" {
				t.Errorf("resource %q is Covered for terraform but has no type_names.terraform", r.Product)
				continue
			}
			if _, dup := claimedCovered[tfType]; dup {
				t.Errorf("duplicate covered manifest entry for %q", tfType)
			}
			claimedCovered[tfType] = r
		case "planned", "unknown", "n/a", "":
			// not expected to be registered
		default:
			t.Errorf("resource %q has unknown terraform status %q", r.Product, status)
		}
	}

	// 1. Every covered product resource must actually be registered.
	for tfType, r := range claimedCovered {
		if !registered[tfType] {
			t.Errorf("PARITY GAP: product resource %q (%s) is Covered for terraform but no resource %q is registered",
				r.Product, r.UIArea, tfType)
		}
	}

	// 2. Every registered resource must be a Covered manifest entry — nothing
	//    ships without being documented in the canonical surface.
	for tfType := range registered {
		if _, ok := claimedCovered[tfType]; !ok {
			t.Errorf("UNDOCUMENTED RESOURCE: Terraform registers %q but it is not Covered in %s; add/flip its row",
				tfType, manifestRelPath)
		}
	}

	// 3. Non-covered rows must NOT already be registered (keeps the manifest
	//    honest — if it shipped, flip its column to covered).
	for _, r := range m.Resources {
		tfType := r.TypeNames["terraform"]
		if tfType == "" {
			continue
		}
		if r.Providers["terraform"] != "covered" && registered[tfType] {
			t.Errorf("STALE MANIFEST: %q is registered in Terraform but its terraform column is %q; set it to covered",
				tfType, r.Providers["terraform"])
		}
	}
}

// TestUIParityReport prints, at a glance, the whole cross-provider coverage
// matrix so `go test -run TestUIParityReport -v ./internal/provider/` answers
// "what can a customer do, and which of our IaC providers can do it too?".
func TestUIParityReport(t *testing.T) {
	m := loadManifest(t)

	rows := make([]surfaceResource, len(m.Resources))
	copy(rows, m.Resources)
	sort.SliceStable(rows, func(i, j int) bool {
		ci := rows[i].Providers["terraform"] == "covered"
		cj := rows[j].Providers["terraform"] == "covered"
		if ci != cj {
			return ci
		}
		return rows[i].Product < rows[j].Product
	})

	glyph := map[string]string{"covered": "✓", "planned": "·", "unknown": "?", "n/a": "—", "": "?"}

	// Header — one fixed-width column per provider.
	header := pad("resource", 34)
	for _, p := range m.Providers {
		header += pad(p, 12)
	}
	t.Log("Syschecks IaC cross-provider coverage")
	t.Log(header)
	t.Log("──────────────────────────────────────────────────────────────────────────")

	counts := map[string]int{}
	for _, r := range rows {
		line := pad(r.Product, 34)
		for _, p := range m.Providers {
			st := r.Providers[p]
			g := glyph[st]
			if g == "" {
				g = "?"
			}
			line += pad(g, 12)
			if st == "covered" {
				counts[p]++
			}
		}
		t.Log(line)
	}
	t.Log("──────────────────────────────────────────────────────────────")
	summary := ""
	for _, p := range m.Providers {
		summary += p + "=" + itoa(counts[p]) + "/" + itoa(len(rows)) + "  "
	}
	t.Logf("covered: %s", summary)
	t.Log("legend: ✓ covered  · planned  ? unknown (unaudited)  — n/a")
}

// pad right-pads s to at least width runes (counting runes so the ✓ glyph, a
// multi-byte rune, doesn't throw the columns off).
func pad(s string, width int) string {
	n := 0
	for range s {
		n++
	}
	for n < width {
		s += " "
		n++
	}
	return s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}
