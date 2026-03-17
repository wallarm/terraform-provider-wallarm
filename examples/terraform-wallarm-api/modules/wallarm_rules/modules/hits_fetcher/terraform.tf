terraform {
  required_version = ">= 1.10.1"

  required_providers {
    wallarm = {
      source  = "wallarm/wallarm"
      version = "2.0.1"
    }
  }
}
