# Live Smoke Tests

`examples/live/` contains real API smoke programs that require real credentials and external network access.

These commands are not normal offline examples. Use them only when you want to validate real Steam Web API behavior against the current public surface.

## Credentials

Place credential files in `examples/live/`:

- `key.txt`
- `access-token.txt`
- `family-group-id.txt` (optional, only for `familygroupsservice`)
- `proxy.txt` (optional, used when `STEAM_PROXY` is unset)

Optional environment variables for web smoke tests:

- `STEAM_AUTH_ACCOUNT_NAME` for `examples/live/authenticationservice`
- `STEAM_PUBLIC_INVENTORY_ID` for `examples/live/webcommunity`

For local transition, the shared helper also falls back to legacy files under `test/` when the new location is empty.

## Run

Run one service-specific smoke command, for example:

- `go run ./examples/live/accountcartservice`
- `go run ./examples/live/authenticationservice`
- `go run ./examples/live/billingservice`
- `go run ./examples/live/communityservice`
- `go run ./examples/live/familygroupsservice`
- `go run ./examples/live/loyaltyrewardsservice`
- `go run ./examples/live/mobilenotificationservice`
- `go run ./examples/live/newsservice`
- `go run ./examples/live/playerservice`
- `go run ./examples/live/questservice`
- `go run ./examples/live/salefeatureservice`
- `go run ./examples/live/steamapps`
- `go run ./examples/live/steamchartsservice`
- `go run ./examples/live/steamdirectory`
- `go run ./examples/live/steamnews`
- `go run ./examples/live/steamnotificationservice`
- `go run ./examples/live/steamuser`
- `go run ./examples/live/steamuseroauth`
- `go run ./examples/live/steamuserstats`
- `go run ./examples/live/steamwebapiutil`
- `go run ./examples/live/storebrowseservice`
- `go run ./examples/live/storecatalogservice`
- `go run ./examples/live/storepreferencesservice`
- `go run ./examples/live/storeservice`
- `go run ./examples/live/storetopsellersservice`
- `go run ./examples/live/useraccountservice`
- `go run ./examples/live/userreviewsservice`
- `go run ./examples/live/userstorevisitservice`
- `go run ./examples/live/wishlistservice`
- `go run ./examples/live/webstorefront`
- `go run ./examples/live/webmarket`
- `go run ./examples/live/webcommunity`

## Opt-in Smoke Report

The package-level smoke test stays offline by default:

```bash
go test ./examples/live/...
```

Set `STEAM_GO_LIVE=1` to run the low-risk live baseline for `steamwebapiutil` and `webstorefront`:

```bash
STEAM_GO_LIVE=1 go test ./examples/live/...
```

For release diagnostics, you can archive JSON and human-readable reports:

```bash
STEAM_GO_LIVE=1 \
STEAM_GO_LIVE_REPORT=tmp/live-smoke.json \
STEAM_GO_LIVE_REPORT_HUMAN=tmp/live-smoke.txt \
go test ./examples/live/...
```

If `STEAM_GO_LIVE` is not `1`, the test still skips without network access. When a report path is set in that skipped mode, the report contains `SKIP` and the skipped reason.

Report fields are intentionally small: check name, status, message, duration, redacted proxy mode, and skipped reason. Reports must not include API keys, access tokens, cookies, raw URLs, raw headers, raw bodies, or proxy passwords.

## Notes

- These programs are intended for manual validation, not CI.
- Keep real credentials out of Git.
- Use the regular `examples/` directory for lightweight usage demos that should not depend on live credentials.
