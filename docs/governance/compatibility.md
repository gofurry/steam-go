# Compatibility Policy

This document summarizes the compatibility policy for `steam-go` v1.x. The normative rules live in [the public compatibility contract](../../contracts/compatibility.md); upstream assumptions live in [the Steam upstream contract](../../contracts/upstream.md).

## Scope

`steam-go` is positioned as:

> A stable Go SDK for Steam APIs, scoped read-only Web surfaces, and optional addons.

The stability promise covers documented public Go APIs. It does not imply that every Steam-hosted endpoint is officially documented or that its upstream payload is stable.

## Stable in v1.x

Unless otherwise documented, the following are intended to be covered by the `v1` compatibility promise:

- Root package `steam`
- Public local-only `steamid` package
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

Bug fixes and internal changes should preserve valid callers. Security or correctness fixes may tighten unsafe or objectively invalid behavior; document these exceptions explicitly, as required by the compatibility contract.

## Not covered by the v1 promise

The following areas are outside the v1 compatibility guarantee unless explicitly documented otherwise:

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

## Unofficial web surfaces

The following are stable Go APIs but volatile upstream surfaces:

- `client.Web.Storefront`
- `client.Web.Community`
- `client.Web.Market`

Handle upstream drift through defensive decoding and additive models while preserving existing method signatures, exported field types, JSON tags, and raw fields. Use `json.RawMessage` for new volatile subtrees when appropriate.

## Request-control infrastructure

The following documented configuration APIs are stable infrastructure for the current Web surfaces:

- `TrafficClassPublicStorePage`
- `TrafficClassCommunityWeb`
- `TrafficClassMarketWeb`
- public store-page header profiles
- Referer strategies
- short cache and block detection infrastructure
- per-class transport hooks for future TLS customization or browser-backed execution

These configuration APIs are stable as supporting infrastructure for the existing `client.Web.*` surface, but they should not be read as a promise that every future Steam web flow will be productized in the SDK.

## Compatible evolution

The project can continue to grow additively in v1.x through:

- new official Steam Web API methods in `v1.x`
- additional typed coverage for stable official payloads
- optional addons with explicit scope and upstream boundaries
