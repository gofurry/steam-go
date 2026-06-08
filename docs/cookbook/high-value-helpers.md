# Cookbook: High-Value Read-Only Helpers

These helpers cover common read-only workflows without expanding `client.Web.*`
into a full Store or Community SDK.

This page is a quick overview. For production boundaries, see
[Batch and pagination](batch-and-pagination.md) and
[Request observability](observability.md).

## Reviews Paginator

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

## Bounded Reviews Collector

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

`CollectAppReviews` requires `MaxPages` or `MaxReviews`. It is useful for small
bounded reads; use `ListAppReviews` when you want streaming page-by-page
processing without retaining reviews in memory.

## Inventory Paginator

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

Inventory pagination is read-only. It does not log in, refresh cookies, or
guarantee access to private inventories.

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

The join helper is local-only. It pairs `assets` with `descriptions` and does
not fetch prices, inspect market state, trade, sell, or automate account
behavior.

## Batch Lookups

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

Batch helpers keep input order and return per-item errors.

`MaxConcurrent` only limits helper-local concurrency. Pair batch helpers with
`WithSafeDefaults()`, `WithTrafficPolicy(...)`, and context timeouts for
request rate, retry, and cancellation behavior.

## Request Observer

```go
client, err := steam.NewClient(
	steam.WithRequestObserver(steam.RequestObserverFunc(func(event steam.RequestEvent) {
		fmt.Println(event.TrafficClass, event.Path, event.StatusCode, event.Duration)
	})),
)
```

Observer events are sanitized. They do not include raw query strings, headers,
bodies, API keys, tokens, cookies, or proxy passwords. Keep the callback fast;
do not perform slow logging or network I/O inside the observer.

For async channel observers, panic-safe wrappers, and metrics label guidance,
see [Request observability](observability.md).
