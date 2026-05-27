# IP Lists

## Resources (`wallarm_allowlist`, `wallarm_denylist`, `wallarm_graylist`)

IP list resources are the most complex resources in the provider due to API behavior. One rule type per resource, max 1000 subnets.

**IP list Reads are config-driven, not API-driven.** `resourceWallarmIPListRead` (shared by `allowlist`, `denylist`, `graylist`) calls `ipListConfigValues(d)` to read the user's HCL values, looks those up in the cached list, and sets only `address_id` / `entry_count`. Other schema fields (`ip_range`, `country`, `datacenter`, `application`, `proxy_type`, `reason`, `time`, `time_format`) are `ForceNew` config inputs — they're never written by Read and that is correct. IP lists are effectively append-only with per-value IDs, so there's no "authoritative state" to reconcile; the user's config IS the state.

## IP list cache

Cache at `ProviderMeta` level with per-rule-type fetching and Create serialization. The `terraform-provider-caching` skill is the canonical reference for cache strategy.

## API limits

`wallarm/provider/constants.go` (authoritative):

| Constant | Value | Purpose |
|---|---|---|
| `IPListPageSize` | 1000 | IP list groups per API call |
| `IPListMaxSubnets` | 1000 | Max subnet values per IP list resource |
| `IPListCacheMaxRetries` | 3 | Cache refresh retries |
| `IPListCacheRetryDelay` | 3s | Wait between retries |
