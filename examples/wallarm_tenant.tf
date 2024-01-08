resource "wallarm_tenant" "tenant" {
  name = "Tenant"
}

resource "wallarm_tenant" "subtenant" {
  name = "Sub Tenant"
  client_id = 123
}
