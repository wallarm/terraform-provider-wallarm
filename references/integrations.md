# Integrations

## Eleven integration resources

`wallarm_integration_data_dog`, `_email`, `_insightconnect`, `_opsgenie`, `_pagerduty`, `_slack`, `_splunk`, `_sumologic`, `_teams`, `_telegram`, `_webhook`.

## Read-completeness blocker

All 11 integration Read functions populate only generic metadata (`integration_id`, `is_active`, `name`, `created_by`, `type`, `client_id`). They drop the type-specific fields the API returns: `active` (user-set flag), `event[]` (TypeSet of event subscriptions), and per-integration config (`emails`, `webhook_url`, `api_url`, `api_token`, `headers`, `timeouts`, `chat_data`, `integration_key`, `with_headers`, etc.).

Consequences: `terraform import` + `-generate-config-out` produces incomplete HCL (verified with `integration_email`: generated config has `null` for `active`, `emails`, and completely missing all `event {}` blocks vs the real config's 7 event subscriptions). Drift detection is also broken — UI edits to events/active/recipient lists are invisible to plan.

**Fix:** each integration's Read needs:
1. A flatten helper for the `event[]` TypeSet (reverse of existing `expandWallarmEventToIntEvents`).
2. `d.Set` calls for every type-specific config field on the API response.
3. Likely a new typed `IntegrationRead` variant in wallarm-go if the generic one doesn't return full payloads per integration type — investigate first.
