---
layout: "wallarm"
page_title: "Wallarm: wallarm_vuln"
subcategory: "Common"
description: |-
  Get details on vulnerabilities detected by Wallarm.
---

# wallarm_vuln

Use this data source to get details on [vulnerabilities][1] detected by  Wallarm.

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

`filter` - (**required**) filters set in the `key=value` format used to look up for vulnerability details. Possible keys:

- `status` - (optional) vulnerability status. Can be: `open` for active vulnerabilities, `closed` for closed vulnerabilities, `falsepositive` for vulnerabilities marked as false positives.
- `limit` - (optional) integer value defining the number of vulnerabilities to return. Possible values: 1-1000.
- `offset` - (optional) integer value defining the number from which vulnerabilities should be returned. Possible values: 0 - 4611686018427387903.

## Attributes Reference

`vulns` - vulnerability attributes in the `key=value` format. Possible keys:

- `vuln_id` - integer vulnerability ID.
- `wid` - string vulnerability ID.
- `status` - vulnerability status. Can be: `open` for active vulnerabilities, `closed` for closed vulnerabilities, `falsepositive` for vulnerabilities marked as false positives.
- `type` - vulnerability type (`ptrav`, `sqli`, `infoleak` and [other types][2]).
- `client_id` - ID of the client with the scanned applications.
- `method` - method of the HTTP request sent to exploit the vulnerability. 
- `domain` - domain where the vulnerability detected.
- `path` - path where the vulnerability detected.
- `parameter` - parameter where the vulnerability detected.
- `title` - vulnerability title.
- `description` - vulnerability description.
- `additional` - vulnerability additional information.
- `exploit_example` - exploit example to check the vulnerability.
- `detection_method` - method of the vulnerability detection. Can be: `active` for the vulnerability detected by the scanner or by the Active threat verification component, `passive` for the vulnerability detected after analyzing the server responses.

[1]: https://docs.wallarm.com/user-guides/vulnerabilities/check-vuln/
[2]: https://docs.wallarm.com/attacks-vulns-list/
