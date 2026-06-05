# Cookbook：批量查询与翻页

这些 helper 用于常见只读 Web 工作流，同时让请求量保持显式可控。

## Reviews 翻页

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := client.Web.Storefront.ListAppReviews(
	ctx,
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

Paginator 每页调用一次 handler，不会把所有 reviews 聚合进内存。handler 应尽快处理并返回；需要提前停止时直接返回 error。

## Inventory 翻页

```go
err := client.Web.Community.ListInventory(
	ctx,
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

## 批量 App Details

```go
results, err := client.Web.Storefront.GetAppDetailsBatch(
	ctx,
	[]uint32{550, 440},
	&storefront.GetAppDetailsBatchOptions{MaxConcurrent: 2},
)
if err != nil {
	return err
}
for _, result := range results {
	if result.Err != nil {
		continue
	}
	fmt.Println(result.AppID)
}
```

## 批量 Market Prices

```go
prices, err := client.Web.Market.GetPriceOverviewBatch(
	ctx,
	[]market.PriceOverviewBatchItem{
		{AppID: 440, MarketHashName: "Mann Co. Supply Crate Key"},
	},
	&market.GetPriceOverviewBatchOptions{Currency: 1, MaxConcurrent: 2},
)
```

批量 helper 保持输入顺序。单项失败会写入该项的 `Err`；顶层 error 只用于参数无效、worker 初始化失败或 context cancellation 等全局失败。

## 生产建议

- 优先从 `steam.WithSafeDefaults()` 开始。
- 使用 `WithTrafficPolicy(...)` 配置不同 surface 的 retry、rate limit、cache、block detection、proxy 或 session control。
- `MaxConcurrent` 只限制 helper 内部并发，不等于安全请求速率。
- 每个 batch 或 paginator 调用都应带 context timeout 或 cancellation。
- 不希望无限遍历上游结果时，应显式设置 `MaxPages`。
