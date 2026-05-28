terraform {
  required_providers {
    wallarm = {
      source = "wallarm/wallarm"
    }
  }
}

provider "wallarm" {
  api_token = var.wallarm_api_token
  api_host  = var.wallarm_api_host
}

variable "wallarm_api_token" {
  type      = string
  sensitive = true
}

variable "wallarm_api_host" {
  type    = string
  default = "https://api.wallarm.com"
}

# Singleton per client_id. Settings → API Discovery in the console.
resource "wallarm_api_discovery_config" "this" {
  enabled                  = true
  apply_extended_filter    = true
  type_detection_threshold = 0.5
  pii_detection_threshold  = 0.1
  disabled_apps            = []

  protocols {
    rest    = true
    graphql = true
    soap    = true
    grpc    = true
    mcp     = true
  }

  endpoint_stability {
    min_count = 2
    min_time  = 300
  }
}
