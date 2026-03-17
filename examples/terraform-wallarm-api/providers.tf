provider "wallarm" {
  api_host           = var.api_host
  api_token          = var.api_token
  client_id          = var.client_id
  api_client_logging = true
}
