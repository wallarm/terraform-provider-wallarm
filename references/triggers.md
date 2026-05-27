# Triggers

## Resource (`wallarm_trigger`)

Bound to counter rules (`wallarm_rule_bruteforce_counter`, `_dirbust_counter`, `_bola_counter`) via `hint_tag` filters. Counters auto-clean ~30 seconds after the last trigger reference is removed.

## Read-completeness blocker

`resourceWallarmTriggerRead` (`resource_trigger.go:245-267`) finds the matching trigger in the API response but only calls `d.Set("trigger_id", ...)` and `d.Set("client_id", ...)`. It drops `template_id`, `enabled`, `name`, `comment`, `filters`, `actions`, `threshold` — every other field the API returned.

Consequences: (1) `terraform import` leaves state with only two fields — user must still hand-write the full config and risks the next `apply` overwriting live state; (2) drift from UI edits is invisible because Read never looks at those fields; (3) acceptance tests can't use `ImportStateVerify: true`.

Fix requires writing `flattenTriggerFilters`, `flattenTriggerActions`, `flattenTriggerThreshold` helpers (reverse of the existing `expandWallarmTrigger*` functions) and calling `d.Set` for every field. Nontrivial — nested TypeList/TypeSet shapes.
