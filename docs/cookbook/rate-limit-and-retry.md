# Cookbook: Rate Limit and Retry

Start with `WithSafeDefaults()` for real external traffic, then tune per workload.

## Safe Defaults

```go
client, err := steam.NewClient(
	steam.WithAPIKey("your-key"),
	steam.WithSafeDefaults(),
)
if err != nil {
	panic(err)
}
defer client.Close()
```

`WithSafeDefaults()` currently applies a conservative retry and rate-limit preset.

## Explicit Retry

```go
client, err := steam.NewClient(
	steam.WithAPIKey("your-key"),
	steam.WithTimeout(10*time.Second),
	steam.WithRetry(2),
)
```

## Per-Class Policy

```go
client, err := steam.NewClient(
	steam.WithAPIKey("your-key"),
	steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
		RateLimiter: &steam.TrafficRateLimiterPolicy{
			Limit: 3,
			Burst: 3,
		},
		Retry: &steam.TrafficRetryPolicy{
			Retry: 2,
		},
	}),
)
```

## Notes

- Prefer lower concurrency for unofficial Web surfaces.
- `WithAPIKeys(...)` plus `WithRetry(...)` can retry `401/429` with another configured key.
- Use context cancellation for caller-level deadlines.
