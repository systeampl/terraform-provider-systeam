package integration_key_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/systeampl/terraform-provider-systeam/internal/acctest"
	"github.com/systeampl/terraform-provider-systeam/internal/client"
)

// TestAccIntegrationKey proves the resource round-trips through the LIVE API:
// create → read-back → import → destroy. Skipped unless TF_ACC=1 (see
// internal/acctest). It needs a real org and an existing escalation policy to
// route to — supplied via SYSTEAM_ACC_ORG_ID / SYSTEAM_ACC_ESCALATION_POLICY_ID.
func TestAccIntegrationKey(t *testing.T) {
	// Read inputs without failing here — resource.Test skips before running any
	// step when TF_ACC is unset, so validation belongs in PreCheck (below),
	// which only fires once the TF_ACC gate has passed.
	orgID := os.Getenv("SYSTEAM_ACC_ORG_ID")
	policyID := os.Getenv("SYSTEAM_ACC_ESCALATION_POLICY_ID")
	name := "tf-acc-integration-key"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.Env(t, "SYSTEAM_ACC_ORG_ID")
			acctest.Env(t, "SYSTEAM_ACC_ESCALATION_POLICY_ID")
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             checkIntegrationKeyRevoked(orgID),
		Steps: []resource.TestStep{
			{
				// Create + read-back: the key exists, has a token prefix, and the
				// full secret token is captured from the create response.
				Config: integrationKeyConfig(orgID, policyID, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("systeam_integration_key.test", "id"),
					resource.TestCheckResourceAttr("systeam_integration_key.test", "name", name),
					resource.TestCheckResourceAttr("systeam_integration_key.test", "escalation_policy_id", policyID),
					resource.TestCheckResourceAttr("systeam_integration_key.test", "grouping_type", "none"),
					resource.TestCheckResourceAttrSet("systeam_integration_key.test", "token_prefix"),
					resource.TestCheckResourceAttrSet("systeam_integration_key.test", "token"),
					resource.TestCheckResourceAttr("systeam_integration_key.test", "is_active", "true"),
				),
			},
			{
				// Import: org_id:key_id reconstructs state. The one-time secret
				// token can't be re-fetched, so it's expected to be absent here.
				ResourceName:            "systeam_integration_key.test",
				ImportState:             true,
				ImportStateIdFunc:       importIDFunc("systeam_integration_key.test", orgID),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
		// Framework auto-runs Destroy at the end; a leaked key would fail CheckDestroy.
	})
}

func integrationKeyConfig(orgID, policyID, name string) string {
	return fmt.Sprintf(`
provider "systeam" {}

resource "systeam_integration_key" "test" {
  organization_id      = %s
  name                 = %q
  escalation_policy_id = %s
}
`, orgID, name, policyID)
}

// checkIntegrationKeyRevoked hits the LIVE API after destroy and asserts the key
// is gone (GetIntegrationKey treats a revoked key as gone) — proving Terraform's
// destroy actually revoked it, not just dropped it from state.
func checkIntegrationKeyRevoked(orgID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		org, err := strconv.Atoi(orgID)
		if err != nil {
			return fmt.Errorf("bad org id %q: %w", orgID, err)
		}
		c := client.NewClient(os.Getenv("SYSTEAM_API_URL"), os.Getenv("SYSTEAM_API_TOKEN"))
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "systeam_integration_key" {
				continue
			}
			id, err := strconv.Atoi(rs.Primary.Attributes["id"])
			if err != nil {
				return fmt.Errorf("bad key id in state: %w", err)
			}
			key, err := c.GetIntegrationKey(context.Background(), org, id)
			if err != nil {
				return fmt.Errorf("checking key %d after destroy: %w", id, err)
			}
			if key != nil {
				return fmt.Errorf("integration key %d still active after destroy", id)
			}
		}
		return nil
	}
}

// importIDFunc builds the "org_id:key_id" import ID from the applied state.
func importIDFunc(resourceName, orgID string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found in state", resourceName)
		}
		return fmt.Sprintf("%s:%s", orgID, rs.Primary.Attributes["id"]), nil
	}
}
