# Provision a private agent END-TO-END with Terraform:
#
#   systeam_agent_registration_token   → mints a one-time enrollment token
#   terraform_data + local-exec        → downloads the agent binary, registers
#                                        it with the token, and runs it
#   terraform destroy                  → stops the running agent
#
# The token minting is host-agnostic. To install on a REMOTE host instead of the
# machine running Terraform, swap the local-exec for a remote-exec provisioner
# with an SSH `connection` block — or, for a fleet, let Ansible/Salt consume the
# token. The token resource stays exactly the same.
#
# NOTE: `terraform destroy` stops the agent process but cannot delete the agent
# RECORD from the backend — there is no delete-agent API endpoint yet. Until one
# exists, prune stale agent records out of band.

terraform {
  required_providers {
    systeam = {
      source = "registry.terraform.io/systeampl/systeam"
    }
  }
}

provider "systeam" {
  # SYSTEAM_API_URL (base host, e.g. https://syschecks.com) + SYSTEAM_API_TOKEN
}

variable "api_url" {
  description = "Base URL the agent registers against."
  type        = string
  default     = "https://syschecks.com"
}

variable "organization_id" {
  type = number
}

variable "agent_name" {
  type    = string
  default = "tf-private-agent"
}

variable "install_dir" {
  description = "Directory the binary, config, pidfile and log live in."
  type        = string
}

resource "systeam_agent_registration_token" "this" {
  organization_id = var.organization_id
  name            = var.agent_name
  mode            = "private"
}

resource "terraform_data" "agent" {
  # Re-install whenever a new token is minted.
  triggers_replace = systeam_agent_registration_token.this.id

  # Stashed so the destroy-time provisioner (which may only read `self`) knows
  # where the pidfile lives.
  input = {
    dir = var.install_dir
  }

  # Download → register → run (backgrounded so the provisioner returns).
  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-c"]
    environment = {
      TOKEN   = systeam_agent_registration_token.this.token
      API_URL = var.api_url
      NAME    = var.agent_name
      DIR     = var.install_dir
    }
    command = <<-EOT
      set -euo pipefail
      mkdir -p "$DIR"
      if [ ! -x "$DIR/healthcheck-agent" ]; then
        curl -fsSL "$API_URL/api/agent/binary?os=linux&arch=amd64" -o "$DIR/healthcheck-agent"
        chmod +x "$DIR/healthcheck-agent"
      fi
      "$DIR/healthcheck-agent" register -url "$API_URL" -token "$TOKEN" -name "$NAME" -config "$DIR/config.yaml"
      nohup "$DIR/healthcheck-agent" run -config "$DIR/config.yaml" > "$DIR/agent.log" 2>&1 &
      echo $! > "$DIR/agent.pid"
      echo "agent started, pid $(cat "$DIR/agent.pid")"
    EOT
  }

  # Stop the agent on destroy.
  provisioner "local-exec" {
    when        = destroy
    interpreter = ["/bin/bash", "-c"]
    command     = <<-EOT
      DIR="${self.input.dir}"
      if [ -f "$DIR/agent.pid" ]; then
        kill "$(cat "$DIR/agent.pid")" 2>/dev/null || true
        rm -f "$DIR/agent.pid"
        echo "agent stopped"
      fi
    EOT
  }
}

output "enrollment_expires_at" {
  value = systeam_agent_registration_token.this.expires_at
}
