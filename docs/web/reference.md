# Web Reference

`steam-go` exposes a small read-only `client.Web.*` layer for high-value Steam web JSON surfaces outside the official `api.steampowered.com` Web API.

## Stability

- The Go method signatures are part of the stable `v1.x` surface.
- Upstream Store / Community / Market payloads remain unofficial or volatile web surfaces.
- High-volatility nested payloads may stay partially typed and use `json.RawMessage`.

## Services

### `client.Web.Storefront`

- `GetAppDetails` / `GetAppDetailsRaw`
- `GetResolvedAppDetails` / local `storefront.ResolveAppDetails`
- `GetPackageDetails` / `GetPackageDetailsRaw`
- `GetAppReviews` / `GetAppReviewsRaw`
- `GetAdjacentPartnerEvents` / `GetAdjacentPartnerEventsRaw`
- `ListAppReviews`
- `CollectAppReviews`
- `GetAppDetailsBatch`
- default traffic class: `TrafficClassPublicStorePage`

`GetAppDetails` includes typed high-value Store fields such as capsule URLs,
screenshots, movies/trailers, background URLs, highlighted achievements,
recommendations, Metacritic, support info, content descriptors, and ratings raw
JSON. Use `GetAppDetailsRaw` when you need fields not yet typed by the SDK.
Use `AppDetailsData.DecodeRatings` for common rating board fields and
`AppDetailsData.SteamGermanyRequiredAge` for Steam Germany age requirements.

#### AppDetails identity warning

`/api/appdetails` is an undocumented, volatile Storefront Web endpoint. On
2026-09-26, response keys were observed to differ from the requested AppID even
when `data.steam_appid` correctly identified the requested application. The cause
is unverified; this is not a confirmed formal Valve contract change. See the
[upstream drift ledger](../governance/upstream-drift.md).

Do not resolve application identity through `envelope[requestedAppID]`.
Use `GetResolvedAppDetails` for ordinary single-application consumption, or
`storefront.ResolveAppDetails(envelope, appID)` when you already have an envelope.
The returned `AppDetailsMatch` includes `RequestedAppID`, the original string
`ResponseKey`, and the validated `Result`.

Resolution requires exactly one `data.steam_appid` match with `success=true`.
Missing or duplicate identity fails closed. If the direct requested-AppID key
exists but is unsuccessful, has no internal AppID, or contains a different AppID,
resolution fails even if another entry matches. No DLC/parent/child relationship
or first-entry fallback is inferred. AppID 0 returns `KindRequestBuild`;
identity-resolution failures use `KindAPIResponse`.

`GetAppDetailsRaw` preserves upstream bytes, and `GetAppDetails` preserves typed
upstream keys. `GetAppDetailsBatch` keeps its existing result structure and
per-item fetch errors: after checking `result.Err`, call
`storefront.ResolveAppDetails(result.Response, result.AppID)` explicitly.
Filters that omit `steam_appid` cannot establish identity and will fail resolution.

Doctor reports safely resolved key drift as WARN with requested AppID, response
key, and internal AppID; it adds no failure and keeps exit code 0 when all other
checks succeed. Unsafe resolution is FAIL. Live smoke accepts either key shape
and validates only the resolved identity.

The original `ReleaseDate` and `SupportedLanguages` fields remain unchanged.
Use the local helpers when you need calculable release-date precision or
structured language metadata:

```go
release := storefront.NormalizeReleaseDate(result.Data.ReleaseDate)
languages := storefront.ParseSupportedLanguages(result.Data.SupportedLanguages)
schinese, ok := storefront.LookupLanguage("schinese")
```

`NormalizeReleaseDate` recognizes exact days, months, quarters, years, and TBA
values. Unrecognized non-empty text is preserved with `unknown` precision.
`ParseSupportedLanguages` preserves unknown language names and the optional
full-audio marker instead of failing the surrounding Storefront response.

`GetAdjacentPartnerEvents` wraps the public Store events JSON endpoint used by
Steam partner news pages. The method exposes a stable typed subset and preserves
raw nested payloads for callers that need fields not yet typed by the SDK.

`CollectAppReviews` builds on `ListAppReviews` and accumulates reviews into
memory only when callers provide an explicit `MaxPages` or `MaxReviews` bound.

### `client.Web.Community`

- `GetInventory` / `GetInventoryRaw`
- `ListInventory`
- `JoinInventoryDescriptions`
- default traffic class: `TrafficClassCommunityWeb`
- inventory access may require caller-supplied cookies through `WithCookieJar(...)` or `WithDefaultCookieJar()`

`JoinInventoryDescriptions` is local-only. It pairs inventory assets with their
matching descriptions by `appid`, `classid`, and `instanceid` without issuing
network requests or performing market, trade, pricing, or account automation.

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
