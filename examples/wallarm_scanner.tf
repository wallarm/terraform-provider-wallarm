resource "wallarm_scanner" "scan" {
    element = ["1.1.1.1", "example.com", "2.2.2.2/31"]
    disabled = true
}

output "scan_id" {
  value = wallarm_scanner.scan.resource_id
}