# Minimal URL-hosted OpenAPI specification.
resource "wallarm_api_spec" "petstore_minimal" {
  client_id       = 6039
  title           = "Petstore"
  file_remote_url = "https://raw.githubusercontent.com/acme/petstore/main/openapi.yaml"
  domains         = ["petstore.example.com"]
  instances       = [1]
}

# Spec fetched from a private URL with authentication headers,
# applied to multiple domains.
resource "wallarm_api_spec" "billing" {
  client_id       = 6039
  title           = "Billing API"
  description     = "Internal billing service, v2"
  file_remote_url = "https://specs.internal.example.com/billing/openapi.yaml"
  domains         = ["billing.example.com", "billing-staging.example.com"]
  instances       = [2, 7]

  auth_headers {
    key   = "X-Source-Token"
    value = var.billing_source_token
  }

  auth_headers {
    key   = "X-Tenant"
    value = "billing"
  }
}

# Spec with hourly refresh and API discovery enabled.
resource "wallarm_api_spec" "storefront" {
  client_id           = 6039
  title               = "Storefront API"
  description         = "Customer-facing storefront, continuously updated"
  file_remote_url     = "https://raw.githubusercontent.com/acme/storefront/main/openapi.yaml"
  regular_file_update = true
  api_detection       = true
  domains             = ["shop.example.com"]
  instances           = [3]
}
