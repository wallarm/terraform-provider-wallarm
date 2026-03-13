resource "wallarm_graylist" "graylist_ip_subnet" {
  ip_range = ["172.16.16.0/16","10.0.0.0/8","192.168.10.1/24","8.8.8.8","1.1.1.1"]
  application = [-1] # Default Application
  reason = "TEST ALLOWLIST SUBNET"
  time_format = "rfc3339"
  time = "2029-06-12T00:00:00Z"
}

resource "wallarm_graylist" "graylist_countries" {
  country = ["AF", "IT"]
  application = [0] # All Applications
  reason = "TEST GRAYLIST COUNTRY"
  time_format = "forever"
}

resource "wallarm_graylist" "graylist_proxy_type" {
  application = [100]
  proxy_type = ["TOR", "VPN", "PUB"]
  reason = "TEST GRAYLIST PROXY"
  time_format = "Minutes"
  time = 60
}

resource "wallarm_graylist" "graylist_datacenter" {
  application = [0] # All Applications
  datacenter = ["aws", "docean","linode", "tencent"]
  reason = "TEST GRAYLIST DATACENTER"
  time_format = "Hours"
  time = 1
}