resource "wallarm_blacklist" "blacklist" {
  ip_range = ["1.1.1.1/32"]
  application = [1]
  reason = "TEST BLACKLIST"
  time = 1
}