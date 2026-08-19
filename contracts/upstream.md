# Steam Upstream Contract

This contract defines how `steam-go` treats different Steam data sources.

## Tier 1 — Official API

Examples:

- documented `api.steampowered.com` Web APIs
- stable Valve-documented API contracts

Policy:

- preferred source when available
- typed request/response models when stable
- suitable for the primary v1 SDK surface
- still decode defensively where Valve payloads vary

## Tier 2 — Observed Steam Services

Examples:

- Steam service endpoints used by Steam web/client surfaces but not listed in the current public Web API reference
- currently observed `IPlayerService`-style profile endpoints

Policy:

- allowed when useful and reasonably stable in practice
- explicitly document upstream status
- do not describe them as Valve-guaranteed public API
- preserve defensive decoding and fallback behavior

A stable `steam-go` method does not make the upstream Steam endpoint officially stable.

## Tier 3 — Web and Resource Surfaces

Examples:

- Storefront JSON
- Community JSON
- Market JSON
- Steam-returned static/media URLs

Policy:

- read-only by default
- upstream payloads may drift
- preserve raw data where structure is volatile
- use bounded, defensive parsing
- prefer Steam-returned resource URLs

## Source Priority

When multiple sources can provide the same fact, prefer:

```text
official structured API
>
observed structured Steam service
>
structured web payload
>
derived interpretation
```

HTML scraping is not a preferred fallback for core SDK behavior.

## No-Invention Rules

Never:

- infer historical state from current page state
- manufacture exact dates from coarse release windows
- convert unknown values into false/zero semantics when absence is meaningful
- guess CDN URLs when Steam returns authoritative URLs
- silently drop unknown languages, states, or identifiers
- claim an observed endpoint is official without evidence

## Normalization

Normalization exists to interpret Steam data, not redefine it.

Required behavior:

- preserve original raw input
- return typed unknown/fallback states for new formats
- avoid fatal errors for non-critical parser drift
- keep application-specific business rules outside `steam-go`

## Credentials and Mutation

Unofficial web surfaces are read-only unless an addon explicitly documents a narrow mutation workflow.

Never automatically inject credentials into read-only web helpers.

Secrets must not appear in:

- logs
- error messages
- fixtures
- examples
- generated documentation

## Documentation Language

Use precise labels such as:

- `official`
- `documented`
- `observed`
- `undocumented`
- `volatile`

Do not use “official” as a synonym for “hosted by Steam”.
