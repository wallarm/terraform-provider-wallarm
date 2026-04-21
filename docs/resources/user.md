---
layout: "wallarm"
page_title: "Wallarm: wallarm_user"
subcategory: "Common"
description: |-
  Provides the resource to manage users of the company.
---

# wallarm_user

Provides the resource to manage [users][1] of the company, their profile and permission levels.

## Example Usage

```hcl
# Creates a new company user "Terraform Deploy"
# with the role "deploy" and other parameters

resource "wallarm_user" "user" {
  email = "testuser+6039@wallarm.com"
  realname = "Terraform Deploy"
  permissions = "deploy"
  password = "1234ABC!@#"
  phone = "+1 900 123 45 67"
}
```

## Argument Reference

* `email` - (**required**) user email. The value will be used as the username for authentication in Wallarm Console.
* `realname` - (**required**) the first and last name of the user.
* `permissions` - (**required**) user role. Can be one of: `admin`, `admin_ext`, `analytic`, `auditor`, `deploy`, `partner_admin`, `partner_admin_ext`, `partner_analytic`, `partner_auditor`. The `admin_ext` and `partner_admin_ext` roles add the ability to manage `disable_stamp` rules (false positive suppression by signature). Roles description is available in the [Wallarm official documentation](https://docs.wallarm.com/user-guides/settings/users/#user-roles).
* `password` - (optional) user password. If the value is not specified, it will be generated automatically and returned in the attribute `generated_password`.
* `phone` - (optional) user phone number.
* `client_id` - (optional) ID of the client (tenant). The value is required for [multi-tenant scenarios][2].

## Attributes Reference

* `generated_password` - Automatically generated password for a new user.
* `user_id` - Unique ID of the created user.
* `username` - Username for authentication in Wallarm Console. User email is used as the username.

## Import

```
$ terraform import wallarm_user.example 1111/2222
```

* `1111` - Client ID.
* `2222` - User ID.

[1]: https://docs.wallarm.com/user-guides/settings/users/
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
