# Agent Playbook

Use these recipes to keep changes consistent with the repository.

## Add an Official API Endpoint

1. Add the endpoint path to `internal/endpoint/endpoint.go`.
2. Add typed and raw methods in the matching `api/<service>/` package.
3. Add request/response types in `types.go`.
4. Reuse existing validation and request execution helpers.
5. Add tests for request construction and decoding.
6. Update API reference / coverage docs.
7. Add a concise `CHANGELOG.md` entry.

Do not create a new service package if the endpoint already belongs to an existing Steam interface.

## Add or Extend a Web Surface

1. Put it under the matching `web/<surface>/` package.
2. Keep the operation read-only.
3. Preserve typed outer structures where useful.
4. Keep volatile nested payloads raw when necessary.
5. Add defensive tests using local HTTP fixtures/servers.
6. Document upstream volatility.

Do not silently promote an unofficial surface to an official contract.

## Add Normalization

1. Keep the original Steam field unchanged.
2. Add a separate typed normalization API.
3. Preserve unknown input.
4. Avoid guessed semantics.
5. Prefer deterministic pure functions.
6. Add edge cases and future-format fallback tests.

A parser failure must not fail an otherwise successful upstream API response unless the caller explicitly invoked a strict parser.

## Add an Asset Capability

1. Prefer Steam-returned URLs over path inference.
2. Compose existing `steamuser`, `playerservice`, Storefront, or other services.
3. Convert resources into the shared `addons/assets` result model.
4. Reuse Verify / Read / Download pipelines.
5. Preserve ownership metadata such as AppID or SteamID.
6. Test success, missing resource, failure, and cancellation propagation.

Do not scrape HTML or guess CDN paths if structured Steam data already provides the resource.

## Change Public Types

Before editing exported names, signatures, fields, or documented semantics:

1. Read `contracts/compatibility.md`.
2. Prefer additive changes.
3. Check whether JSON tags are already part of observable behavior.
4. Preserve existing raw fields and import paths.
5. Add compatibility-focused tests when behavior could regress.

If an existing valid caller would stop compiling or change meaning, treat the change as breaking.

## Change Request Infrastructure

Changes to retries, timeouts, proxy routing, redaction, body limits, or traffic policy have a wide blast radius.

1. Keep service code independent from transport details.
2. Preserve credential redaction.
3. Preserve method-aware retry behavior.
4. Add focused regression tests.
5. Run race tests.

## Before Finishing

Check:

```bash
go test ./...
go test -race ./...
go vet ./...
```

Then confirm:

- public API remains compatible
- contracts are respected
- no secrets appear in tests/examples
- relevant docs are updated
- `CHANGELOG.md` records user-visible changes only
