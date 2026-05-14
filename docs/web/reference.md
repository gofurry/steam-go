# Web Reference

`steam-go` exposes a small read-only `client.Web.*` layer for high-value Steam web JSON surfaces outside the official `api.steampowered.com` Web API.

## Stability

- The Go method signatures are part of the stable `v1.x` surface.
- Upstream Store / Community / Market payloads remain unofficial or volatile web surfaces.
- High-volatility nested payloads may stay partially typed and use `json.RawMessage`.

## Services

### `client.Web.Storefront`

- `GetAppDetails` / `GetAppDetailsRaw`
- `GetPackageDetails` / `GetPackageDetailsRaw`
- `GetAppReviews` / `GetAppReviewsRaw`
- default traffic class: `TrafficClassPublicStorePage`

### `client.Web.Community`

- `GetInventory` / `GetInventoryRaw`
- default traffic class: `TrafficClassCommunityWeb`
- inventory access may require caller-supplied cookies through `WithCookieJar(...)` or `WithDefaultCookieJar()`

### `client.Web.Market`

- `GetPriceOverview` / `GetPriceOverviewRaw`
- default traffic class: `TrafficClassMarketWeb`

## Request behavior

- `client.Web.*` never injects Steam Web API `key` or `access_token`
- proxy selection, rate limiting, retry, short-cache, block detection, header profiles, referer policies, and cookie jars still flow through the existing client option system
- no built-in login, cookie refresh, browser fallback, purchase, sell, trade, or other account automation
