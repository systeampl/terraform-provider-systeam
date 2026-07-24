// Package acctest holds the shared harness for Terraform acceptance tests.
//
// Acceptance tests run REAL plan/apply/destroy against a live Syschecks API —
// this is the "and it actually works" half of our promise (the parity gate in
// internal/provider proves the UI surface is covered; these prove the covered
// resources genuinely round-trip through the API). They are gated on TF_ACC so
// `go test ./...` in CI stays a fast, offline unit run.
//
// To run against production (or any environment):
//
//	export TF_ACC=1
//	export SYSTEAM_API_URL=https://syschecks.com/api
//	export SYSTEAM_API_TOKEN=pat_xxx
//	export SYSTEAM_ACC_ORG_ID=1
//	export SYSTEAM_ACC_ESCALATION_POLICY_ID=131
//	make testacc
package acctest

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/pawel-cygal/terraform-provider-systeam/internal/provider"
)

// ProtoV6ProviderFactories wires the in-process provider into the test framework
// so apply/destroy exercise the exact code Terraform core would call.
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"systeam": providerserver.NewProtocol6WithError(provider.New()),
}

// PreCheck fails fast (before any apply) if the environment isn't configured for
// a live run. Every acceptance test calls it from its PreCheck hook.
func PreCheck(t *testing.T) {
	t.Helper()
	required := []string{"SYSTEAM_API_URL", "SYSTEAM_API_TOKEN"}
	for _, k := range required {
		if os.Getenv(k) == "" {
			t.Fatalf("%s must be set for TF_ACC acceptance tests", k)
		}
	}
}

// Env returns the value of an env var or fails the test if it is unset. Use for
// per-test inputs like the org id / escalation policy id a resource needs.
func Env(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Fatalf("%s must be set for this acceptance test", key)
	}
	return v
}
