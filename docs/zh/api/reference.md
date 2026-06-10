# API 参考

本文档记录 `steam-go` 当前公开的 API 分组和使用边界。更完整的方法签名以源码和 Go 文档为准。

## Generated 覆盖报告

- [Generated API 覆盖报告](../../api/coverage.generated.md)
- [API 覆盖差异](../../api/coverage-diff.md)

这些报告会对比 Steam 公开 `GetSupportedAPIList` inventory、SDK endpoint 常量和 `api/*` service methods。`extra_sdk` 是 drift 信号，不是自动删除要求。

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

## Web Helper

`client.Web.*` 提供一小组官方 Steam Web API 之外的只读 Web helper：

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

说明：

- Web helper 只读，不会注入 Steam Web API 的 `key` 或 `access_token`。
- paginator 和 batch helper 复用底层单项方法的 timeout、retry、rate limit、body cap、proxy、cookie jar 与 traffic policy。
- `CollectAppReviews` 必须显式设置 `MaxPages` 或 `MaxReviews`。
- `JoinInventoryDescriptions` 是纯本地 helper，不会发起请求。
- batch helper 保持输入顺序，并通过 per-item error 表示单项失败。
- Community inventory helper 不负责登录、不刷新 cookie，也不保证能访问 private inventory。

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
- 记录可能包含 `Authorization`、`Cookie`、`Set-Cookie`、proxy 凭据或 Web API key header 的 headers 前，优先使用 `steam.RedactSensitiveHeaders(...)`。

## Addons

独立 addon 位于：

- `addons/openid`
- `addons/a2s`
- `addons/a2s/master`
- `addons/a2s/scanner`
- `addons/assets`
- `addons/markup`
- `addons/vdf`
- `addons/websession`
- `addons/freeclaim`

说明：

- `addons/websession.NewClientFromSteamClient(...)` 和 `addons/freeclaim.NewClientFromSteamClient(...)` 会复用根 SDK 的按分类 `WithTrafficPolicy(...)` 执行链。
- `addons/assets` 提供纯 URL 构造 helper，以及显式调用的商店媒体发现、验证、读取和下载 helper；它不创建 client。
- `addons/vdf` 是对 `github.com/gofurry/vdf-go` 的轻量桥接；它只解析调用方显式提供的文本 VDF / KeyValues 输入，不自动扫描本机 Steam 安装目录。
- 旧的 addon `NewClient(...)` 构造器仍然保留，继续作为手动模式，依赖调用方提供的 `http.Client`、proxy、timeout、base URL 和 `CookieJar`。

## Proxy 和 Traffic Policy

根包提供代理、限流、重试、cookie jar、traffic class、请求控制和公开商店页策略基础能力。

`client.Web.Storefront.*` 默认使用 `TrafficClassPublicStorePage`，`client.Web.Community.*` 默认使用 `TrafficClassCommunityWeb`，`client.Web.Market.*` 默认使用 `TrafficClassMarketWeb`。

另外，`(*steam.Client).DoRawHTTPRequest(...)` 提供了一个面向 addon/raw HTTP 场景的公开入口，让非 typed 服务也能复用根 SDK 的按类别执行链。

Raw HTTP 只接受 absolute URL。不要把不可信的用户输入 URL 直接传给 `DoRawHTTPRequest(...)`；当 URL 来源不完全受控时，优先通过 `RawHTTPRequestOptions.HostPolicy`、`NewAllowedRawHTTPHostPolicy(...)` 或 `NewSteamRawHTTPHostPolicy()` 限制允许访问的 host。

Retry 会识别请求方法：`GET`、`HEAD`、`OPTIONS` 默认可重试；`POST`、`PUT`、`PATCH`、`DELETE` 等非幂等方法只有在 SDK 方法或 `RawHTTPRequestOptions.Retryable` 显式 opt-in 后才会自动重试。

`WithRequestObserver(...)` 可以安装轻量 request observer。事件包含 traffic class、method、host、不带 raw query 的 path、status、error kind、attempts、cache hit、block detected 和 duration；不包含 header、body、API key、token、cookie、raw query 或 proxy 密码。

## 示例

- 普通示例：`examples/`
- 真实外部 API smoke：`examples/live/`
- addon 手动示例：`go run ./examples/openid`
- addon 手动示例：`go run ./examples/websession`
- addon 手动示例：`go run ./examples/assets -app-ids 550,107100`
- addon 手动示例：`go run ./examples/vdf -file ./steamapps/appmanifest_730.acf -key AppState`
- addon 手动示例：`go run ./examples/freeclaim`

`examples/live/` 需要真实凭证，不属于离线示例。
