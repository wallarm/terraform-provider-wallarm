# Minimal enforcement: spec attached, policy in pure observation mode.
# Every violation is logged, nothing is blocked — useful during the initial
# rollout of a new spec while you verify it matches real traffic.
resource "wallarm_api_spec" "petstore" {
  client_id       = 6039
  title           = "Petstore"
  file_remote_url = "https://raw.githubusercontent.com/acme/petstore/main/openapi.yaml"
  domains         = ["petstore.example.com"]
  instances       = [1]
}

resource "wallarm_api_spec_policy" "petstore_monitor" {
  client_id   = wallarm_api_spec.petstore.client_id
  api_spec_id = wallarm_api_spec.petstore.api_spec_id
  # All defaults = "monitor", so only enabled needs to be set explicitly here.
  enabled = true
}

# Strict block on undefined endpoints, everything else observed.
# Scope limits enforcement to the production host so staging traffic is
# unaffected.
resource "wallarm_api_spec_policy" "petstore_block_undefined" {
  client_id   = wallarm_api_spec.petstore.client_id
  api_spec_id = wallarm_api_spec.petstore.api_spec_id

  undefined_endpoint_mode = "block"
  missing_auth_mode       = "block"

  # Explicit defaults shown for readability.
  undefined_parameter_mode     = "monitor"
  missing_parameter_mode       = "monitor"
  invalid_parameter_value_mode = "monitor"
  invalid_request_mode         = "monitor"

  conditions {
    type  = "iequal"
    value = "petstore.example.com"
    point = {
      header = "HOST"
    }
  }
}

# Disabled policy preserved: all settings configured, enforcement paused.
# This "soft off" pattern keeps the desired modes in state so a later
# enabled = true flip restores the same policy without re-reviewing every
# setting.
resource "wallarm_api_spec_policy" "petstore_paused" {
  client_id   = wallarm_api_spec.petstore.client_id
  api_spec_id = wallarm_api_spec.petstore.api_spec_id

  enabled = false

  undefined_endpoint_mode      = "block"
  undefined_parameter_mode     = "monitor"
  missing_parameter_mode       = "block"
  invalid_parameter_value_mode = "block"
  missing_auth_mode            = "block"
  invalid_request_mode         = "block"
}
