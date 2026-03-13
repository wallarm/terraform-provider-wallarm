data "wallarm_hits" "this" {
  count        = var.fetch_hits ? 1 : 0
  client_id    = var.client_id
  request_id   = var.request_id
  mode         = var.mode
  attack_types = length(var.attack_types) > 0 ? var.attack_types : null
  time         = var.time
}

# ─── Aggregation ──────────────────────────────────────────────────────────────
# Group raw hits by point_hash, collecting stamps and attack metadata per point.

locals {
  raw_hits = try(data.wallarm_hits.this[0].hits, [])

  aggregated = {
    action      = try(data.wallarm_hits.this[0].action, [])
    action_hash = try(data.wallarm_hits.this[0].action_hash, "")
    domain      = try(local.raw_hits[0].domain, "")
    path        = try(local.raw_hits[0].path, "")
    poolid      = try(local.raw_hits[0].poolid, 0)

    points = {
      for ph in distinct([for h in local.raw_hits : sha256(jsonencode(h.point_wrapped))]) :
      ph => {
        point_wrapped = [for h in local.raw_hits : h.point_wrapped if sha256(jsonencode(h.point_wrapped)) == ph][0]
        stamps        = sort(distinct(flatten([for h in local.raw_hits : try(h.stamps, []) if sha256(jsonencode(h.point_wrapped)) == ph])))
        attack_types  = distinct([for h in local.raw_hits : try(h.type, "") if sha256(jsonencode(h.point_wrapped)) == ph && try(h.type, "") != ""])
        attack_ids    = distinct([for h in local.raw_hits : try(h.attack_id, "") if sha256(jsonencode(h.point_wrapped)) == ph && try(h.attack_id, "") != ""])
        hit_ids       = distinct([for h in local.raw_hits : try(h.id, "") if sha256(jsonencode(h.point_wrapped)) == ph && try(h.id, "") != ""])
      }
    }
  }
}

# ─── Persist aggregated data in Terraform state ──────────────────────────────
# Write-once: ignore_changes keeps the original data from the first apply.
# On first apply:  input is the aggregated result from the data source.
# On subsequent:   input evaluates to empty defaults (data source count=0),
#                  but ignore_changes preserves the stored value in state.

resource "terraform_data" "hits_state" {
  input = local.aggregated

  lifecycle {
    ignore_changes = [input]
  }
}

# ─── Effective values ─────────────────────────────────────────────────────────
# Both paths are plan-time known:
#   fetch_hits=true  → data source runs during plan refresh → local.aggregated
#   fetch_hits=false → terraform_data.output read from state → known immediately

locals {
  effective = var.fetch_hits ? local.aggregated : try(terraform_data.hits_state.output, local.aggregated)
}
