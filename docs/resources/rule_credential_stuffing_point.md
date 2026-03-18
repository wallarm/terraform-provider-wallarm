---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_credential_stuffing_point"
subcategory: "Rule"
description: |-
  Provides the "Authentication endpoints by exact location of parameters in Credential Stuffing" rule resource.
---

# wallarm_rule_credential_stuffing_point

Provides the resource to configure authentication endpoints for [Credential Stuffing](https://docs.wallarm.com/about-wallarm/credential-stuffing/) by using request point approach.

## Example Usage

```hcl
resource "wallarm_rule_credential_stuffing_point" "point1" {
  point = [["HEADER", "HOST"]]
  login_point = ["HEADER", "SESSION-ID"]
}

resource "wallarm_rule_credential_stuffing_point" "point2" {
  client_id = 123

  action {
    type = "iequal"
    point = {
        action_name = "login"
    }
  }

  point = [["HEADER", "HOST"]]
  login_point = ["HEADER", "SESSION-ID"]
}
```

## Argument Reference

* `client_id` - (optional) ID of the client to apply the rules to. The value is required for [multi-tenant scenarios][2].
* `action` - (optional) rule conditions. See the [Action Guide](../guides/action) for full documentation on action conditions, point types, and usage examples.
* `cred_stuff_type` - (optional) defines which database of compromised credentials to use. Can be: `default`, `custom`. Default value: `default`.
* `point` - (**required**) request parts to apply the rules to. See the [Point Guide](../guides/point) for the full list of possible values and examples.
* `login_point` - (**required**) request point used for specifying login parameters. For more details about request points, see wallarm [documentation][1].

**point**, **login_point**

Should be a correct point belonging to the request, that finishes by _all

Example:

Correct:

* [["post"],["form_urlencoded", "test"],["array_all"]]
* [["post"],["form_urlencoded_all"]]
* [["post"],["json_doc"],["array_all"]]
* [["header_all"]]

Incorrect:

* [["post"],["form_urlencoded", "test"]]
* [["post"]]
* [["path_all"]]
* [["header","HOST"]]

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - the action ID (The conditions to apply on request).
* `rule_type` - type of the created rule. For example, `rule_type = "credentials_point"`.

[1]: https://docs.wallarm.com/user-guides/rules/rules/#points
[2]: https://docs.wallarm.com/installation/multi-tenant/overview/
