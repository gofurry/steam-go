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
