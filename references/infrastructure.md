# Other Infrastructure Resources

## Tenant (`wallarm_tenant`)

Delete safety: disables first, only permanently deletes if `prevent_destroy=false` AND `WALLARM_ALLOW_CLIENT_DELETE=1`.

## Node (`wallarm_node`)

Default application (`app_id=-1`) is protected from deletion.

## User (`wallarm_user`)

User accounts are created with `username` (which equals `email` per the domain model).
