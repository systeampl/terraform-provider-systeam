# terraform-provider-systeam

Terraform / OpenTofu provider for **Syschecks** (SysTeam Healthchecks) — manage
your entire monitoring stack as code: checks, projects, status pages,
maintenance windows, notification channels, on-call schedules, escalation
policies, SLOs, integration keys, teams, services, contact methods, lifecycle
watches and multi-stage incident **playbooks**.

Built on `terraform-plugin-framework` (standard Terraform Provider Plugin
Protocol v6) — works with both **Terraform** and **OpenTofu**.

## Usage

```hcl
terraform {
  required_providers {
    systeam = {
      source  = "systeampl/systeam"
      version = "~> 0.1"
    }
  }
}

provider "systeam" {
  api_url   = "https://syschecks.com"   # base URL — or env SYSTEAM_API_URL
  api_token = var.systeam_token         # a PAT (pat_...) — or env SYSTEAM_API_TOKEN
}
```

Authentication is a Personal Access Token (`Authorization: Bearer pat_...`). Use
an org-scoped PAT so resources land in the right organization.

OpenTofu users: the same configuration works unchanged — just run `tofu` instead
of `terraform`.

## Resources & data sources

| Resource | Manages |
|---|---|
| `systeam_project` | Projects that group checks |
| `systeam_check` | Monitoring checks (11 types — see below) |
| `systeam_check_slo` | SLO targets per check |
| `systeam_notification_channel` | Notification channels (Slack, email, Discord, …) |
| `systeam_escalation_policy` | Multi-step escalation policies |
| `systeam_oncall_schedule` | On-call rotation schedules |
| `systeam_status_page` | Public status pages |
| `systeam_maintenance_window` | Scheduled maintenance windows |
| `systeam_integration_key` | Inbound-events keys (Alertmanager / Grafana / Prometheus → escalation) |
| `systeam_team` | Teams for ownership & routing |
| `systeam_service` | Service catalog entries |
| `systeam_playbook` | Multi-stage incident-response automation |
| `systeam_lifecycle_watch` | Technology Watch (AI model / resource lifecycle) |
| `systeam_contact_method` | Per-user contact methods (SMS / voice / email / push) |
| `systeam_agent_registration_token` | One-time token to enrol a private/geo agent |

### Data sources

Look up resources created outside Terraform (in the UI or by another stack) and
reference them for GitOps — by `slug` for the organization, by `name` (scoped to
`organization_id`) for the rest:

| Data source | Looks up |
|---|---|
| `systeam_organization` | An organization, by slug |
| `systeam_team` | A team, by name |
| `systeam_service` | A service, by name |
| `systeam_project` | A project, by name |
| `systeam_escalation_policy` | An escalation policy, by name |
| `systeam_oncall_schedule` | An on-call schedule, by name |
| `systeam_notification_channel` | A notification channel, by name |

### Importing existing resources

Every resource supports `terraform import` (except `systeam_agent_registration_token`,
whose token is one-time and cannot be read back). Org-scoped resources use a
composite `org_id:id`; the rest take a bare numeric id:

```bash
terraform import systeam_team.core        1:42     # org_id:id
terraform import systeam_check.primary     5       # bare id
```

Two attributes cannot round-trip on import because the API never returns them:
`systeam_contact_method.value` (masked server-side) and any write-only check
secret. Set them in config; Terraform treats a change to `value` as a replace.

Full per-resource documentation is in [`docs/`](docs/) and rendered on the
[Terraform Registry](https://registry.terraform.io/providers/systeampl/systeam).

### Check types

`heartbeat`, `uptime`, `icmp`, `tcp`, `udp`, `dns`, `ftp`, `mail`, `database`,
`api_scenario`, `oidc`. JSON-object fields (`http_headers`, `api_scenario_steps`,
`oidc_config`, `dns_records_config`, `response_assertions`, …) are set via
`jsonencode({...})` and compared semantically, so whitespace and key order never
cause spurious diffs. Write-only secrets (`*_password`, `auth_bearer_token`,
`api_scenario_secrets`) are encrypted server-side and never read back into state.

### Database check example

```hcl
resource "systeam_check" "primary_db" {
  name       = "Primary Postgres"
  type       = "database"
  project_id = systeam_project.production.id

  db_type     = "postgresql" # postgresql | mysql | mongodb | redis
  db_host     = "db.internal"
  db_port     = 5432
  db_name     = "app"
  db_username = "monitor"
  db_password = var.db_password # write-only — encrypted server-side, never read back
  db_query    = "SELECT 1"

  interval    = 60
  runbook_url = "https://wiki.example.com/runbooks/postgres"
}
```

## Examples

Runnable examples live under [`examples/`](examples/):

- `examples/oncall-scenario/` — channels → schedule → multi-step escalation → integration key
- `examples/playbook/` — a spike.sh-style multi-stage incident playbook
- `examples/private-agent/` — provision a private agent end-to-end from Terraform

## Development

```bash
go build ./...
go test ./...          # unit tests
make testacc           # acceptance tests (TF_ACC) against a live API
```

Local install for testing before a release:

```hcl
# ~/.terraformrc (Terraform) or ~/.tofurc (OpenTofu)
provider_installation {
  dev_overrides {
    "systeampl/systeam" = "/path/to/terraform-provider-systeam/bin"
  }
  direct {}
}
```

## License

[Apache 2.0](LICENSE).
