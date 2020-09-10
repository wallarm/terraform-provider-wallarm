resource "wallarm_node" "cloud_node" {
  count = 3
  hostname = "tf-${var.node_names[count.index]}"
}