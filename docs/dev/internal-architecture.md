# Internal Architecture

`steam-go` keeps `github.com/gofurry/steam-go` as the stable public SDK facade.
The root package owns exported user-facing types, constructors, options, and
compatibility-sensitive sentinel errors.

Implementation-heavy logic should live under `internal/` when it does not need
to be imported by users directly.

## Root Package Boundary

The root package should continue to expose stable integration points:

- `Client`, `API`, `Web`, and `NewClient`
- functional options such as `WithAPIKey`, `WithTimeout`, and `WithTrafficPolicy`
- public traffic, proxy, request profile, raw HTTP, observer, redaction, error, and runtime stats types
- thin adapters that translate between public root types and internal implementation types

Do not move exported root types to internal packages. If implementation data has
to cross the boundary, keep the public type in root and map to or from an
internal type explicitly.

## Internal Package Rules

Internal packages must not import the root `steam` package. This keeps import
cycles out of the SDK and makes implementation packages easier to test.

Preferred patterns:

- define small structural interfaces inside internal packages
- let root public interfaces satisfy internal interfaces naturally
- use root adapters for public metrics, public errors, and public config structs
- keep implementation state machines, context keys, normalization helpers, and policy matching logic internal

Avoid:

- aliasing public root exported types to internal exported types
- moving public constructors to new import paths
- changing root method signatures during internal refactors
- introducing external dependencies for small implementation helpers

## Current Internal Extractions

`internal/rawhttp` owns raw HTTP host policy implementation and host/suffix
normalization. Root constructors return values that satisfy the public
`RawHTTPHostPolicy` interface.

`internal/requestprofile` owns browser-like header application, Referer selector
implementations, Referer context storage, and Referer URL normalization. The
root package keeps `HeaderProfile`, `RefererSelector`, and `RefererRoute`.

`internal/proxyselector` owns proxy selector parsing and state machines,
including round-robin selection, health cooldown, sticky session caching, route
matching, and internal metrics snapshots. The root package keeps public proxy
types and maps internal metrics and errors back to root-owned API values.

## Compatibility Checklist

Before merging an internal refactor:

- run `go test ./...`
- run `go test -race ./...` when practical
- run `go vet ./...`
- check public API compatibility
- confirm root import path and addon import paths are unchanged
- confirm exported root names and method signatures are unchanged
- confirm sentinel errors such as `ErrAllProxiesCoolingDown` keep `errors.Is` behavior
- confirm sensitive values remain redacted in observer output, runtime stats, and proxy metrics
