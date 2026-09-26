# Endpoint Stability

This document explains API stability in `steam-go` v1.x. The normative rules are the [compatibility contract](../../contracts/compatibility.md) and [upstream contract](../../contracts/upstream.md).

## Stability levels

### Stable

Stable means the exported API shape is intended to be covered by the `v1` compatibility promise.

This includes:

- current typed service entrypoints under `client.API.*`
- documented methods and exported request and response structs under `client.API.*` and `client.Web.*`
- documented root-package configuration APIs such as proxy, retry, traffic policy, and request controls
- the public local-only `steamid` package and documented addon import paths

### Request-control infrastructure

The following are documented, stable configuration APIs used by the existing Web surfaces:

- `TrafficClassPublicStorePage`
- public store-page header profiles
- Referer selectors
- short cache
- block detection
- per-class transport hooks

These building blocks do not promise a general-purpose scraper or browser client.

### Experimental

Only capabilities explicitly documented as experimental are outside the stable v1 contract on that basis. An unofficial upstream does not by itself make a documented Go API experimental.

### Out of Scope

The SDK does not promise:

- complete Steam endpoint coverage
- general-purpose Store or Community HTML scraping and automatic browser fallback
- stable upstream schemas or availability for unofficial Web surfaces

## Steam Web API services

`client.API.*` includes documented official APIs and observed Steam services. Their upstream status must be described accurately; hosting on `api.steampowered.com` alone is not an official stability guarantee.

Documented method signatures and exported types remain subject to the v1 compatibility contract. New coverage is additive; security and correctness exceptions must be documented explicitly.

## Raw payload subtrees

Some official responses contain high-volatility subtrees and are intentionally modeled as `json.RawMessage`.

The presence of a raw subtree does not make the surrounding method unstable.  
It means only that the fine-grained internal JSON shape of that subtree is not promised as a typed stable contract.

## Unofficial Web Surfaces

`client.Web.*` provides read-only Storefront, Community, and Market JSON endpoints outside `api.steampowered.com`.

These Go method signatures are stable under the `v1` compatibility policy.

Valve does not guarantee these upstream payloads and their availability as stable official Web API contracts. Handle drift through defensive decoding and additive models while preserving existing signatures, field types, JSON tags, and raw fields.

## Addons

The documented addon import paths are part of the supported repository structure, but addon behavior should still be interpreted based on its own documented scope.

Examples:

- `addons/openid`
- `addons/a2s`
- `addons/a2s/master`
- `addons/a2s/scanner`
- `addons/assets`
- `addons/markup`
- `addons/vdf`
- `addons/websession`
- `addons/freeclaim`

Resource discovery and downloads in `addons/assets` are existing supported capabilities. Authentication and mutation boundaries for other addons are documented in the [addon reference](../addons/reference.md) and [safety guidance](../addons/safety.md).
