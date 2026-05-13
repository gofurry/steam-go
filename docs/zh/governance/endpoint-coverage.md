# Endpoint 覆盖范围

本文档说明 `steam-go v1.0.0` 当前已经覆盖的官方 Steam Web API 范围。

## 覆盖策略

`v1.0.0` 的目标不是宣称“完整 Steam 覆盖”，而是稳定当前已经存在且已文档化的官方 Web API surface。

后续新的官方 API 覆盖可以在未来兼容版本中继续增加。

## 当前服务分组

当前仓库在 `client.API.*` 下暴露这些服务分组：

- `AccountCartService`
- `BillingService`
- `CommunityService`
- `FamilyGroupsService`
- `LoyaltyRewardsService`
- `MobileNotificationService`
- `NewsService`
- `PlayerService`
- `QuestService`
- `SaleFeatureService`
- `SteamApps`
- `SteamChartsService`
- `SteamDirectory`
- `SteamNews`
- `SteamNotificationService`
- `SteamUser`
- `SteamUserOAuth`
- `SteamUserStats`
- `SteamWebAPIUtil`
- `StoreBrowseService`
- `StoreCatalogService`
- `StorePreferencesService`
- `StoreService`
- `StoreTopSellersService`
- `UserAccountService`
- `UserReviewsService`
- `UserStoreVisitService`
- `WishlistService`

## 发布解释

对 `v1.0.0` 来说，关键点是：

- 当前官方 service group 已足够支撑首个稳定版本
- 缺失的官方 endpoint 不阻塞 `v1.0.0`
- 新官方 endpoint 可以在后续 `v1.x` 中兼容加入

## 非覆盖范围

以下内容不属于当前 endpoint 覆盖范围：

- 新的公开 Steam Store 页面抓取 API
- Steam Community 页面抓取 API
- CDN 或静态资源 helper API
- 未记录的网页 JSON endpoint

这些能力如果未来加入，应独立设计和记录，不应悄悄混入首个稳定版的承诺范围。
