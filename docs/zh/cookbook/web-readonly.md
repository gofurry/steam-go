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
details, err := client.Web.Storefront.GetAppDetails(context.Background(), 440, &storefront.GetAppDetailsOptions{
	CountryCode: "US",
	Language:    "english",
})
if err != nil {
	panic(err)
}

if app, ok := details["440"]; ok && app.Success {
	fmt.Println(app.Data.Name)
}
```

需要查询多个 AppID 时，使用 `GetAppDetailsBatch`。参见：[高价值只读 Helper](high-value-helpers.md)。

## 说明

- `client.Web.*` 不会注入 Steam Web API `key` 或 `access_token`。
- Go 方法签名稳定，但上游 Store / Community / Market payload 非官方且可能漂移。
- Inventory 可能需要调用方通过 `WithCookieJar(...)` 或 `WithDefaultCookieJar()` 提供 cookie。
- paginator 和 batch helper 复用底层单项方法的同一套请求控制。
