# wallarm-go client library

Reference for the `wallarm-go` HTTP client module (`../wallarm-go`): the request
behaviors the provider relies on and the conventions for adding API structs. The
per-field zero-value contract that drives struct tags is in
`rules_api_fields.md §4.2`.

## 1. Overview

`wallarm-go` is a separate Go module that wraps the Wallarm HTTP API. The
provider stores a `wallarm.API` (wrapped in `CachedClient`) on `ProviderMeta`
and calls it from every CRUD path. Changes to struct field identifiers here are
breaking for the provider, which pattern-matches on the exposed Go names.

## 2. Model

The client sends JSON over HTTP, requests gzip-compressed responses, retries
transient failures, and returns a typed `APIError` on non-success. Paginated
methods reset the response slice before each page to avoid reuse bugs.

## 3. Elements

| Element | Role |
|---|---|
| `wallarm.API` | the client interface the provider consumes |
| `APIError` | typed error with `StatusCode` and `Body` fields |
| paginated `*List*` methods | page-at-a-time fetch; reset `response.Body.Objects = nil` per page |

## 4. Behavior

- **Retry** (`MaxRetries = 12`, so up to 13 attempts per request: 1 initial +
  12 retries): HTTP 423 waits 5s, 5xx waits 10s, 429 uses exponential backoff
  (1s doubling, capped at 30s).
- **Gzip responses**: the client sends `Accept-Encoding: gzip` and decompresses
  gzip responses; request bodies are uncompressed JSON.
- **Pagination safety**: the paginating methods set `response.Body.Objects = nil`
  before each `json.Unmarshal`, preventing slice reuse across pages.
- Integration tests for the retry logic are a planned addition (roadmap **WG1**).

## 5. Parameters

Not applicable - this is a client library, not a resource.

## 6. Reference data

### Naming convention for new structs

Use Go idiomatic initialisms for identifiers (`ID`, `URL`, `HTTP`, `API`,
`UUID`, `JSON`), not `Id`/`Url`/`Http` - matching the `golint`/`revive`
`var-naming` rule. JSON tags stay snake_case; only Go identifiers change
(`HTTPMethod string` with `json:"http_method"`). Renames are breaking for the
provider and downstream consumers.

### Zero-value pointer convention

A field whose valid API range includes `0` must be `*int` + `omitempty`, not
`int` + `omitempty`: `encoding/json` drops a literal `0` from `int,omitempty`,
so the field would be absent on the wire and the API rejects it as
`can't be blank`. Applied to `ActionCreate.{Rate,Burst,Delay,OverlimitTime}`.
`RspStatus` stays `int` (range starts at 400; `0` is never valid, so dropping it
via `omitempty` is correct). Analyze "is `0` a meaningful input?" before picking
the type. Full per-field classification: `rules_api_fields.md §4.2`.

## 7. References

- `rules_api_fields.md §4.2` - the pointer/int classification per field.
- Roadmap `WG1` - retry-logic integration tests.
- `rules-core.md §3.4` - `CachedClient` wrapping on `ProviderMeta`.
