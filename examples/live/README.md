# Live Smoke Tests

`examples/live/` contains real API smoke programs that require real credentials and external network access.

These commands are not normal offline examples. Use them only when you want to validate real Steam Web API behavior against the current public surface.

## Credentials

Place credential files in `examples/live/`:

- `key.txt`
- `access-token.txt`
- `family-group-id.txt` (optional, only for `familygroupsservice`)
- `proxy.txt` (optional, used when `STEAM_PROXY` is unset)

For local transition, the shared helper also falls back to legacy files under `test/` when the new location is empty.

## Run

Run one service-specific smoke command, for example:

- `go run ./examples/live/accountcartservice`
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

## Notes

- These programs are intended for manual validation, not CI.
- Keep real credentials out of Git.
- Use the regular `examples/` directory for lightweight usage demos that should not depend on live credentials.
