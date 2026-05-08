# Real API Smoke Tests

Credential files stay in this directory:

- `key.txt`
- `access-token.txt`
- `family-group-id.txt` (optional, only for `familygroupsservice`)
- `proxy.txt` (optional, used when `STEAM_PROXY` is unset)

Run a specific API-group smoke test with:

- `go run ./test/accountcartservice`
- `go run ./test/billingservice`
- `go run ./test/communityservice`
- `go run ./test/familygroupsservice`
- `go run ./test/loyaltyrewardsservice`
- `go run ./test/mobilenotificationservice`
- `go run ./test/newsservice`
- `go run ./test/playerservice`
- `go run ./test/steamnews`
- `go run ./test/steamuser`
- `go run ./test/steamuserstats`
