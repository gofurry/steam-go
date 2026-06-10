# Addon Safety Boundaries

Addons keep optional workflows outside the core SDK. They should stay explicit, narrow, and easy to audit.

## `addons/openid`

- Verifies browser-based Steam OpenID callbacks.
- Does not create an application user system.
- Does not fetch player profile data.
- Does not manage long-lived sessions or cookies.
- Callers must validate and bind `state` to their own CSRF/session model.

## `addons/websession`

- Drives one manual Steam web-login workflow.
- Handles cookie jars returned by Steam, so callers must protect those cookies like credentials.
- Does not persist passwords, refresh tokens, access tokens, or cookies.
- Does not read browser cookies or local Steam client state.
- Should be used deliberately by applications that already have a credential handling policy.

## `addons/freeclaim`

- Provides read-only promotion/package discovery and one explicit free-license claim.
- Does not auto-claim every free package.
- Does not manage account pools or bulk account automation.
- Does not retry forever or bypass upstream limits.
- Claim mode should require an explicit caller decision and protected refresh-token handling.

## `addons/markup`

- Sanitizes generated or caller-provided HTML by default.
- `WithSanitize(false)` is only appropriate for trusted input and trusted render targets.
- Does not fetch remote content.
- Does not decide which sanitized tags your application should render.

## `addons/vdf`

- Parses only caller-provided text VDF / KeyValues input.
- Does not parse binary VDF or `shortcuts.vdf`.
- Does not scan local Steam installations automatically.
- Does not read user directories unless the caller explicitly passes a path.
- Does not extract accounts, tokens, cookies, or sessions.

## `addons/assets`

- Builds and optionally fetches public Steam asset URLs.
- Constructed static URLs do not require network access; verify/read/download helpers do.
- Direct URL helpers must use `URLValidator` when URLs come from users or another untrusted source.
- Asset existence is not guaranteed because Steam can add, move, or remove resources.

## `addons/a2s`

- Queries game servers and master-server discovery endpoints.
- Does not proxy, hide, or rotate caller identity.
- Batch scanning should use explicit caller-side limits and timeouts.
- The bridge should follow upstream `a2s-go` behavior instead of reimplementing protocol logic locally.
