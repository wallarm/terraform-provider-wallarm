# Terraform Provider

- Website: https://www.terraform.io

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) 0.15.5 or higher
-	[Go](https://golang.org/doc/install) 1.13 or higher (to build the provider)


## Building The Provider

Clone repository to: `$GOPATH/src/terraform-provider-wallarm`

```sh
$ cd $GOPATH/src/
$ git clone https://github.com/wallarm/terraform-provider-wallarm.git
```

When it comes to building you have two options:

#### `make install` and install it globally

If you don't mind installing the development version of the provider
globally, you can use `make build` in the provider directory which
builds and links the binary into your `$GOPATH/bin` directory.

```sh
$ cd $GOPATH/src/terraform-provider-wallarm
$ make install
```

#### `make build` and build a binary locally

If you want to test it locally and have a binary right near this README.md you might use the following commands. They will build a binary `terraform-provider-wallarm_v0.0.0` where v0.0.0 is the current version of the provider.

```sh
$ cd $GOPATH/src/terraform-provider-wallarm
$ make build
```

The following code is required to be defined in a module:

```hcl-terraform
terraform {
  required_version ">= 0.15.5"

  required_providers {
    wallarm = {
      source = "wallarm/wallarm"
    }
  }
}
```

then run `terraform init`

## Development The Provider

To start using this provider you have to set up you environment with the required variables:
```sh
WALLARM_API_UUID
WALLARM_API_SECRET
```
Optional:
`WALLARM_API_HOST` with the default value `https://api.wallarm.com`

Another variant is to define the provider attributes within HCL files:

```hcl
provider "wallarm" {
  api_host = var.api_host
  api_uuid = var.api_uuid
  api_secret = var.api_secret
}
```
Create the files `variables.tf` and `variables.auto.tfvars`

`variables.tf`:
```hcl
variable "api_host" {
  type    = string
  default = "https://us1.api.wallarm.com"
}

variable "api_uuid" {
  type    = string
}

variable "api_secret" {
  type    = string
}

```
`variables.auto.tfvars`:
```
api_uuid = "00000000-0000-0000-0000-000000000000"
api_secret = "000000000000000000000000000000000000000000"
```

## Using The Provider

All the examples have been divided by the resource or datasource name in the `examples` folder.

For instance, this rule configures the blocking mode for GET requests aiming to `dvwa.wallarm-demo.com`.

```hcl
resource "wallarm_rule_mode" "dvwa_mode" {
  mode =  "block"

  action {
    type = "equal"
    value = "dvwa.wallarm-demo.com"
    point = {
      header = "HOST"
    }
  }

  action {
    type = "equal"
    point = {
      method = "GET"
    }
  }
}
```
