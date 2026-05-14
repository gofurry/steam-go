# Web 参考

`steam-go` 提供了一小层只读的 `client.Web.*`，用于接入官方 `api.steampowered.com` Web API 之外的高价值 Steam Web JSON 接口。

## 稳定性

- Go 方法签名属于稳定的 `v1.x` 公开接口
- 上游 Store / Community / Market payload 仍然属于非官方或高波动 Web surface
- 高波动嵌套字段可能保持部分 typed，并使用 `json.RawMessage`

## 服务分组

### `client.Web.Storefront`

- `GetAppDetails` / `GetAppDetailsRaw`
- `GetPackageDetails` / `GetPackageDetailsRaw`
- `GetAppReviews` / `GetAppReviewsRaw`
- 默认 traffic class：`TrafficClassPublicStorePage`

### `client.Web.Community`

- `GetInventory` / `GetInventoryRaw`
- 默认 traffic class：`TrafficClassCommunityWeb`
- 如果库存读取需要鉴权，可通过 `WithCookieJar(...)` 或 `WithDefaultCookieJar()` 提供 Cookie

### `client.Web.Market`

- `GetPriceOverview` / `GetPriceOverviewRaw`
- 默认 traffic class：`TrafficClassMarketWeb`

## 请求行为

- `client.Web.*` 不会注入 Steam Web API 的 `key` 或 `access_token`
- proxy、rate limit、retry、短缓存、block detection、header profile、referer policy 与 cookie jar 仍然复用现有 client option 体系
- 不提供内建登录、Cookie 刷新、浏览器 fallback、购买、出售、交易或其他账号自动化能力
