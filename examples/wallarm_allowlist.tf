resource "wallarm_allowlist" "allowlist_minutes" {
  ip_range = ["1.1.1.1/32"]
  application = [1]
  reason = "TEST ALLOWLIST"
  time_format = "Minutes"
  time = 1
}

resource "wallarm_allowlist" "allowlist_date" {
  ip_range = ["2.2.2.2/32"]
  application = [1]
  reason = "TEST ALLOWLIST"
  time_format = "RFC3339"
  time = "2026-01-02T15:04:05+07:00"
}

resource "wallarm_allowlist" "allowlist_date" {
  ip_range = ["117.69.14.54/32"]
  application = [1]
  reason = "TEST ALLOWLIST"
  time_format = "Hours"
  time = 5
}

resource "wallarm_allowlist" "allowlist_date" {
  ip_range = ["117.69.14.54/32"]
  application = [1]
  reason = "TEST ALLOWLIST"
  time_format = "Days"
  time = 3
}

resource "wallarm_allowlist" "allowlist_date" {
  ip_range = ["117.69.14.54/32"]
  application = [1]
  reason = "TEST ALLOWLIST"
  time_format = "Weeks"
  time = 3
}

resource "wallarm_allowlist" "allowlist_date" {
  ip_range = ["117.69.14.54/32"]
  application = [1]
  reason = "TEST ALLOWLIST"
  time_format = "Months"
  time = 3
}
