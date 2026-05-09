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
- `go run ./test/questservice`
- `go run ./test/salefeatureservice`
- `go run ./test/storebrowseservice`
- `go run ./test/storecatalogservice`
- `go run ./test/storepreferencesservice`
- `go run ./test/storeservice`
- `go run ./test/storetopsellersservice`
- `go run ./test/steamdirectory`
- `go run ./test/steamapps`
- `go run ./test/steamchartsservice`
- `go run ./test/steamnews`
- `go run ./test/steamnotificationservice`
- `go run ./test/steamuser`
- `go run ./test/steamuseroauth`
- `go run ./test/steamuserstats`
- `go run ./test/steamwebapiutil`
- `go run ./test/useraccountservice`
- `go run ./test/userreviewsservice`
- `go run ./test/userstorevisitservice`
- `go run ./test/wishlistservice`
