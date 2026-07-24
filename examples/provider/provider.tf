terraform {
  required_providers {
    systeam = {
      source  = "systeampl/systeam"
      version = "~> 0.1"
    }
  }
}

provider "systeam" {
  # Base URL of the Syschecks API (no /api suffix). Or set SYSTEAM_API_URL.
  api_url = "https://syschecks.com"

  # A Personal Access Token (pat_...). Or set SYSTEAM_API_TOKEN.
  api_token = var.systeam_token
}
