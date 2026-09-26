# Cookbook：只读 Web 接口

使用 `client.Web.*` 调用少量受支持的只读 Steam Web JSON helper。

## Store 应用评论

```go
client, err := steam.NewClient(steam.WithSafeDefaults())
if err != nil {
	panic(err)
}
defer client.Close()

reviews, err := client.Web.Storefront.GetAppReviews(context.Background(), 440, &storefront.GetAppReviewsOptions{
	Language:   "english",
	NumPerPage: 20,
})
if err != nil {
	panic(err)
}

fmt.Println(reviews.QuerySummary.TotalReviews)
```

需要按 cursor 翻页时，使用 `ListAppReviews`。参见：[高价值只读 Helper](high-value-helpers.md)。

## 应用详情

```go
match, err := client.Web.Storefront.GetResolvedAppDetails(context.Background(), 440, &storefront.GetAppDetailsOptions{
	CountryCode: "US",
	Language:    "english",
})
if err != nil {
	panic(err)
}

fmt.Println(match.Result.Data.Name)
```

外层 response key 不能可靠地代表应用身份。该方法以 `data.steam_appid` 匹配，
在 `match.ResponseKey` 中保留原始 key，并在身份缺失、重复、结果失败或冲突时
拒绝解析。详见[已观察到的上游事故](../../governance/upstream-drift.md)。

已经持有 envelope 时，调用 `storefront.ResolveAppDetails(envelope, appID)`。
需要保留上游表示时使用 `GetAppDetails` 或 `GetAppDetailsRaw`，两者均不重写
Steam 的 key。批量查询使用 `GetAppDetailsBatch`，检查每项的 `result.Err` 后，
调用 `storefront.ResolveAppDetails(result.Response, result.AppID)`。
参见：[高价值只读 Helper](high-value-helpers.md)。

## 说明

- `client.Web.*` 不会注入 Steam Web API `key` 或 `access_token`。
- Go 方法签名稳定，但上游 Store / Community / Market payload 非官方且可能漂移。
- Inventory 可能需要调用方通过 `WithCookieJar(...)` 或 `WithDefaultCookieJar()` 提供 cookie。
- paginator 和 batch helper 复用底层单项方法的同一套请求控制。
