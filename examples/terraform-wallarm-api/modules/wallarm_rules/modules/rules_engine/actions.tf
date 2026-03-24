# ─── Action map ───────────────────────────────────────────────────────────────
# Maps conditions_hash → { action_id, dir_name, conditions } for all known actions.
# Built from two sources:
#   A) .action.yaml files on disk (from fileset)
#   B) generated_rules (from hits, known at plan time)
# Default action is always included.

locals {
  _default_action_hash = "5b8b61bd5ed79de9b3d130436a1e5a63ec663e224557ccb981bbb491a891b4dc"

  # Source A: from .action.yaml files
  _action_map_from_files = {
    for path, entry in local._action_files :
    (try(entry.data.conditions_hash, "") != "" ? entry.data.conditions_hash : path) => {
      action_id  = try(entry.data.action_id, null)
      conditions = try(entry.data.conditions, [])
      dir_name   = try(entry.data.dir_name, dirname(entry.filename))
    }
    if can(entry.data.conditions)
  }

  # Source B: from generated_rules (plan-time known, enables single-apply)
  _action_map_from_generated = merge(
    # Default action
    {
      (local._default_action_hash) = {
        action_id  = null
        conditions = []
        dir_name   = "_default"
      }
    },
    # Actions from hits/imports — deduplicate by hash (multiple rules share same action)
    { for hash, entries in {
        for r in var.generated_rules :
        r._action_hash => {
          action_id  = null
          conditions = r._action_conditions
          dir_name   = basename(trimprefix(r._config_dir, "./"))
        }...
        if r._action_hash != ""
      } : hash => entries[0]
    },
  )

  # Merge: files take priority (may have action_id from discovery)
  action_map = merge(local._action_map_from_generated, local._action_map_from_files)
}
