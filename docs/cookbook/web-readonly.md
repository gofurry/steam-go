# Cookbook: Read-Only Web Endpoints

Use `client.Web.*` for the small supported set of read-only Steam Web JSON helpers outside the official Web API.

## Store App Reviews

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

For cursor pagination, use `ListAppReviews`. See [High-value read-only helpers](high-value-helpers.md).

## App Details

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

The outer response key is not a reliable application identity. This helper uses
`data.steam_appid`, preserves the original key in `match.ResponseKey`, and fails
closed on missing, duplicate, unsuccessful, or conflicting identity. See the
[observed upstream incident](../governance/upstream-drift.md).

For an existing envelope, call `storefront.ResolveAppDetails(envelope, appID)`.
Use `GetAppDetails` or `GetAppDetailsRaw` when you need upstream fidelity; neither
rewrites Steam's keys. For multiple AppIDs, use `GetAppDetailsBatch`, check each
`result.Err`, then call `storefront.ResolveAppDetails(result.Response, result.AppID)`.
See [High-value read-only helpers](high-value-helpers.md).

## Notes

- `client.Web.*` never injects Steam Web API `key` or `access_token`.
- These Go methods are stable, but upstream Store / Community / Market payloads are unofficial and may drift.
- Inventory may require caller-supplied cookies through `WithCookieJar(...)` or `WithDefaultCookieJar()`.
- Paginator and batch helpers reuse the same request controls as the underlying single-item methods.
