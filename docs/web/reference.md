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
- `GetAdjacentPartnerEvents` / `GetAdjacentPartnerEventsRaw`
- `ListAppReviews`
- `GetAppDetailsBatch`
- default traffic class: `TrafficClassPublicStorePage`

`GetAppDetails` includes typed high-value Store fields such as capsule URLs,
screenshots, movies/trailers, background URLs, highlighted achievements,
recommendations, Metacritic, support info, content descriptors, and ratings raw
JSON. Use `GetAppDetailsRaw` when you need fields not yet typed by the SDK.
Use `AppDetailsData.DecodeRatings` for common rating board fields and
`AppDetailsData.SteamGermanyRequiredAge` for Steam Germany age requirements.

`GetAdjacentPartnerEvents` wraps the public Store events JSON endpoint used by
Steam partner news pages. The method exposes a stable typed subset and preserves
raw nested payloads for callers that need fields not yet typed by the SDK.

### `client.Web.Community`

- `GetInventory` / `GetInventoryRaw`
- `ListInventory`
- default traffic class: `TrafficClassCommunityWeb`
- inventory access may require caller-supplied cookies through `WithCookieJar(...)` or `WithDefaultCookieJar()`

### `client.Web.Market`

- `GetPriceOverview` / `GetPriceOverviewRaw`
- `GetPriceOverviewBatch`
- default traffic class: `TrafficClassMarketWeb`

## Request behavior

- `client.Web.*` never injects Steam Web API `key` or `access_token`
- proxy selection, rate limiting, retry, short-cache, block detection, header profiles, referer policies, and cookie jars still flow through the existing client option system
- no built-in login, cookie refresh, browser fallback, purchase, sell, trade, or other account automation
- paginator and batch helpers use the same request controls as the underlying single-item methods
- `WithRequestObserver(...)` emits sanitized request events without raw query strings, headers, bodies, credentials, cookies, or proxy passwords
