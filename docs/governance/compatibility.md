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
- The grouped `client.Web.*` access pattern and its exported request and response structs
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
- documented grouping under `client.Web.*`

Bug fixes, validation tightening, and internal implementation changes are allowed as long as they do not break the documented stable surface.

## Not covered by the v1 promise

The following areas are intentionally outside the `v1.0.0` compatibility guarantee unless a future document says otherwise:

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

## v1.1.0 unofficial web surfaces

The following are stable Go APIs but volatile upstream surfaces:

- `client.Web.Storefront`
- `client.Web.Community`
- `client.Web.Market`

Breaking changes caused by upstream payload drift should be handled by preserving typed outer structures and moving volatile subtrees to `json.RawMessage` where possible.

## Experimental and preview areas

The following capabilities may exist in the repository, but should be treated as non-coverage or preview-oriented foundations rather than as a promise of built-in productized fetch APIs:

- `TrafficClassPublicStorePage`
- `TrafficClassCommunityWeb`
- `TrafficClassMarketWeb`
- public store-page header profiles
- Referer strategies
- short cache and block detection infrastructure
- per-class transport hooks for future TLS customization or browser-backed execution

These configuration APIs are stable as supporting infrastructure for the existing `client.Web.*` surface, but they should not be read as a promise that every future Steam web flow will be productized in the SDK.

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
