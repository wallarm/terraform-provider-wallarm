# Counter resources are paired with `wallarm_trigger` via `hint_tag` filters.
# Counter Create succeeds without a trigger, but the API auto-cleans counters
# ~30s after their last trigger reference is removed.

resource "wallarm_rule_bola_counter" "users_endpoint_counter" {
  # Scope: requests where path[0] = "api" and the last path segment = "users".
  # `action_name` and `path` are PointValuePoints — the value lives inside the
  # point map; the sibling `value` field must be empty.
  action {
    type  = "equal"
    value = "api"
    point = {
      path = 0
    }
  }
  action {
    type = "equal"
    point = {
      action_name = "users"
    }
  }
}
