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
- `client.API.SteamNews`
- `client.API.SteamUser`
- `client.API.SteamUserStats`

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
