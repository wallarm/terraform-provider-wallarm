resource "wallarm_allowlist" "allowlist_ip_subnet" {
  ip_range = ["172.16.16.0/16","10.0.0.0/8","192.168.10.1/24","8.8.8.8","1.1.1.1"]
  application = [-1] # Default Application
  reason = "TEST ALLOWLIST SUBNET"
  time_format = "rfc3339"
  time = "2029-06-12T00:00:00Z"
}

resource "wallarm_allowlist" "allowlist_countries" {
  country = ["US", "GB"]
  application = [0] # All Applications
  reason = "TEST ALLOWLIST COUNTRY"
  time_format = "forever"
}

resource "wallarm_allowlist" "allowlist_proxy_type" {
  application = [100]
  proxy_type = ["TOR", "VPN", "PUB"]
  reason = "TEST ALLOWLIST PROXY"
  time_format = "Minutes"
  time = 60
}

resource "wallarm_allowlist" "allowlist_datacenter" {
  application = [0] # All Applications
  datacenter = ["aws", "docean","linode", "tencent"]
  reason = "TEST ALLOWLIST DATACENTER"
  time_format = "Minutes"
  time = 30
}