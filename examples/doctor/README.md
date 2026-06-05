# steam-go Doctor

`examples/doctor` runs live diagnostics against Steam surfaces and local client configuration.

```bash
go run ./examples/doctor
go run ./examples/doctor -json
```

## JSON Output

`-json` prints a report with two top-level fields:

```json
{
  "summary": {
    "ok": 3,
    "warn": 1,
    "fail": 0
  },
  "checks": [
    {
      "category": "official_api",
      "name": "steamwebapiutil.get_server_info",
      "status": "OK",
      "message": "reachable",
      "detail": "optional redacted detail"
    }
  ]
}
```

`summary.ok`, `summary.warn`, and `summary.fail` count final check statuses. Each check contains:

- `category`: `environment`, `credential`, `official_api`, `web`, or `proxy`.
- `name`: stable check name inside this example.
- `status`: `OK`, `WARN`, or `FAIL`.
- `message`: short result text.
- `detail`: optional redacted context.

This JSON is useful for scripts and release diagnostics, but `examples/doctor` remains an example rather than a stable CLI product.

## Exit Codes

- `0`: all checks are `OK` or `WARN`.
- `1`: at least one check reports `FAIL`.
- `2`: flag parsing, configuration, or output rendering failed.

`WARN` means a check was skipped or needs attention, but does not make the command fail.

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

Doctor output must not include API keys, access tokens, cookies, raw headers, raw query strings, or proxy passwords.

## Checks

- Go runtime and request policy summary.
- Credential presence.
- Official API reachability through `SteamWebAPIUtil.GetServerInfo`.
- Optional key-backed `SteamUser.GetPlayerSummaries`.
- Storefront app details and reviews.
- Market price overview.
- Optional Community inventory when `STEAM_PUBLIC_INVENTORY_ID` is set.

Doctor is a diagnostic example, not a stable CLI API.
