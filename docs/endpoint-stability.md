# Endpoint Stability

This document explains how to interpret API stability in `steam-go` as the project prepares for `v1.0.0`.

## Stability levels

### Stable

Stable means the exported API shape is intended to be covered by the `v1` compatibility promise.

This includes:

- current typed service entrypoints under `client.API.*`
- exported request and response structs for official Steam Web API methods
- documented root-package configuration APIs such as proxy, retry, traffic policy, and request controls

### Preview

Preview means the public configuration surface exists and is documented, but the project is not yet promising a complete productized fetch API behind it.

This currently includes:

- `TrafficClassPublicStorePage`
- public store-page header profiles
- Referer selectors
- short cache
- block detection
- per-class transport hooks

These are valid building blocks, but they should not be read as a promise that `steam-go v1.0.0` already ships a complete public store-page client.

### Experimental

Experimental means a future direction that may appear later as an addon or explicitly experimental package, and is not part of the stable `v1.0.0` contract.

Examples:

- Steam Store page scraping entrypoints
- Steam Community page helpers
- CDN-derived resource helpers
- browser-backed fallback implementations

### Out of Scope

Out of scope means not currently promised before `v1.0.0`.

This includes:

- expanding new API coverage before `v1.0.0`
- shipping non-official Store / Community / CDN fetch APIs as part of the first stable surface

## Official Web API services

The official `api.steampowered.com` services already exposed through `client.API.*` are treated as the main `v1.0.0` stable surface target.

Their method signatures and exported typed response models should remain stable unless a blocker-level issue forces a breaking rethink before `v1.0.0`.

## Raw payload subtrees

Some official responses contain high-volatility subtrees and are intentionally modeled as `json.RawMessage`.

The presence of a raw subtree does not make the surrounding method unstable.  
It means only that the fine-grained internal JSON shape of that subtree is not promised as a typed stable contract.

## Addons

The documented addon import paths are part of the supported repository structure, but addon behavior should still be interpreted based on its own documented scope.

Examples:

- `addons/openid`
- `addons/a2s`
- `addons/a2s/master`
- `addons/a2s/scanner`
