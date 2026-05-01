resource "wallarm_rule_graphql_detection" "graphql_block" {
  mode = "block"

  action {
    type  = "iequal"
    value = "graphql.example.com"
    point = {
      header = "HOST"
    }
  }

  # All max_*, introspection, debug_enabled fields are Optional+Computed —
  # API defaults apply when omitted (max_depth=10, max_value_size_kb=10,
  # max_doc_size_kb=100, max_doc_per_batch=10, max_alias_size_kb=5,
  # introspection=true, debug_enabled=true). Override here if needed:
  #
  #   max_depth         = 20
  #   max_value_size_kb = 50
  #   introspection     = false
  #   debug_enabled     = false
}
