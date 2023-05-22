resource "wallarm_list" "graylist_minutes" {
  ip_range = ["1.1.1.1/32"]
  application = [1]
  reason = "TEST GRAYLIST"
  time_format = "Minutes"
  time = 1
}

resource "wallarm_graylist" "graylist_date" {
  ip_range = ["2.2.2.2/32"]
  application = [1]
  reason = "TEST GRAYLIST"
  time_format = "RFC3339"
  time = "2026-01-02T15:04:05+07:00"
}
