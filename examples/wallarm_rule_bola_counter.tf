# Counter resources are paired with `wallarm_trigger` via `hint_tag` filters.
# Counter Create succeeds without a trigger, but the API auto-cleans counters
# ~30s after their last trigger reference is removed.

resource "wallarm_rule_bola_counter" "user_id_bola_counter" {
  action {
    type  = "iequal"
    value = "/api/users"
    point = {
      action_name = "users"
    }
  }
}
