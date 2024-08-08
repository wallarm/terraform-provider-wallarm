resource "wallarm_integration_data_dog" "data_dog_integration" {
  name = "New Terraform DataDog Integration"
  region = "US1"
  token = "eb7ddfc33acaaacaacaca55a398"

  event {
    event_type = "hit"
    active = true
  }

  event {
    event_type = "scope"
    active = true
  }

  event {
    event_type = "system"
    active = false
  }
}
