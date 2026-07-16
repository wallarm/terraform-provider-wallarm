# Infrastructure resources

Reference for the account- and node-level resources that are not rules:
`wallarm_tenant`, `wallarm_node`, `wallarm_user`, `wallarm_application`,
`wallarm_global_mode`. Full field lists are the registry docs
(`docs/resources/*.md`); this doc is the shared model and the delete/Read
gotchas.

## 1. Overview

These resources manage the account itself rather than traffic rules:
multi-tenancy (`tenant`), filtering nodes (`node`), user accounts (`user`), the
application/pool registry (`application`), and the account-wide filtration mode
(`global_mode`).

## 2. Model

| Resource | Manages |
|---|---|
| `wallarm_tenant` | a tenant (client) account under a partner account |
| `wallarm_node` | a filtering node, and its default application |
| `wallarm_user` | a user account within a client |
| `wallarm_application` | an application/pool (`app_id`) other resources scope to |
| `wallarm_global_mode` | account-wide filtration + rechecker + overlimit settings |

## 3. Elements

Each resource wraps the matching wallarm-go client calls. `application` is the
scoping unit referenced by rules and IP lists via `app_id`; `global_mode` is a
singleton-style account setting.

## 4. Behavior

- **`wallarm_tenant` delete safety.** Delete disables the tenant first, and
  permanently deletes only when `prevent_destroy = false` **and**
  `WALLARM_ALLOW_CLIENT_DELETE` is set (`resource_tenant.go:211-218`); otherwise
  it logs a warning and no-ops, leaving the tenant disabled but not deleted.
- **`wallarm_node`.** `partner_mode` is `Optional+Computed` and **not set by
  Read**, so out-of-band toggles are not drift-detected (roadmap **INF2**). The
  node schema has no `app_id` field.
- **`wallarm_user`.** Accounts are created with `username`, which equals `email`
  per the domain model. `email` is `Required` but **not set by Read**, so after
  import the user must re-add `email = "..."` by hand (roadmap **INF1**; Read
  could mirror `email = username`).
- **`wallarm_application`.** Carries `app_id`, `client_id`, `name`; the default
  application (`app_id = -1`) is protected from deletion
  (`resource_application.go:146-149`).
- **`wallarm_global_mode`.** Sets `filtration_mode`, `rechecker_mode`,
  `overlimit_time`, `overlimit_mode` for the account (§6).

## 5. Parameters

Key fields (full shapes in the registry docs):

| Resource | Fields |
|---|---|
| `wallarm_tenant` | `name`, `prevent_destroy`, computed client id |
| `wallarm_node` | `hostname`, `node_uuid`, `token`, `partner_mode` (no `app_id`) |
| `wallarm_user` | `email` (== `username`), role, `client_id` |
| `wallarm_application` | `app_id` (`-1` = protected default), `client_id`, `name` |
| `wallarm_global_mode` | `filtration_mode`, `rechecker_mode`, `overlimit_time`, `overlimit_mode`, `client_id` |

## 6. Reference data

`wallarm_global_mode` enums (`resource_global_mode.go`):

| Field | Values |
|---|---|
| `filtration_mode` | `default` / `monitoring` / `block` / `safe_blocking` / `off` |
| `rechecker_mode` | `on` / `off` |
| `overlimit_mode` | `blocking` / `monitoring` |

- `wallarm_tenant`: env `WALLARM_ALLOW_CLIENT_DELETE` (any non-empty value)
  gates permanent delete.
- `wallarm_application`: `app_id = -1` (the default application) is undeletable.

## 7. References

- Roadmap `INF1` (`user.email` Read), `INF2` (`node.partner_mode` drift).
- `docs/resources/{tenant,node,user,application,global_mode}.md` - full fields.
