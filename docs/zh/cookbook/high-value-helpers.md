# Cookbook：高价值只读 Helper

这些 helper 覆盖常见只读工作流，但不会把 `client.Web.*` 扩成完整
Store 或 Community SDK。

本页是快速总览。生产边界请继续参考
[批量查询与翻页](batch-and-pagination.md) 和
[请求观测](observability.md)。

## Reviews 翻页

```go
err := client.Web.Storefront.ListAppReviews(
	context.Background(),
	550,
	&storefront.ListAppReviewsOptions{MaxPages: 3},
	func(page storefront.AppReviewsPage) error {
		for _, review := range page.Reviews {
			fmt.Println(review.RecommendationID)
		}
		return nil
	},
)
```

## 有界 Reviews 聚合

```go
collection, err := client.Web.Storefront.CollectAppReviews(
	context.Background(),
	550,
	&storefront.CollectAppReviewsOptions{MaxPages: 2, MaxReviews: 150},
)
if err != nil {
	panic(err)
}
for _, review := range collection.Reviews {
	fmt.Println(review.RecommendationID)
}
```

`CollectAppReviews` 必须设置 `MaxPages` 或 `MaxReviews`。它适合小规模有界读取；
如果希望逐页处理且不把 reviews 留在内存中，继续使用 `ListAppReviews`。

## Inventory 翻页

```go
err := client.Web.Community.ListInventory(
	context.Background(),
	"76561198370695025",
	730,
	"2",
	&community.ListInventoryOptions{MaxPages: 2},
	func(page community.InventoryPage) error {
		for _, asset := range page.Assets {
			fmt.Println(asset.AssetID)
		}
		return nil
	},
)
```

Inventory 翻页只读，不负责登录、不刷新 cookie，也不保证能访问 private inventory。

## Inventory Description Join

```go
inv, err := client.Web.Community.GetInventory(ctx, "76561198370695025", 730, "2", nil)
if err != nil {
	panic(err)
}
items := community.JoinInventoryDescriptions(inv)
for _, item := range items {
	if item.Description == nil {
		continue
	}
	fmt.Println(item.Asset.AssetID, item.Description.MarketHashName)
}
```

Join helper 是纯本地能力，只把 `assets` 与 `descriptions` 配对；它不获取价格、
不判断市场状态、不交易、不出售，也不做账号自动化。

## 批量查询

```go
details, err := client.Web.Storefront.GetAppDetailsBatch(
	context.Background(),
	[]uint32{550, 440},
	&storefront.GetAppDetailsBatchOptions{MaxConcurrent: 2},
)
if err != nil {
	panic(err)
}
for _, result := range details {
	if result.Err != nil {
		continue
	}
	fmt.Println(result.AppID)
}
```

```go
prices, err := client.Web.Market.GetPriceOverviewBatch(
	context.Background(),
	[]market.PriceOverviewBatchItem{
		{AppID: 440, MarketHashName: "Mann Co. Supply Crate Key"},
	},
	&market.GetPriceOverviewBatchOptions{Currency: 1, MaxConcurrent: 2},
)
```

批量 helper 保持输入顺序，并把单项失败放在 per-item `Err` 中。

`MaxConcurrent` 只限制 helper 内部并发。请求速率、重试和取消行为应配合
`WithSafeDefaults()`、`WithTrafficPolicy(...)` 和 context timeout 使用。

## Request Observer

```go
client, err := steam.NewClient(
	steam.WithRequestObserver(steam.RequestObserverFunc(func(event steam.RequestEvent) {
		fmt.Println(event.TrafficClass, event.Path, event.StatusCode, event.Duration)
	})),
)
```

Observer event 已脱敏，不包含 raw query、header、body、API key、token、cookie
或 proxy 密码。回调是同步执行的，应保持轻量，不要在里面做慢日志或网络 I/O。

异步 channel observer、panic-safe wrapper 和 metrics label 建议见
[请求观测](observability.md)。
