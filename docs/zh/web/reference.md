# Web 参考

`steam-go` 提供了一小层只读的 `client.Web.*`，用于接入官方 `api.steampowered.com` Web API 之外的高价值 Steam Web JSON 接口。

## 稳定性

- Go 方法签名属于稳定的 `v1.x` 公开接口
- 上游 Store / Community / Market payload 仍然属于非官方或高波动 Web surface
- 高波动嵌套字段可能保持部分 typed，并使用 `json.RawMessage`

## 服务分组

### `client.Web.Storefront`

- `GetAppDetails` / `GetAppDetailsRaw`
- `GetResolvedAppDetails` / 本地 `storefront.ResolveAppDetails`
- `GetPackageDetails` / `GetPackageDetailsRaw`
- `GetAppReviews` / `GetAppReviewsRaw`
- `GetAdjacentPartnerEvents` / `GetAdjacentPartnerEventsRaw`
- `ListAppReviews`
- `CollectAppReviews`
- `GetAppDetailsBatch`
- 默认 traffic class：`TrafficClassPublicStorePage`

`GetAppDetails` 已补充高价值商店字段的 typed 结构，包括 capsule URL、截图、
视频 / trailer、背景图、高亮成就、推荐数、Metacritic、支持信息、内容描述和
ratings 原始 JSON。需要 SDK 暂未 typed 的字段时，继续使用 `GetAppDetailsRaw`。
常见 rating board 字段可通过 `AppDetailsData.DecodeRatings` 读取，德国年龄限制可
通过 `AppDetailsData.SteamGermanyRequiredAge` 读取。

#### AppDetails 身份警告

`/api/appdetails` 是未公开承诺稳定的、易漂移的 Storefront Web endpoint。
2026-09-26 已观察到：外层 response key 与请求 AppID 不一致，但
`data.steam_appid` 仍正确标识请求的应用。原因尚未确认，不能据此声称 Valve
已正式修改契约。详见 [upstream drift 账本](../../governance/upstream-drift.md)。

不要通过 `envelope[requestedAppID]` 解析应用身份。普通单应用消费使用
`GetResolvedAppDetails`；已经持有 envelope 时，使用
`storefront.ResolveAppDetails(envelope, appID)`。返回的 `AppDetailsMatch`
包含 `RequestedAppID`、保留原样的字符串 `ResponseKey` 和已验证的 `Result`。

解析要求恰好一个 `data.steam_appid` 匹配，且该结果 `success=true`。
没有匹配或身份重复均失败。如果请求 AppID 对应的 direct key 存在，但它
`success=false`、缺少内部 AppID，或内部 AppID 冲突，即使另有有效匹配也必须
失败。不会推断 DLC/parent/child 关系，也不会直接取第一项。AppID 为 0 返回
`KindRequestBuild`，身份解析失败复用 `KindAPIResponse`。

`GetAppDetailsRaw` 保留原始 bytes，`GetAppDetails` 保留 typed upstream key。
`GetAppDetailsBatch` 保持原有结果结构和逐项抓取错误语义；检查 `result.Err`
后，显式调用 `storefront.ResolveAppDetails(result.Response, result.AppID)`。
如果 Filters 省略 `steam_appid`，则无法建立身份，解析会失败。

Doctor 对安全解析成功的 key drift 显示 WARN，并提供请求 AppID、response key
和内部 AppID；Fail 不增加，其他检查成功时退出码仍为 0。无法安全解析时才
显示 FAIL。Live smoke 接受正常或漂移的 key，仅验证解析后的身份。

原有 `ReleaseDate` 与 `SupportedLanguages` 字段保持不变。需要可计算的发行日期
精度或结构化语言 metadata 时，使用本地 helper：

```go
release := storefront.NormalizeReleaseDate(result.Data.ReleaseDate)
languages := storefront.ParseSupportedLanguages(result.Data.SupportedLanguages)
schinese, ok := storefront.LookupLanguage("schinese")
```

`NormalizeReleaseDate` 支持 exact day、month、quarter、year 和 TBA；无法识别的
非空文本会原样保留并标记为 `unknown`。`ParseSupportedLanguages` 会保留未知语言
名称与可选的 full-audio 标记，不会让外围 Storefront response 失败。

`GetAdjacentPartnerEvents` 封装公开 Store events JSON 接口，适合读取 Steam
商店新闻 / 活动页附近的 partner event。方法提供稳定 typed 子集，并保留嵌套
raw payload，方便调用方读取 SDK 暂未 typed 的字段。

`CollectAppReviews` 基于 `ListAppReviews` 聚合 reviews，但必须由调用方显式设置
`MaxPages` 或 `MaxReviews`，避免默认无限抓取。

### `client.Web.Community`

- `GetInventory` / `GetInventoryRaw`
- `ListInventory`
- `JoinInventoryDescriptions`
- 默认 traffic class：`TrafficClassCommunityWeb`
- 如果库存读取需要鉴权，可通过 `WithCookieJar(...)` 或 `WithDefaultCookieJar()` 提供 Cookie

`JoinInventoryDescriptions` 是纯本地 helper。它按 `appid`、`classid` 和
`instanceid` 把 inventory assets 与 descriptions 配对，不发起网络请求，也不做市场、
交易、定价或账号自动化。

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
