# Contributing

Thanks for helping improve `steam-go`.

This project is a stable Go SDK for the official Steam Web API, with practical request controls and carefully scoped read-only Steam Web helpers. Contributions should preserve that boundary.

## Development Setup

Use Go 1.25 or newer.

Before opening a pull request, run:

```bash
go test ./...
go test -race ./...
go vet ./...
staticcheck ./...
govulncheck ./...
```

If `staticcheck` or `govulncheck` is not installed, use the versions pinned in `.github/workflows/ci.yml`.

## Contribution Rules

- Keep public API changes compatible with the `v1` policy in `docs/governance/compatibility.md`.
- Prefer small, reviewable changes.
- Add tests for new behavior and regression fixes.
- Update documentation when behavior, examples, endpoint coverage, addon scope, or security guidance changes.
- Do not commit real Steam credentials, cookies, proxy passwords, local token files, screenshots containing secrets, or generated local state.
- Do not add hidden browser automation, scraping expansion, purchase, trade, sell, or bulk account automation to the core package.

## Adding or Updating Official API Endpoints

For new `client.API.*` coverage:

- place code in the matching `api/<service>/` package
- keep the service grouped under `client.API.*`
- provide typed request and response structs when the payload is stable enough
- use `json.RawMessage` for large or volatile subtrees
- provide matching raw methods when that is the package pattern
- add unit tests or documented live smoke coverage
- update `docs/api/reference.md`
- update `docs/governance/endpoint-coverage.md` when coverage changes

Do not claim complete Steam coverage unless the documentation and tests prove it.

## Adding or Updating Web Surfaces

`client.Web.*` is intentionally small and read-only.

For Storefront, Community, or Market web helpers:

- document that the upstream surface is unofficial or volatile
- avoid credential injection through `key` or `access_token`
- preserve typed outer structures where possible
- keep high-volatility nested payloads raw
- add tests with local HTTP servers
- avoid browser-backed fallback behavior in the core SDK

## Adding or Updating Addons

Addons may cover capabilities that should not live in the core SDK.

Each addon should document:

- what it does
- what it does not do
- whether it performs network requests
- whether it needs credentials, cookies, tokens, or proxy configuration
- how examples avoid leaking secrets
- whether the behavior is read-only or mutating

Examples that need sensitive values should read them from environment variables or hidden terminal prompts, not command-line flags.

## Error and Credential Safety

- Do not log raw request URLs that may contain credentials.
- Use `steam.RedactSensitiveURL(...)` for URL logging examples.
- Keep error body previews bounded.
- Do not print raw `Cookie` headers, refresh tokens, access tokens, API keys, proxy userinfo, or final redirect URLs containing credentials.

## Pull Request Checklist

Before requesting review:

- tests pass locally or the PR explains why they were not run
- new public API has documentation or examples
- endpoint coverage docs are updated when relevant
- Web surface volatility is documented when relevant
- addon boundaries are documented when relevant
- no credentials or local-only files are included
- release notes or checklist updates are included when the change affects releases
