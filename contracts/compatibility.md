# Public Compatibility Contract

This is the normative compatibility contract for `steam-go` v1.x.

## v1 Rule

v1.x evolves additively by default.

Stable public surface includes documented:

- root `steam` package APIs
- `Client` and functional options
- `client.API.*`
- `client.Web.*`
- exported methods and public types
- documented addon import paths
- documented error kinds and behavior

## Safe Additive Changes

Normally allowed:

- new packages or methods
- new response types
- new constants
- new optional struct fields
- additional defensive decoding
- new helpers that preserve existing behavior
- internal refactors behind unchanged public APIs

## Breaking Changes

Do not make these in v1.x without an explicit compatibility decision:

- remove or rename exported symbols
- change existing method signatures
- move documented public import paths
- change the established semantic meaning of options
- replace a raw upstream field with normalized/derived data
- change valid existing input into invalid input without a strong correctness or security reason
- reuse an existing enum/string value for a different meaning

## Struct and JSON Rules

Exported response/request structs are observable API.

- preserve existing field names and JSON tags
- add fields rather than repurposing fields
- avoid converting an existing field to a materially incompatible Go type
- use pointer/optional fields when absence has distinct meaning

## Raw Payload Rule

Use:

```text
stable payload
→ typed structs

volatile subtree
→ typed outer + json.RawMessage
```

The internal schema of `json.RawMessage` is not guaranteed unless separately documented.

## Normalization Rule

When interpreting weak Steam fields:

```text
raw upstream value
+
normalized view
```

Never:

```text
raw upstream value
→ replaced by guessed normalized value
```

Unknown formats must remain observable.

## Upstream Drift

A stable Go API may wrap an unstable Steam surface.

When upstream drift occurs:

1. preserve the public Go signature when practical
2. make decoding more defensive
3. move volatile internals to raw data if required
4. avoid inventing fallback values

## Exceptions

Security and correctness fixes may tighten behavior when preserving the old behavior would be unsafe or objectively invalid.

Such changes must be documented clearly.
