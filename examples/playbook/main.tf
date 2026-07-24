# A multi-stage incident-response playbook, spike.sh-style, declared in Terraform.
# Fires when an inbound incident is created, then runs ordered steps.

terraform {
  required_providers {
    systeam = { source = "registry.terraform.io/systeampl/systeam" }
  }
}

provider "systeam" {}

variable "organization_id" { type = number }

resource "systeam_playbook" "sev1" {
  organization_id = var.organization_id
  name            = "SEV1 Auto-Response"
  description     = "Escalate, open a war room and set severity for high-priority inbound incidents."
  trigger_type    = "inbound_incident_created"
  trigger_conditions = jsonencode({ severity = "critical" })

  step {
    step_order  = 1
    name        = "Set priority P1"
    action_type = "set_priority"
    config      = jsonencode({ priority = "P1" })
  }

  step {
    step_order  = 2
    name        = "Open war room"
    action_type = "setup_war_room"
    config      = jsonencode({ auto_invite = true })
  }

  step {
    step_order     = 3
    name           = "Notify subscribers"
    action_type    = "notify_subscribers"
    config         = jsonencode({ message = "SEV1 declared — responders paged." })
    timeout_seconds = 60
  }
}
