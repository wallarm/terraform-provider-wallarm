resource "wallarm_api_spec" "api_spec" {
  client_id          = 1
  title              = "Example API Spec"
  description        = "This is an example API specification created by Terraform."
  file_remote_url    = "https://raw.githubusercontent.com/OAI/OpenAPI-Specification/main/examples/v3.0/api-with-examples.yaml"
  regular_file_update = true
  api_detection      = true
  domains = ["ex.com"]
  instances = [1]
}