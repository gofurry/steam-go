# API Reference

This directory keeps reference-style documentation that is too detailed for the repository homepage.

## Generated Coverage

- [Generated API coverage](coverage.generated.md)
- [API coverage diff](coverage-diff.md)

The generated coverage reports compare Steam's public `GetSupportedAPIList` inventory with SDK endpoint constants and `api/*` service methods. `extra_sdk` is a drift signal, not an automatic removal request.

## API Groups

- `client.API.AccountCartService`
- `client.API.BillingService`
- `client.API.CommunityService`
- `client.API.FamilyGroupsService`
- `client.API.LoyaltyRewardsService`
- `client.API.MobileNotificationService`
- `client.API.NewsService`
- `client.API.PlayerService`
- `client.API.QuestService`
- `client.API.SaleFeatureService`
- `client.API.StoreBrowseService`
- `client.API.StoreCatalogService`
- `client.API.StorePreferencesService`
- `client.API.StoreService`
- `client.API.StoreTopSellersService`
- `client.API.SteamDirectory`
- `client.API.SteamApps`
- `client.API.SteamChartsService`
- `client.API.SteamNews`
- `client.API.SteamNotificationService`
- `client.API.SteamUser`
- `client.API.SteamUserOAuth`
- `client.API.SteamUserStats`
- `client.API.SteamWebAPIUtil`
- `client.API.UserAccountService`
- `client.API.UserReviewsService`
- `client.API.UserStoreVisitService`
- `client.API.WishlistService`

## Web Helpers

`client.Web.*` exposes a small read-only Web helper surface outside the official Steam Web API:

- `client.Web.Storefront.GetAppDetails` / `GetAppDetailsRaw`
- `client.Web.Storefront.GetPackageDetails` / `GetPackageDetailsRaw`
- `client.Web.Storefront.GetAppReviews` / `GetAppReviewsRaw`
- `client.Web.Storefront.ListAppReviews`
- `client.Web.Storefront.CollectAppReviews`
- `client.Web.Storefront.GetAppDetailsBatch`
- `client.Web.Community.GetInventory` / `GetInventoryRaw`
- `client.Web.Community.ListInventory`
- `community.JoinInventoryDescriptions`
- `client.Web.Market.GetPriceOverview` / `GetPriceOverviewRaw`
- `client.Web.Market.GetPriceOverviewBatch`

Notes:
- Web helpers are read-only and never inject Steam Web API `key` or `access_token`.
- Paginator and batch helpers reuse the same timeout, retry, rate limit, body cap, proxy, cookie jar, and traffic policy behavior as their underlying single-item methods.
- `CollectAppReviews` requires an explicit `MaxPages` or `MaxReviews` bound.
- `JoinInventoryDescriptions` is local-only and does not issue requests.
- Batch helpers preserve input order and report per-item errors.
- Community inventory helpers do not log in, refresh cookies, or guarantee access to private inventories.

## Selected Endpoint Coverage

These are not exhaustive lists, but they reflect the main typed SDK coverage available today.

### `client.API.PlayerService`

- `GetOwnedGames`
- `ClientGetLastPlayedTimes`
- `GetRecentlyPlayedGames`
- `GetAchievementsProgress`
- `GetTopAchievementsForGames`
- `GetAnimatedAvatar`
- `GetAvatarFrame`
- `GetMiniProfileBackground`
- `GetProfileBackground`
- `GetProfileItemsEquipped`
- `GetProfileItemsOwned`
- `GetBadges`
- `GetCommunityBadgeProgress`
- `GetCommunityPreferences`
- `GetFavoriteBadge`
- `GetFriendsGameplayInfo`
- `GetNicknameList`
- `GetPlayerLinkDetails`
- `GetProfileCustomization`
- `GetProfileThemesAvailable`
- `GetPurchasedProfileCustomizations`
- `GetPurchasedAndUpgradedProfileCustomizations`
- `GetSteamLevel`
- `GetSteamLevelDistribution`

### `client.API.MobileNotificationService`

- `GetUserNotificationCounts`

### `client.API.NewsService`

- `ConvertHTMLToBBCode`

### `client.API.QuestService`

- `GetCommunityInventory`
- `GetNumTradingCardsEarned`

### `client.API.SaleFeatureService`

- `GetFriendsSharedYearInReview`
- `GetUserYearAchievements`
- `GetUserYearInReview`

### `client.API.StoreBrowseService`

- `GetContentHubConfig`

### `client.API.StoreCatalogService`

- `GetDevPageLinks`

### `client.API.StorePreferencesService`

- `GetIgnoreList`

### `client.API.StoreService`

- `GetAppList`
- `GetGamesFollowed`
- `GetGamesFollowedCount`
- `GetMostPopularTags`
- `GetUserGameInterestState`

### `client.API.StoreTopSellersService`

- `GetCountryList`
- `GetWeeklyTopSellers`

### `client.API.SteamDirectory`

- `GetCMListForConnect`
- `GetSteamPipeDomains`

### `client.API.SteamApps`

- `GetSDRConfig`
- `GetServersAtAddress`
- `UpToDateCheck`

### `client.API.SteamChartsService`

- `GetBestOfYearPages`
- `GetGamesByConcurrentPlayers`
- `GetMonthTopAppReleases`
- `GetMostPlayedGames`
- `GetTopReleasesPages`
- `GetYearTopAppReleases`

### `client.API.SteamNews`

- `GetNewsForApp`

### `client.API.SteamNotificationService`

- `GetPreferences`
- `GetSteamNotifications`

### `client.API.SteamUser`

- `GetFriendList`
- `GetPlayerBans`
- `GetPlayerSummaries`
- `GetUserGroupList`

### `client.API.SteamUserOAuth`

- `GetFriendList`
- `GetUserSummaries`

### `client.API.SteamUserStats`

- `GetGlobalAchievementPercentagesForApp`
- `GetNumberOfCurrentPlayers`
- `GetPlayerAchievements`
- `GetSchemaForGame`
- `GetUserStatsForGame`

### `client.API.SteamWebAPIUtil`

- `GetServerInfo`
- `GetSupportedAPIList`

### `client.API.UserAccountService`

- `GetUserCountry`

### `client.API.UserReviewsService`

- `GetFriendsRecommendedApp`

### `client.API.UserStoreVisitService`

- `GetFrequentlyVisitedPages`
- `GetMostVisitedItemsOnStore`

### `client.API.WishlistService`

- `GetWishlist`
- `GetWishlistItemCount`
- `GetWishlistItemsOnSale`

Notes:
- `GetWishlist` and `GetWishlistItemCount` accept `steamid`.
- `GetWishlistItemsOnSale` accepts `accessToken`, `countryCode`, and optional `data_request` fields through `input_json`.
- The `store_item` field in `GetWishlistItemsOnSale` is intentionally exposed as `json.RawMessage` because the payload is large and Steam changes it frequently.

## Raw Payload Strategy

`steam-go` defaults to typed responses whenever one Steam payload is stable, small enough to maintain, and useful as a Go-first API.

Use `json.RawMessage` only when one payload or subtree is clearly volatile:

- the field is very large or deeply nested
- Steam changes the shape frequently
- language, region, login state, or experiments can materially reshape the field
- most callers do not need the whole subtree eagerly decoded

Preferred modeling order:

1. stable official payload: fully typed response
2. stable outer response with volatile subtree: typed outer struct plus `json.RawMessage` on the unstable field
3. highly volatile payload with unclear long-term shape: raw JSON first, then promote stable portions to typed fields later

Additional rules:

- Prefer `typed outer + raw subtree` over turning the entire response into `map[string]any`.
- Document every exported `json.RawMessage` field with the reason it is raw.
- Do not use `json.RawMessage` on stable official Web API fields just to avoid writing types.
- If a future volatile subtree starts showing long-term stability, it can be promoted from raw JSON to typed fields in a later release.

## Credential Notes

- `key` and `access_token` are treated as different credentials.
- If a method explicitly accepts `key` or `accessToken`, pass the caller-specific credential to that method.
- Client-level credentials still act as defaults for shared/public endpoints that do not require explicit method-level credentials.
- Steam expects many of these credentials on the URL query string. Do not log raw request URLs in production; use `steam.RedactSensitiveURL(...)` before emitting URLs to logs, traces, or monitoring.
- Use `steam.RedactSensitiveHeaders(...)` before logging headers that may contain `Authorization`, `Cookie`, `Set-Cookie`, proxy credentials, or Web API key headers.
- `WithSafeDefaults()` enables one conservative preset for external traffic: retry `2` with a `3 rps / burst 3` client-level limiter.
- `WithHealthCheckedAPIKeys(...)` keeps round-robin rotation but temporarily cools down keys that repeatedly fail with `401/429`, so one bad key does not keep poisoning every retry loop.

## Addons

- `addons/a2s`
- `addons/a2s/master`
- `addons/a2s/scanner`
- `addons/assets`
- `addons/markup`
- `addons/vdf`
- `addons/openid`
- `addons/websession`
- `addons/freeclaim`
- [Addon usage notes](../addons/reference.md)

Notes:
- `addons/websession.NewClientFromSteamClient(...)` and `addons/freeclaim.NewClientFromSteamClient(...)` reuse the root SDK per-class `WithTrafficPolicy(...)` execution stack.
- `addons/assets` has pure URL builders plus explicit Store media discovery, verification, read, and download helpers; it does not create a client.
- `addons/vdf` is a thin bridge to `github.com/gofurry/vdf-go`; it parses only caller-provided text VDF / KeyValues input and does not scan local Steam installations.
- The legacy addon `NewClient(...)` constructors remain manual mode and still rely on caller-supplied `http.Client`, proxy, timeout, base URL, and `CookieJar`.

## Proxy Helpers

- `ProxyHealthConfig`
- `ProxyMetricsProvider`
- `ProxyMetricsSnapshot`
- `ProxyEndpointMetrics`
- `DefaultProxyHealthConfig()`
- `ErrAllProxiesCoolingDown`
- `WithProxySelector(selector ProxySelector)`
- `WithProxySessionKey(ctx context.Context, key string) context.Context`
- `NewStaticProxySelector(rawURL string)`
- `NewRoundRobinProxySelector(rawURLs ...string)`
- `NewHealthCheckedRoundRobinProxySelector(cfg ProxyHealthConfig, rawURLs ...string)`
- `NewStickyProxySelector(base ProxySelector)`
- `NewRoutingProxySelector(routes ...ProxyRoute)`
- `NewHTTPClientWithProxySelector(selector ProxySelector, timeout time.Duration)`

Notes:
- `NewHealthCheckedRoundRobinProxySelector(...)` only targets explicit proxy pools in the first version.
- `ErrAllProxiesCoolingDown` means every proxy in that health-checked pool is still inside its cooldown window.
- `ProxyMetricsProvider` exposes one read-only in-memory snapshot for health-checked proxy pools.
- `WithProxySessionKey(...)` only affects selectors that explicitly support sticky session lookup.
- `NewStickyProxySelector(...)` is designed as a wrapper and can be composed with static, round-robin, or routing selectors.

## Traffic Policy Helpers

- `TrafficClass`
- `TrafficClassOfficialAPI`
- `TrafficClassPublicStorePage`
- `RetryBackoffConfig`
- `DefaultRetryBackoffConfig()`
- `HeaderProfile`
- `DefaultPublicStoreHeaderProfileZH()`
- `DefaultPublicStoreHeaderProfileEN()`
- `RefererSelector`
- `RefererRoute`
- `TransportHook`
- `TransportHookFunc`
- `RequestObserver`
- `RequestObserverFunc`
- `RequestEvent`
- `TrafficPolicy`
- `TrafficCachePolicy`
- `TrafficBlockPolicy`
- `TrafficRateLimiterPolicy`
- `TrafficRetryPolicy`
- `RawHTTPRequestOptions`
- `RawHTTPBlockResult`
- `RawHTTPResult`
- `WithTrafficPolicy(class TrafficClass, policy TrafficPolicy)`
- `WithRequestObserver(observer RequestObserver)`
- `WithTrafficClass(ctx context.Context, class TrafficClass) context.Context`
- `WithRefererSource(ctx context.Context, rawURL string) context.Context`
- `NewStaticRefererSelector(rawURL string)`
- `NewRoutingRefererSelector(routes ...RefererRoute)`
- `NewContextRefererSelector(fallback RefererSelector)`
- `(*Client).DoRawHTTPRequest(...)`

Notes:
- Existing typed `client.API.*` methods default to `TrafficClassOfficialAPI`.
- `client.Web.Storefront.*` defaults to `TrafficClassPublicStorePage`, `client.Web.Community.*` defaults to `TrafficClassCommunityWeb`, and `client.Web.Market.*` defaults to `TrafficClassMarketWeb`.
- `(*Client).DoRawHTTPRequest(...)` is intended for addon-style raw HTTP flows that still need the SDK's class-aware execution stack.
- `WithTrafficPolicy(...)` only overrides the fields you set; unset fields continue to use the client-level defaults.
- `TrafficCachePolicy` currently applies only to `GET` requests and uses in-memory short TTL caching with `ETag` / `Last-Modified` revalidation.
- `TrafficBlockPolicy` is currently supported only on `TrafficClassPublicStorePage` and detects `429`, `403`, and HTML challenge responses.
- `HeaderProfile` only fills missing request headers and does not override explicit values already set on the request.
- Referer selectors run before transport execution; an explicit `Referer` header on the request still wins.
- `TransportHook` runs during client construction after the class-specific base `http.Client` has already been assembled with timeout, proxy routing, and cookie jar settings.
- `WithRequestObserver(...)` emits sanitized request events after SDK requests complete. Events include traffic class, method, host, path without raw query, status, error kind, attempts, cache hit, block detection, and duration.
- Request observer events do not include headers, bodies, API keys, tokens, cookies, raw query strings, or proxy passwords.

## Examples

- `go run ./examples/a2s -server 1.2.3.4:27015 -query info`
- `go run ./examples/a2s -server 1.2.3.4:27015 -query players`
- `go run ./examples/a2s -server 1.2.3.4:27015 -query rules`
- `go run ./examples/openid`
- `go run ./examples/openid --proxy http://127.0.0.1:7897`
- `go run ./examples/websession`
- `go run ./examples/assets -app-ids 550,107100`
- `go run ./examples/vdf -file ./steamapps/appmanifest_730.acf -key AppState`
- `go run ./examples/freeclaim`
- `go run ./examples/proxy`
- `go run ./examples/traffic`
