# v1 Release Plan

This document records the release intent for `steam-go` as it moves from `alpha` to `rc` and finally to `v1.0.0`.

## Release positioning

`steam-go v1.0.0` is intended to be released as:

> A stable Go SDK for Steam Web API with production-oriented request controls.

It is not positioned as:

- a complete Steam SDK
- a full Steam Store SDK
- a full Steam Community SDK

## RC phase goals

### v1.0.0-rc.1

Focus:

- freeze the stable public surface
- define compatibility boundaries
- publish release-governance documents
- align README, docs, and examples with the RC message

Rules:

- no new API coverage
- no new large feature directions
- only changes needed to support release governance and boundary clarity

### v1.0.0-rc.2

Focus:

- regression validation
- examples and docs consistency
- bug fixes
- release hardening

Rules:

- no new API coverage
- no major architecture changes
- only bug fixes, tests, and documentation improvements

## v1.0.0 release gate

The project should not ship `v1.0.0` until the following are true:

- the stable public surface is documented
- compatibility policy is published
- endpoint stability and endpoint coverage are documented
- README and Chinese docs align with the release positioning
- examples still match the documented public APIs
- `go test ./...`, `go vet ./...`, and `go test -race ./...` pass
- CI remains green, including `staticcheck` and `govulncheck`

## Stable surface summary

The intended stable surface for `v1.0.0` includes:

- the root `steam` package
- `NewClient(...)`
- the current `Option` system
- `Client` and `client.API.*`
- current exported service methods
- current exported request and response structs
- proxy APIs
- traffic policy APIs
- error APIs
- redaction helpers
- documented addon import paths

Details are tracked in [compatibility.md](compatibility.md).

## Out-of-scope for v1.0.0

The following should remain outside the `v1.0.0` promise:

- new public Store / Community / CDN fetch APIs
- unstable HTML parsing workflows
- browser fallback implementations
- undocumented page-derived payload shapes
- external proxy orchestration features

These may be explored after `v1.0.0`, likely through `v1.1.0+` or explicit experimental surfaces.

## Release notes drafting guidance

`v1.0.0` release notes should emphasize:

- stable official Steam Web API support
- production-oriented request controls
- compatibility policy and documentation completeness
- what the project does not promise yet

They should avoid implying that `steam-go` already provides a full public Steam Store or Steam Community scraping SDK.
