# Cookbook：高价值只读 Helper

这些 helper 覆盖常见只读工作流，但不会把 `client.Web.*` 扩成完整
Store 或 Community SDK。

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

