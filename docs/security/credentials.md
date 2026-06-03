# Credential Safety

Steam credentials are often long-lived, account-bound, or session-bound. Treat them as secrets even when they appear in URLs, redirects, cookies, examples, or debug output.

## Credential Types

| Value | Typical use | Safety notes |
|---|---|---|
| Steam Web API key | Official Web API methods that require `key` | Backend-only. Reset if leaked. |
| Publisher / partner key | Steamworks partner backend APIs | Higher privilege than normal keys. Never distribute to clients. |
| `access_token` | Token-backed user or web flows | Shorter-lived but still account-sensitive. |
| `refresh_token` | Exchanging for new web/session tokens | High-risk. Do not pass through command-line flags. |
| `steamLoginSecure` / `sessionid` | Store / Community web cookies | Session-bound. Keep inside a controlled cookie jar. |
| Proxy URL userinfo | Proxy authentication | Redact before logging. |

OpenID is different: it verifies a Steam identity and returns a SteamID64. It does not replace API keys, access tokens, or application sessions.

## Why URLs Are Risky

Steam APIs commonly carry credentials in query parameters:

```text
https://api.steampowered.com/.../?key=SECRET&access_token=TOKEN
```

These URLs can leak through:

- application logs
- HTTP traces
- metrics labels
- redirect final URLs
- panic output
- screenshots
- bug reports

Use `steam.RedactSensitiveURL(...)` before logging URLs.

## Header and Cookie Redaction

Use `steam.RedactSensitiveHeaders(...)` before logging request or response headers:

```go
safeHeaders := steam.RedactSensitiveHeaders(req.Header)
```

Sensitive headers such as `Authorization`, `Proxy-Authorization`, `Cookie`, `Set-Cookie`, `X-WebAPI-Key`, and `X-API-Key` are replaced with `[REDACTED]`.

`RedactSensitiveHeaders(...)` returns a clone and does not mutate the original header map.

## Cookie Jar Lifecycle

- Keep cookie jars scoped to one user session or one explicit workflow.
- Do not reuse authenticated web cookies across unrelated users.
- Do not serialize cookie jars into logs or public artifacts.
- Clear or replace jars when a user logs out or rotates credentials.
- Inventory and web-session flows may need caller-supplied cookies, but the core SDK does not manage long-term login state for you.

## Examples and CLI Safety

Project examples use environment variables or hidden terminal prompts for sensitive values. Avoid command-line flags for passwords, refresh tokens, and cookies because shell history and process listings can expose them.

Prefer:

- environment variables for local demos
- hidden prompts for interactive examples
- CI/CD secrets for automation
- dedicated secret managers in production

## If a Secret Leaks

1. Revoke or rotate the leaked credential.
2. Remove the value from logs, issues, traces, and artifacts where possible.
3. Check CI variables and deployment configuration.
4. Audit recent use of the credential.
5. Add or improve redaction tests for the leak path.
