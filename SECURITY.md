# Security Policy

`steam-go` handles Steam Web API keys, access tokens, refresh tokens, web cookies, and proxy URLs in several workflows. Treat all of these values as sensitive.

## Reporting a Vulnerability

Please do not report security issues in public GitHub issues.

If you believe you found a vulnerability, contact the maintainers privately through the security contact configured on the GitHub repository. Include:

- a short description of the issue
- affected package or addon
- reproduction steps or a minimal proof of concept
- whether credentials, cookies, logs, redirects, or proxy URLs are involved
- the expected impact

Avoid sending real Steam credentials. Use fake values or redacted samples whenever possible.

## Sensitive Values

Never paste these values into public issues, pull requests, logs, traces, screenshots, or examples:

- Steam Web API keys, including publisher or partner keys
- `access_token`
- `refresh_token`
- `steamLoginSecure`
- `sessionid`
- `webapi_token`
- `loyalty_webapi_token`
- raw `Cookie` headers
- proxy URLs containing usernames or passwords
- final redirect URLs that include credentials or cookies

Steam credentials often appear in URL query strings. Before logging URLs, use `steam.RedactSensitiveURL(...)` or an equivalent redaction step in your own application.

## Supported Scope

The stable `v1` security surface is centered on:

- the root `steam` package
- `client.API.*` official Steam Web API services
- documented `client.Web.*` read-only web JSON services
- documented addon import paths
- credential injection, redaction, proxy, cookie jar, retry, rate limit, and traffic policy helpers

Experimental or undocumented scraping behavior is outside the supported security scope unless it is explicitly documented.

## Safe Usage Expectations

- Keep API keys and tokens on trusted backend systems.
- Do not expose Steam credentials to frontend JavaScript, mobile clients, game clients, or public repositories.
- Prefer environment variables or a secret manager for local and production configuration.
- Do not pass refresh tokens or passwords through command-line arguments.
- Keep live smoke credentials out of Git. The repository ignores common local credential files, but callers are still responsible for secret handling.
- Use addon flows only for their documented purpose. Addons must stay explicit, opt-in, and avoid hidden account automation.

## Dependency and Tooling Security

Release validation should include:

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- `staticcheck ./...`
- `govulncheck ./...`

Mainline CI pins static analysis tools for repeatability. A scheduled advisory workflow may run the latest tools to detect upcoming ecosystem changes without changing normal pull request gates.
