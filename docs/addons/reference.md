# Addons

This document keeps the addon-level usage notes that would otherwise make the repository homepage too noisy.

## `addons/openid`

Use `addons/openid` when you want browser-based Steam sign-in.

What it does:

- builds the Steam OpenID login URL
- verifies the callback against Steam with `check_authentication`
- returns `SteamID64`, `ClaimedID`, and the recovered `state`

What it does not do:

- it does not replace Web API credentials
- it does not fetch profile data by itself
- it does not manage sessions for you

The example now shows a more realistic pattern:

- generate a random `state`
- store it in a cookie before redirecting to Steam
- verify the callback with `addons/openid`
- compare the returned `state` with the cookie
- clear the cookie after a successful login

Run it in direct mode:

```bash
go run ./examples/openid
```

Run it with a proxy for the server-side verification request:

```bash
go run ./examples/openid --proxy http://127.0.0.1:7897
```

This is especially useful for China-region networks where the browser login page may work while the server-side `check_authentication` request still needs a proxy.

## `addons/websession`

Use `addons/websession` when you want one manual Steam web-login workflow built on top of `client.API.AuthenticationService`.

What it does:

- starts one credentials-based auth session
- accepts one optional Steam Guard code submission
- polls the auth session until Steam issues tokens
- exchanges one refresh token for Store / Community web cookies
- validates both Store and Community sessions

What it does not do:

- it does not persist passwords, refresh tokens, or cookies for you
- it does not read browser cookies or Steam client local state
- it does not silently print secrets in the example output

The example accepts `-account`, `-password`, `-guard-code`, and `-proxy`, and also supports:

- `STEAM_ACCOUNT_NAME`
- `STEAM_PASSWORD`
- `STEAM_GUARD_CODE`

Run it:

```bash
go run ./examples/websession
```

## `addons/freeclaim`

Use `addons/freeclaim` when you want one addon-level bridge for Store promotion discovery and one explicit free-license claim.

What it does:

- searches current free Store promotions from the Store search HTML fragment response
- resolves one app's free package candidates from `client.Web.Storefront.GetAppDetails`
- checks app ownership through `dynamicstore/userdata`
- claims one free package only when you explicitly request it

What it does not do:

- it does not manage account passwords or browser cookies
- it does not read Steam client local tokens or any local account database
- it does not auto-claim everything
- it does not retry forever

The example is read-only by default. Claim mode requires a refresh token through `-refresh-token` or `STEAM_REFRESH_TOKEN`.

Run the read-only search / package resolution example:

```bash
go run ./examples/freeclaim
```

Run an explicit claim after choosing an app and package:

```bash
go run ./examples/freeclaim -app-id 480 -package-id 12345 -claim
```

## `addons/a2s`

Use `addons/a2s` when you want to query a game server directly without pulling `a2s-go` into your own import tree.

The bridge currently re-exports the upstream `a2s-go` client and key result types, so you can call:

- `QueryInfo`
- `QueryPlayers`
- `QueryRules`

Example:

```bash
go run ./examples/a2s -server 1.2.3.4:27015 -query info
go run ./examples/a2s -server 1.2.3.4:27015 -query players
go run ./examples/a2s -server 1.2.3.4:27015 -query rules
```

## `addons/a2s/master`

Use `addons/a2s/master` when you want discovery against the A2S master server protocol.

The bridge follows the upstream `a2s-go/master` surface and is intended for:

- one-page discovery with `Query`
- streamed discovery with `Stream`

## `addons/a2s/scanner`

Use `addons/a2s/scanner` when you want to batch probe discovered addresses.

The bridge follows the upstream `a2s-go/scanner` package and supports:

- probing address lists
- consuming discovery streams
- batch `info`, `players`, and `rules` queries

When `a2s-go` publishes new stable releases, `steam-go` should keep the bridge version and examples in sync rather than re-implementing A2S logic locally.
