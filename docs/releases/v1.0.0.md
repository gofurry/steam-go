# steam-go v1.0.0

`steam-go v1.0.0` is the first stable release of the project.

It is positioned as a stable Go SDK for the official Steam Web API, with production-oriented request controls and clear release boundaries.

## Highlights

- stable root `Client` with grouped access through `client.API.*`
- typed Steam Web API support across the current public service surface
- functional options for API keys, access tokens, timeout, retry, rate limit, proxy selection, and traffic policy routing
- production-oriented request controls, including:
  - `WithSafeDefaults()`
  - retry and backoff controls
  - rate limiting
  - host/session request controls
  - sticky proxy selection
  - health-checked proxy pools
  - proxy metrics snapshots
- optional traffic-class isolation between official Web API traffic and future public store-page traffic foundations
- stable addon entrypoints for:
  - `addons/openid`
  - `addons/a2s`

## Documentation and release readiness

- published compatibility policy
- published endpoint stability guide
- published endpoint coverage guide
- published v1 release governance notes
- live smoke validation consolidated under `examples/live/`

## Project boundaries

`v1.0.0` intentionally does **not** promise:

- a full Steam Store SDK
- a full Steam Community SDK
- built-in browser fallback implementations
- long-term stability for undocumented page-derived payload shapes
- new API coverage beyond the current documented surface

## Validation

The `v1.0.0` release line is prepared on top of:

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- `staticcheck`
- `govulncheck`

## Notes

- `TrafficClassPublicStorePage` remains a policy and infrastructure foundation. It is not a promise that `v1.0.0` already ships complete public Steam store-page fetch APIs.
- Future official Steam Web API expansion is expected to resume after `v1.0.0`, starting from the `v1.1.0+` roadmap line.
