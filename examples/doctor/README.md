# steam-go Doctor

`examples/doctor` runs live diagnostics against Steam surfaces and local client configuration.

```bash
go run ./examples/doctor
go run ./examples/doctor -json
```

The command exits with code `1` when any check reports `FAIL`. `WARN` means a check was skipped or needs attention, but does not make the command fail.

## Credentials

Environment variables take priority:

- `STEAM_API_KEY`
- `STEAM_ACCESS_TOKEN`
- `STEAM_PROXY`
- `STEAM_PUBLIC_INVENTORY_ID`

When environment variables are empty, doctor falls back to files under `examples/live/`:

- `key.txt`
- `access-token.txt`
- `proxy.txt`

Secrets are reported only as present or missing. Proxy URLs are redacted before output.

## Checks

- Go runtime and request policy summary.
- Credential presence.
- Official API reachability through `SteamWebAPIUtil.GetServerInfo`.
- Optional key-backed `SteamUser.GetPlayerSummaries`.
- Storefront app details and reviews.
- Market price overview.
- Optional Community inventory when `STEAM_PUBLIC_INVENTORY_ID` is set.

Doctor is a diagnostic example, not a stable CLI API.
