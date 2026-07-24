terraform {
  required_providers {
    systeam = {
      source = "registry.terraform.io/systeampl/systeam"
    }
  }
}

provider "systeam" {
  api_url   = "https://syschecks.com"
  api_token = var.systeam_token
}

variable "systeam_token" {
  type      = string
  sensitive = true
}

variable "slack_webhook" {
  type      = string
  sensitive = true
}

data "systeam_organization" "main" {
  slug = "systeam"
}

resource "systeam_project" "production" {
  name            = "Production"
  organization_id = data.systeam_organization.main.id
}

resource "systeam_check" "api_health" {
  name       = "API Health"
  type       = "uptime"
  project_id = systeam_project.production.id
  url        = "https://api.example.com/healthz"
  interval   = 60
  timeout    = 10
}

resource "systeam_notification_channel" "slack_alerts" {
  name            = "Slack Alerts"
  channel_type    = "slack"
  organization_id = data.systeam_organization.main.id
  config = {
    webhook_url = var.slack_webhook
  }
}

resource "systeam_status_page" "public" {
  name            = "Status"
  slug            = "status"
  organization_id = data.systeam_organization.main.id
  check_ids       = [systeam_check.api_health.id]
}

resource "systeam_check_slo" "api_slo" {
  check_id          = systeam_check.api_health.id
  target_percentage = 99.9
  window_days       = 30
}

variable "db_password" {
  type      = string
  sensitive = true
  default   = ""
}

resource "systeam_check" "primary_db" {
  name        = "Primary Postgres"
  type        = "database"
  project_id  = systeam_project.production.id
  db_type     = "postgresql"
  db_host     = "db.internal"
  db_port     = 5432
  db_name     = "app"
  db_username = "monitor"
  db_password = var.db_password
  db_query    = "SELECT 1"
  interval    = 60
  runbook_url = "https://wiki.example.com/runbooks/postgres"
}

resource "systeam_escalation_policy" "default" {
  name            = "Default Escalation"
  organization_id = data.systeam_organization.main.id

  step = [{
    step_order        = 0
    delay_minutes     = 0
    target_type       = "channel"
    target_channel_id = systeam_notification_channel.slack_alerts.id
  }]
}
