# wallarm-go Client Library

The `wallarm-go` library (`../wallarm-go`) is the HTTP client for the Wallarm API.

## Features

- `APIError` struct with `StatusCode` and `Body` fields
- Automatic retry: 423 (5s × 12), 5xx (10s × 12), 429 (exponential backoff × 12)
- Gzip compression on all requests (~19x reduction)
- All paginated methods set `response.Body.Objects = nil` before each `json.Unmarshal` (prevents slice reuse bugs)

## Naming convention for new API structs

Follow Go idiomatic initialism style when adding or editing struct fields in `wallarm-go` — `ID`, `URL`, `HTTP`, `API`, `UUID`, `JSON`, etc., not `Id`/`Url`/`Http`. This matches the `golint`/`revive` "var-naming" rule and the rest of the upstream Go ecosystem. JSON tags keep their snake_case (`json:"http_method"`), only Go identifiers change. Example: `HTTPMethod string `json:"http_method"``. Provider-side code and downstream consumers pattern-match on the exposed Go identifier — renames here are breaking.

## Zero-value pointer convention

Fields whose API range includes zero as a valid value (`Rate 0..1000`, `Burst 0..N`, `Delay 0..N`, `OverlimitTime 0..N`) MUST be `*int+omitempty` in the Go struct, NOT `int+omitempty` — Go's `encoding/json` drops a literal zero from `int+omitempty`, so the wire payload would be missing the field and the API rejects with `can't be blank`. Pattern is applied to `ActionCreate.{Rate,Burst,Delay,OverlimitTime}`. `RspStatus` keeps `int` because its valid range starts at 400 — zero is not a legal value, so dropping absent via `omitempty` is correct. New API fields must be analyzed for "is zero a meaningful input" before picking the type.
