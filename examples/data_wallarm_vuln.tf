data "wallarm_vuln" "vulns" {

  filter {
    status = "open"
    limit = 1000
  }
}

output "vulns" {
  value = data.wallarm_vuln.vulns.vuln
}