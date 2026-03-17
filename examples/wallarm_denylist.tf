resource "wallarm_denylist" "denylist_ip_subnet" {
  ip_range = ["168.183.181.0/24","186.249.44.123","231.73.226.4","42.74.57.135"]
  application = [-1] # Default Application
  reason = "TEST DENYLIST SUBNET"
  time_format = "rfc3339"
  time = "2030-06-13T23:33:00+07:00"
}

resource "wallarm_denylist" "denylist_countries" {
  country = ["US", "GB"]
  application = [0] # All Applications
  reason = "TEST DENYLIST COUNTRY"
  time_format = "forever"
}

resource "wallarm_denylist" "denylist_proxy_type" {
  application = [100] 
  proxy_type = ["TOR", "VPN", "PUB"]
  reason = "TEST DENYLIST PROXY"
  time_format = "Minutes"
  time = 60
}

resource "wallarm_denylist" "denylist_datacenter" {
  application = [0] # All Applications
  datacenter = ["aws", "docean","linode", "tencent"]
  reason = "TEST DENYLIST DATACENTER"
  time_format = "Minutes"
  time = 60
}