resource "wallarm_global_mode" "global_block" {
  filtration_mode = "default" # Global filtration mode
  rechecker_mode = "off" # Threat Replay
  overlimit_time = 1000 # Default recommended
  overlimit_mode = "monitoring" # Default
}
