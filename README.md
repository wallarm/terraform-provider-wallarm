# Terraform Provider

- Website: https://www.terraform.io

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x+
-	[Go](https://golang.org/doc/install) 1.14+ (to build the provider plugin)


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

#### terraform 0.12 and earlier

*You have to either sideload a binary with the correct name. Terraform searches for plugins in the format of:*

```
terraform-<TYPE>-<NAME>
````

*In the case above, the plugin is of type `provider` and of name `wallarm` as it is required by Terraform.*

#### terraform 0.13+

The following code is required to be defined in a module:

```hcl-terraform
terraform {
  required_providers {
    wallarm = {
      source = "wallarm/wallarm"
    }
  }
}
```

then run `terraform init`

## Development The Provider

To start using this provider you should set up you environment with the required variables:
```sh
WALLARM_API_UUID
WALLARM_API_SECRET
```
Optional:
`WALLARM_API_HOST` with default value `https://api.wallarm.com`

Another method to do the same is to defined a provider via `.tf` files:

```hcl
provider "wallarm" {
  api_host = var.api_host
  api_uuid = var.api_uuid
  api_secret = var.api_secret
  client_id = 1
}
```
**DO NOT** forget to create the files `variables.tf` and `variables.auto.tfvars` which is by agreement a secret and is exported nowhere as it contains sensitive information.

Example of `variables.tf`:
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
And corresponding `variables.auto.tfvars`:
```
api_uuid = "00000000-0000-0000-0000-000000000000"
api_secret = "000000000000000000000000000000000000000000"
```

### Assistance commands

*It is assumed that some of the following commands will be removed*

#### `make apply` and apply local `.tf` files

This is a testing command which will be removed once the development stage is over. It helps to stick to business.
It builds, initializes the provider and applies changes in a plan.

#### `make destroy` and destroy created resources

It builds, initializes the provider and destroys the created resources.


## Using The Provider

Some of the examples are divided by the resource/data source name, however, most of them still defined in the `examples/main.tf`.

For instance, this rule configures the blocking mode for GET requests to `dvwa.wallarm-demo.com`.

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