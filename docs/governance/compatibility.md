# Compatibility Policy

This document defines what `steam-go` intends to keep stable starting with `v1.0.0`.

## Scope

`steam-go v1.0.0` is positioned as:

> A stable Go SDK for Steam Web API with production-oriented request controls.

This stability promise is centered on the official `api.steampowered.com` client surface and the request-control tool layer already exposed by the root package.

## Stable in v1.0.0

Unless otherwise documented, the following are intended to be covered by the `v1` compatibility promise:

- Root package `steam`
- `NewClient(...)`
- The existing `Option` system
- `Client` and the grouped `client.API.*` access pattern
- Existing exported service method signatures
- Existing exported request and response structs
- Proxy-related public APIs
- Traffic-policy related public APIs
- Error types and error kinds
- URL redaction helpers
- Public addon import paths that are already documented

## Stable behavior expectations

For the stable surface above, `v1` should preserve:

- exported names
- method signatures
- option semantics at a behavior level
- documented error kinds
- documented grouping under `client.API.*`

Bug fixes, validation tightening, and internal implementation changes are allowed as long as they do not break the documented stable surface.

## Not covered by the v1 promise

The following areas are intentionally outside the `v1.0.0` compatibility guarantee unless a future document says otherwise:

- future Store / Community / CDN fetch entrypoints
- HTML parsing rules and page-shape assumptions
- browser-backed fallback implementations
- undocumented web payload structures
- fine-grained shape of high-volatility raw JSON subtrees
- complex external proxy-pool management strategies
- future experimental packages or addons

## Raw payload policy

`steam-go` uses three payload strategies:

- stable official payloads should prefer typed structs
- large or fast-changing subtrees may use `typed outer + json.RawMessage`
- high-volatility payloads should remain raw until their shape is stable enough to promote

The exact internal shape of `json.RawMessage` subtrees is not part of the `v1` compatibility promise unless explicitly documented as stable.

## Experimental and preview areas

The following capabilities may exist in the repository, but should be treated as non-coverage or preview-oriented foundations rather than as a promise of built-in productized fetch APIs:

- `TrafficClassPublicStorePage`
- public store-page header profiles
- Referer strategies
- short cache and block detection infrastructure
- per-class transport hooks for future TLS customization or browser-backed execution

These are stable as root-package configuration APIs only to the extent already documented, but they do not imply that `steam-go v1.0.0` ships a full public Steam Store page SDK.

## Before and after v1.0.0

Before `v1.0.0`, the project will prioritize:

- freezing the stable surface
- documenting boundaries
- tightening tests and examples
- avoiding new API expansion

After `v1.0.0`, the project can continue to grow compatibly through:

- new official Steam Web API methods in `v1.x`
- additional typed coverage for stable official payloads
- optional experimental addons for high-volatility web surfaces
