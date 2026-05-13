# API 参考

本文档记录 `steam-go` 当前公开的 API 分组和使用边界。更完整的方法签名以源码和 Go 文档为准。

## API 分组

当前 `client.API.*` 下包含这些服务分组：

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

## 重点覆盖

`PlayerService` 覆盖了常用玩家资料、徽章、装饰物、最近游戏、成就进度、Steam 等级等接口。

`WishlistService` 覆盖：

- `GetWishlist`
- `GetWishlistItemCount`
- `GetWishlistItemsOnSale`

`GetWishlistItemsOnSale` 中的 `store_item` 使用 `json.RawMessage`，因为 Steam 商店 payload 很大且变化频繁。

## Raw Payload 策略

`steam-go` 使用三类建模策略：

- 稳定的官方 Web API payload 优先使用强类型结构体。
- 整体稳定但局部高波动的大字段，使用 `typed outer + json.RawMessage`。
- 高波动页面或客户端风格 payload，在结构稳定前优先保留 raw。

每个导出的 `json.RawMessage` 字段都应该说明为什么保持 raw。

## 凭证说明

- `key` 和 `access_token` 是不同凭证。
- 某些方法签名会显式要求 `key` 或 `accessToken`，这类凭证需要传给方法本身。
- 客户端级凭证适合作为默认配置，用于不要求调用方显式传凭证的方法。
- 日志中不要直接打印包含 `key` 或 `access_token` 的 URL，优先使用 `steam.RedactSensitiveURL(...)`。

## Addons

独立 addon 位于：

- `addons/openid`
- `addons/a2s`
- `addons/a2s/master`
- `addons/a2s/scanner`

## Proxy 和 Traffic Policy

根包提供代理、限流、重试、cookie jar、traffic class、请求控制和公开商店页策略基础能力。

`TrafficClassPublicStorePage` 目前只是策略隔离和请求基础设施，不代表 `v1.0.0` 已经内置完整公开商店页抓取 API。

## 示例

- 普通示例：`examples/`
- 真实外部 API smoke：`examples/live/`

`examples/live/` 需要真实凭证，不属于离线示例。
