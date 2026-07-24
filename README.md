# terraform-provider-systeam

Terraform / OpenTofu provider for the SysTeam Healthchecks API. Manage checks,
projects, status pages, maintenance windows, notification channels, on-call
schedules, escalation policies and SLOs as code.

Built on `terraform-plugin-framework` (standard Terraform Provider Plugin
Protocol v6) — works with both **Terraform** and **OpenTofu**.

## Provider config

```hcl
provider "systeam" {
  api_url   = "https://systeam.pl/api"   # or env SYSTEAM_API_URL
  api_token = var.systeam_token          # or env SYSTEAM_API_TOKEN — a PAT (pat_...)
}
```

Auth is a Personal Access Token (`Authorization: Bearer pat_...`). Use an
org-scoped PAT so resources land in the right organization.

## Resources & data sources

Resources: `systeam_check`, `systeam_check_slo`, `systeam_project`,
`systeam_status_page`, `systeam_maintenance_window`, `systeam_notification_channel`,
`systeam_oncall_schedule`, `systeam_escalation_policy`.
Data source: `systeam_organization`.

### Check types

`heartbeat`, `uptime`, `icmp`, `tcp`, `udp`, `dns`, `ftp`, `mail`, `database`,
`api_scenario`, `oidc` — full parity with the backend `CheckCreate` contract
(scalars, lists, and JSON-object fields). JSON-object fields (`http_headers`,
`http_form_login_data`, `api_scenario_steps`, `oidc_config`, `dns_records_config`,
`response_assertions`, `check_source_critical`) are set via `jsonencode({...})`
and compared semantically (whitespace / key order don't cause diffs). Write-only
secrets (`*_password`, `auth_bearer_token`, `api_scenario_secrets`) are encrypted
server-side and never read back. See `docs/integrations/CHECK_CONTRACT.md`.

#### Database check example

```hcl
resource "systeam_check" "primary_db" {
  name       = "Primary Postgres"
  type       = "database"
  project_id = systeam_project.production.id

  db_type     = "postgresql"   # postgresql | mysql | mongodb | redis
  db_host     = "db.internal"
  db_port     = 5432
  db_name     = "app"
  db_username = "monitor"
  db_password = var.db_password # write-only/Sensitive — encrypted server-side, never read back
  db_query    = "SELECT 1"

  interval     = 60
  runbook_url  = "https://wiki.example.com/runbooks/postgres"
}
```

> `db_password` (and other secret fields) are write-only: the API encrypts them
> and never returns them, so the provider does not refresh them into state — they
> won't show spurious diffs, but changing them in config will update the resource.

## OpenTofu

The provider uses the standard plugin protocol, so OpenTofu runs it unchanged.
Until it's published to a registry, install it locally via a dev override.

`~/.terraformrc` (Terraform) **or** `~/.tofurc` (OpenTofu):

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/pawel-cygal/systeam" = "/path/to/terraform-provider-systeam/bin"
  }
  direct {}
}
```

Build the binary into that dir: `go build -o bin/terraform-provider-systeam`.

### Smoke test

```bash
cd terraform-provider/examples
tofu init       # or: terraform init
tofu validate   # or: terraform validate
# tofu plan      # against a real API_URL + PAT
```

With a dev override `init` may warn that overrides are in effect — that's expected;
`validate`/`plan` use the local binary.
