# Full on-call scenario, declared end-to-end in Terraform:
#
#   notification channels  →  on-call schedule  →  multi-step escalation policy
#   →  integration key that external monitoring (Alertmanager/Grafana/Prometheus)
#      POSTs alerts to.
#
# This is the "prepare the whole scenario from TF" case: a multi-stage escalation
# (page the on-call schedule immediately, then a person after 5 min, then a
# fallback channel after 15 min) plus the inbound key that feeds it.

terraform {
  required_providers {
    systeam = {
      source = "registry.terraform.io/pawel-cygal/systeam"
    }
  }
}

provider "systeam" {
  # api_url / api_token via SYSTEAM_API_URL / SYSTEAM_API_TOKEN env vars.
  # api_url is the base host, e.g. https://syschecks.com (no /api suffix).
}

variable "organization_id" {
  type = number
}

variable "primary_user_id" {
  description = "User paged at escalation step 2."
  type        = number
}

variable "discord_webhook" {
  type      = string
  sensitive = true
}

# --- Notification channels -------------------------------------------------
resource "systeam_notification_channel" "discord" {
  name            = "On-call Discord"
  channel_type    = "discord"
  organization_id = var.organization_id
  config = {
    webhook_url = var.discord_webhook
  }
}

# --- On-call schedule ------------------------------------------------------
resource "systeam_oncall_schedule" "primary" {
  name            = "Primary Rotation"
  organization_id = var.organization_id
  timezone        = "Europe/Warsaw"
}

# --- Multi-step escalation policy -----------------------------------------
# Ordered steps: each `delay_minutes` is how long to wait before firing THIS
# step if the incident is still unacknowledged.
resource "systeam_escalation_policy" "primary" {
  name            = "Primary Escalation"
  organization_id = var.organization_id

  # Step 1 (t+0): page whoever is on-call right now.
  step = [
    {
      step_order         = 1
      delay_minutes      = 0
      target_type        = "schedule"
      target_schedule_id = systeam_oncall_schedule.primary.id
    },
    # Step 2 (t+5m): still unacked → page a specific person directly.
    {
      step_order     = 2
      delay_minutes  = 5
      target_type    = "user"
      target_user_id = var.primary_user_id
    },
    # Step 3 (t+15m): last resort → shout into the Discord channel.
    {
      step_order        = 3
      delay_minutes     = 15
      target_type       = "channel"
      target_channel_id = systeam_notification_channel.discord.id
    },
  ]
}

# --- Inbound integration key ----------------------------------------------
# Point Alertmanager/Grafana at:
#   https://syschecks.com/api/events/alertmanager/${token}
resource "systeam_integration_key" "monitoring" {
  organization_id      = var.organization_id
  name                 = "Prometheus / Grafana"
  escalation_policy_id = systeam_escalation_policy.primary.id

  # Group a storm of alerts into one incident within a 5-minute window.
  grouping_type           = "time_window"
  grouping_window_seconds = 300
}

output "alertmanager_webhook_url" {
  description = "Put this in your Alertmanager/Grafana contact point."
  value       = "https://syschecks.com/api/events/alertmanager/${systeam_integration_key.monitoring.token}"
  sensitive   = true
}
