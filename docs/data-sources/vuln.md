---
layout: "wallarm"
page_title: "Wallarm: wallarm_vuln"
subcategory: "Common"
description: |-
  Get details on vulnerabilities detected by the WAF node.
---

# wallarm_vuln

Use this data source to get details on [vulnerabilities][1] detected by the WAF node.

## Example usage

```hcl
# Looks up for 1000 (max) open vulnerabilities.

data "wallarm_vuln" "vulns" {

  filter {
    status = "open"
    limit = 1000
  }
}
```

```hcl
# Looks up for 1000 (max) open vulnerabilities with the offset to 1000.
# Vulnerabilities from 1001 to 2000 will be returned if they exist.

data "wallarm_vuln" "vulns" {

  filter {
    status = "open"
    limit = 1000
    offset = 1000
  }
}
```

## Argument Reference

`filter` - (Required) Filters set in the `key=value` format used to look up for vulnerability details. Possible keys:

- `status` - (Optional) Vulnerability status. Can be: `open` for active vulnerabilities, `closed` for closed vulnerabilities, `falsepositive` for vulnerabilities marked as false positive.
- `limit` - (Optional) Integer value defining the number of vulnerabilities to return. Possible values: 0-1000.
- `offset` - (Optional) Integer value defining the number from which vulnerabilities should be returned. Possible values: 0 - 9199999999999999999.

## Attributes Reference

`vulns` - Vulnerability attributes in the `key=value` format. Possible keys:

- `vuln_id` - Integer vulnerability ID.
- `wid` - String vulnerability ID.
- `status` - Vulnerability status. Can be: `open` for active vulnerabilities, `closed` for closed vulnerabilities, `falsepositive` for vulnerabilities marked as false positive.
- `type` - Vulnerability type (`ptrav`, `sqli`, `infoleak`).
- `client_id` - ID of the client with the scanned applications.
- `method` - Method of the HTTP request sent to exploit the vulnerability. 
- `domain` - Domain where the vulnerability detected.
- `path` - Path where the vulnerability detected.
- `parameter` - Parameter where the vulnerability detected.
- `title` - Vulnerability title.
- `description` - Vulnerability description.
- `additional` - Vulnerability additional information.
- `exploit_example` - Exploit example to check the vulnerability.
- `detection_method` - Method of the vulnerability detection. Can be: `active` for the vulnerability detected by the scanner or by the rechecker, `passive` for the vulnerability detected after analyzing the server responses.

[1]: https://docs.wallarm.com/user-guides/vulnerabilities/check-vuln/
