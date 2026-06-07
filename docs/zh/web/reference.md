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
- `GetAdjacentPartnerEvents` / `GetAdjacentPartnerEventsRaw`
- `ListAppReviews`
- `GetAppDetailsBatch`
- 默认 traffic class：`TrafficClassPublicStorePage`

`GetAppDetails` 已补充高价值商店字段的 typed 结构，包括 capsule URL、截图、
视频 / trailer、背景图、高亮成就、推荐数、Metacritic、支持信息、内容描述和
ratings 原始 JSON。需要 SDK 暂未 typed 的字段时，继续使用 `GetAppDetailsRaw`。

`GetAdjacentPartnerEvents` 封装公开 Store events JSON 接口，适合读取 Steam
商店新闻 / 活动页附近的 partner event。方法提供稳定 typed 子集，并保留嵌套
raw payload，方便调用方读取 SDK 暂未 typed 的字段。

### `client.Web.Community`

- `GetInventory` / `GetInventoryRaw`
- `ListInventory`
- 默认 traffic class：`TrafficClassCommunityWeb`
- 如果库存读取需要鉴权，可通过 `WithCookieJar(...)` 或 `WithDefaultCookieJar()` 提供 Cookie

### `client.Web.Market`

- `GetPriceOverview` / `GetPriceOverviewRaw`
- `GetPriceOverviewBatch`
- 默认 traffic class：`TrafficClassMarketWeb`

## 请求行为

- `client.Web.*` 不会注入 Steam Web API 的 `key` 或 `access_token`
- proxy、rate limit、retry、短缓存、block detection、header profile、referer policy 与 cookie jar 仍然复用现有 client option 体系
- 不提供内建登录、Cookie 刷新、浏览器 fallback、购买、出售、交易或其他账号自动化能力
- paginator 和 batch helper 复用底层单项方法的同一套请求控制
- `WithRequestObserver(...)` 只输出脱敏事件，不包含 raw query、header、body、凭据、cookie 或 proxy 密码
