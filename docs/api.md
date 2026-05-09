# API Reference

This directory keeps reference-style documentation that is too detailed for the repository homepage.

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

## Credential Notes

- `key` and `access_token` are treated as different credentials.
- If a method explicitly accepts `key` or `accessToken`, pass the caller-specific credential to that method.
- Client-level credentials still act as defaults for shared/public endpoints that do not require explicit method-level credentials.

## Addons

- `addons/a2s`
- `addons/a2s/master`
- `addons/a2s/scanner`
- `addons/openid`
- [Addon usage notes](addons.md)

## Proxy Helpers

- `WithProxySelector(selector ProxySelector)`
- `NewStaticProxySelector(rawURL string)`
- `NewRoundRobinProxySelector(rawURLs ...string)`
- `NewRoutingProxySelector(routes ...ProxyRoute)`
- `NewHTTPClientWithProxySelector(selector ProxySelector, timeout time.Duration)`

## Examples

- `go run ./examples/a2s -server 1.2.3.4:27015 -query info`
- `go run ./examples/a2s -server 1.2.3.4:27015 -query players`
- `go run ./examples/a2s -server 1.2.3.4:27015 -query rules`
- `go run ./examples/openid`
- `go run ./examples/openid --proxy http://127.0.0.1:7897`
- `go run ./examples/proxy`
