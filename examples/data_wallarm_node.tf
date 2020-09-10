data "wallarm_node" "waf" {

  filter {
    type = "cloud_node"
    # hostname = "4f5f7b48bf13"
    # uuid = "b16156f9-33d2-491e-a584-513521d312db"
  }
}

output "waf_nodes" {
  value = data.wallarm_node.waf.nodes
}