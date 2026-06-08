# Cookbook: Batch and Pagination

Use these helpers for common read-only Web workflows while keeping request volume explicit.

## Reviews Paginator

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

The paginator calls the handler once per page and does not accumulate all reviews in memory. Keep the handler fast and return an error to stop early.

When you intentionally want a small in-memory result, use the bounded collector:

```go
collection, err := client.Web.Storefront.CollectAppReviews(
	ctx,
	550,
	&storefront.CollectAppReviewsOptions{MaxPages: 2, MaxReviews: 150},
)
```

`CollectAppReviews` rejects calls without `MaxPages` or `MaxReviews`.

## Inventory Paginator

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

Inventory pagination is read-only. It does not log in, refresh cookies, or guarantee access to private inventories.

Use `community.JoinInventoryDescriptions(response)` after `GetInventory` when
you need assets paired with their matching descriptions. The join is local-only
and preserves asset order.

## Batch App Details

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

## Batch Market Prices

```go
prices, err := client.Web.Market.GetPriceOverviewBatch(
	ctx,
	[]market.PriceOverviewBatchItem{
		{AppID: 440, MarketHashName: "Mann Co. Supply Crate Key"},
	},
	&market.GetPriceOverviewBatchOptions{Currency: 1, MaxConcurrent: 2},
)
```

Batch helpers preserve input order. A single item failure is returned in that item's `Err`; top-level errors are reserved for invalid options, setup errors, or context cancellation.

## Production Notes

- Start with `steam.WithSafeDefaults()`.
- Use `WithTrafficPolicy(...)` for per-surface retry, rate limit, cache, block detection, proxy, or session controls.
- `MaxConcurrent` only limits helper-local parallelism. It is not a safe request rate by itself.
- Use context timeout or cancellation for every batch or paginator call.
- Keep `MaxPages` or `MaxReviews` explicit for workflows that should not walk unbounded upstream result sets.
