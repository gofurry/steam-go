# Changelog

All notable changes to `steam-go` will be documented in this file.

The format is intentionally simple during the pre-v1 release phase.

## Unreleased

### Planned for v1.0.0-rc.1

- freeze the `v1.0.0` stable surface
- publish compatibility, release, endpoint stability, and endpoint coverage documents
- align README and Chinese docs with the release candidate message

### Planned for v1.0.0-rc.2

- regression validation and release hardening
- example and documentation consistency checks
- bug fixes only, no new API coverage

## v1.0.0-alpha-3

### Added

- per-traffic-class request policy routing
- public store-page request profiles and Referer strategies
- short in-memory caching with conditional revalidation
- block detection for public store-page traffic
- host and session request controls
- sticky proxy selection, proxy health checks, and proxy metrics snapshots
- per-class transport hook foundations

### Improved

- safer retry and rate-limit configuration
- cookie jar based session persistence
- proxy and traffic policy documentation
- OpenID hardening, proxy support, and test coverage
- CI checks including race, vet, staticcheck, and govulncheck

## v1.0.0-alpha-2

### Added

- `WishlistService`
- expanded `PlayerService` coverage
- proxy selection helpers and request-layer hardening

### Improved

- retry behavior
- verifier body size limits
- safer default timeouts
- request safety around large or volatile responses

## v1.0.0-alpha-1

### Added

- initial root `Client` and grouped `client.API.*` structure
- functional option based client configuration
- initial typed Steam Web API support
- addon foundations for `openid` and `a2s`
