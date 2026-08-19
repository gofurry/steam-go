# Changelog

All notable changes to this project are documented in this file.

Only tagged releases are listed as versions below. Development work that has not been released stays under `Unreleased`.

## Unreleased

## 1.3.9 - 2026-08-19

### Added

- Add Storefront release-date normalization for day, month, quarter, year, TBA, none, and unknown precision while preserving the raw `release_date` fields.
- Add the Steam language registry and `ParseSupportedLanguages`, including canonical codes, aliases, full-audio markers, and unknown-language preservation.
- Add typed and raw `ISteamUser/ResolveVanityURL/v1` support.
- Add player avatar and equipped Profile asset discovery with shared verify, read, and download helpers.
- Add optional `SteamID` metadata to shared asset URL and result types.
- Add Storefront normalization and live player-assets examples.
- Add `AGENTS.md`, `.agents/*`, and `contracts/*` as a compact maintenance layer for agent architecture, workflows, compatibility, and Steam upstream rules.

### Changed

- Prefer environment variables over local credential files in live smoke helpers when both are available.
- Refresh API coverage documentation and update the Go 1.26 CI lane to Go 1.26.6.
- Simplify bug-report, feature-request, and pull-request templates and remove the dedicated question issue template.

## 1.3.8 - 2026-07-08

### Added

- Add `IGameServersService/GetServerList/v1` with typed server data, filter builders, and a live example.
- Add known Steam static CDN prefix helpers for shared Steam, Akamai, Cloudflare, Fastly, and Steam China hosts.

### Changed

- Extend Steam static URL validation and raw HTTP host policy support to known shared CDN hosts.

## 1.3.7 - 2026-06-16

### Changed

- Move proxy selector, raw HTTP host-policy, and request-profile implementations behind `internal/*` packages while preserving the public v1 API.
- Reduce the root package to thinner public wrappers and document the resulting internal architecture.

## 1.3.6 - 2026-06-15

### Added

- Add typed and raw `IStoreBrowseService/GetItems/v1` support.
- Add official Store item asset discovery from Steam-returned metadata, including hashed asset paths and higher-resolution asset kinds.
- Add `Digest`, `Filename`, and `Source` metadata to asset results.
- Add verify, read, and download helpers for StoreBrowse-backed Store item assets.

### Changed

- Keep legacy AppID-only static URL builders unchanged while separating them from official StoreBrowse-backed asset discovery.

## 1.3.5 - 2026-06-11

### Added

- Add `Client.RuntimeStats()` with sanitized cache, request-control, and proxy runtime counters.
- Add traffic-cache tuning for TTL, maximum entries, and optional GET cache-miss singleflight.
- Add `IContentServerDirectoryService` coverage.
- Expand `IAuthenticationService` coverage for auth-session metadata, risk information, mobile confirmation, and related low-level session operations.
- Add web-session credential persistence helpers.
- Add benchmark baselines for cache, transport, request controls, proxy selection, and raw HTTP hot paths.

### Changed

- Improve cache boundedness, request-control cleanup, runtime instrumentation, and transport hot paths without changing the existing public request model.

## 1.3.4 - 2026-06-10

### Added

- Add raw HTTP host policies for exact hosts, suffix matches, Steam hosts, and Steam static/CDN hosts.
- Add `RedactSensitiveText(...)` and conditional-cache observability.
- Add production traffic-policy guidance, addon safety documentation, and a request-observer example.

### Changed

- Make retry behavior method-aware so non-idempotent requests are not retried by default.
- Unify SteamID64 validation across selected API, Web, and addon surfaces.
- Improve asset batch cancellation and raw HTTP retry-body handling.

### Security

- Harden raw HTTP destination controls and malformed URL/text redaction.

## 1.3.3 - 2026-06-08

### Added

- Add `addons/vdf` as a thin bridge to `github.com/gofurry/vdf-go` for Valve Data Format / KeyValues text handling.
- Add bounded Storefront review collection through `CollectAppReviews`.
- Add local-only Community inventory joining through `JoinInventoryDescriptions`.

## 1.3.2 - 2026-06-07

### Added

- Add the Storefront adjacent partner-events endpoint with typed outer models and raw preservation for volatile nested payloads.
- Add `addons/markup` for Steam BBCode conversion, HTML sanitization, plain-text extraction, and summaries.

### Changed

- Normalize golden-test line endings so fixtures remain stable across Windows and Unix checkouts.

## 1.3.1 - 2026-06-05

### Added

- Add API coverage triage documentation, additional Web fixtures, live smoke reporting, and request-observer benchmarks.
- Add batch/pagination and observability maintenance guidance.

### Changed

- Improve Steam API coverage-drift workflow triage and doctor/live-smoke diagnostics without expanding the public SDK surface.

## 1.3.0 - 2026-06-04

### Added

- Add `internal/tools/steamapi-sync` and scheduled Steam Web API coverage-drift reporting.
- Add fixture and golden regression coverage for official APIs, volatile Web payloads, assets, and redaction.
- Add opt-in live smoke validation and the `examples/doctor` diagnostic tool.
- Add bounded pagination and batch helpers for high-value read-only Web operations.
- Add `WithRequestObserver(...)` and sanitized `RequestEvent` observability.

## 1.2.2 - 2026-06-03

### Added

- Add `SECURITY.md`, `CONTRIBUTING.md`, GitHub issue/PR templates, release checklist, and cookbook documentation.
- Add an internal public-API diff checker for release compatibility review.
- Add header and cookie redaction helpers and broader credential-safety tests.
- Add a scheduled latest-toolchain advisory workflow.

### Changed

- Expand repository hygiene, release validation, and adoption documentation.
- Pin mainline analyzer/toolchain versions for repeatable CI.

### Security

- Strengthen credential-redaction guidance and coverage for headers, cookies, URLs, tokens, and proxy credentials.

## 1.2.1 - 2026-05-21

### Added

- Add `addons/assets` for Store and Library static asset URL construction.
- Add Storefront media discovery for screenshots, movies, and backgrounds.
- Add shared asset verification, reading, downloading, manifest, filename, overwrite, and URL-validation helpers.
- Add the runnable assets example and English/Chinese static-asset documentation.

### Security

- Add caller-configurable URL validation for direct asset verify/read/download operations.

## 1.2.0 - 2026-05-17

### Added

- Add low-level `IAuthenticationService` support for RSA password-key lookup and Steam auth-session flows.
- Add `addons/websession` for explicit Steam web-session handling.
- Add `addons/freeclaim` for read-only free-promotion discovery and explicit single-package claiming.
- Add raw HTTP execution through the SDK runtime for carefully scoped addon-style flows.

### Changed

- Route web-session and free-claim HTTP behavior through the shared SDK request/runtime controls.

## 1.1.0 - 2026-05-14

### Added

- Add read-only `client.Web.Storefront`, `client.Web.Community`, and `client.Web.Market` JSON surfaces.
- Add Web-surface traffic/base-URL configuration, block/cache handling, and live examples.

### Changed

- Extend compatibility and traffic-policy documentation to distinguish official `client.API.*` methods from volatile read-only Web surfaces.

## 1.0.1 - 2026-05-13

### Changed

- Normalize the module, imports, documentation, and related repositories to the lowercase `github.com/gofurry/...` namespace.
- Update `github.com/gofurry/a2s-go` to v1.0.2.

## 1.0.0 - 2026-05-13

### Added

- Publish the first stable `steam-go` release.
- Add the stable root `Client` with grouped official Steam Web API access through `client.API.*`.
- Add functional options for API keys, access tokens, timeout, retry, rate limiting, proxy selection, and traffic policies.
- Add production-oriented request controls including safe defaults, bounded response bodies, proxy health/metrics, and traffic-class isolation.
- Add stable addon entrypoints for Steam OpenID and A2S helpers.
- Add compatibility, endpoint-stability, endpoint-coverage, and release-governance documentation.
