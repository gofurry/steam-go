# Architecture

This document gives agents the minimum structural model needed to modify `steam-go` safely.

## Public Shape

```text
steam package
│
├─ Client
│  ├─ API.*
│  │   └─ api/*
│  └─ Web.*
│      └─ web/*
│
└─ addons/*
    └─ higher-level optional capabilities
```

Shared request machinery lives behind public packages in `internal/`.

## Root Package

The root package owns cross-cutting SDK behavior:

- `Client`
- functional options
- retry, timeout and rate limiting
- proxy and traffic policies
- credential handling and redaction
- shared error surface

Do not place Steam service-specific response models here.

## `api/`

`api/` contains Steam Web API service bindings.

Typical package shape:

```text
api/<service>/
  service.go
  methods.go
  types.go
  *_test.go
```

Rules:

- endpoint constants belong in `internal/endpoint/endpoint.go`
- stable official payloads should prefer typed structs
- provide `Raw` methods where the package pattern already does
- use `json.RawMessage` for large or volatile subtrees
- validate identifiers before building requests
- service-specific logic stays inside the matching package

## `web/`

`web/` contains read-only Steam web surfaces outside the main official Web API contract.

Current intent includes Storefront, Community, and Market-style JSON endpoints.

Important distinction:

```text
stable Go method
!=
stable upstream payload
```

Prefer typed outer structures and defensive handling for volatile nested data.

## `addons/`

Addons provide optional capabilities that should not bloat the core SDK.

Good addon candidates:

- combine multiple existing SDK services
- discover or transform resource URLs
- verify, read, or download Steam resources
- provide optional workflows with a clear boundary

Prefer reusing root request infrastructure and existing addon pipelines instead of duplicating transport behavior.

`addons/assets` is the shared resource pipeline for URL discovery, verification, reading, and downloading.

## `internal/`

`internal/` is implementation-only.

Common responsibilities include:

- endpoint constants
- request execution
- response decoding
- Steam ID helpers
- static host metadata
- internal error construction

Do not expose `internal/*` packages as user-facing dependencies.

## Data Modeling

Use the least brittle representation that preserves meaning:

```text
stable upstream field
→ typed field

volatile nested subtree
→ typed outer + json.RawMessage

weak display text
→ preserve raw + optional normalization API
```

Normalization must not replace the raw upstream contract.

## Dependency Direction

Prefer:

```text
root / api / web
      ↓
   internal

addons
  ↓
public SDK packages + addon internals
```

Avoid dependencies from core API packages into addons.

## Change Placement Examples

| Goal | Place |
|---|---|
| Add `ISteamUser/...` | `api/steamuser/` |
| Parse Storefront release text | `web/storefront/` |
| Download profile images | `addons/assets/` |
| Add retry behavior | root / `internal/request` |
| Add Steam-specific endpoint constant | `internal/endpoint` |

When placement is unclear, choose the narrowest package that owns the upstream concept.
