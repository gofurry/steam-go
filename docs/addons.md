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
