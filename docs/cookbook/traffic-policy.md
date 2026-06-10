# Cookbook: Traffic Policy

Use `WithTrafficPolicy(...)` when different Steam surfaces need different retry, rate limit, cache, proxy, cookie, host, or session behavior.

## Baseline

```go
client, err := steam.NewClient(
	steam.WithSafeDefaults(),
)
```

`WithSafeDefaults()` is a good starting point for external traffic, but production systems should still tune policies by surface.

## Official API

Official API calls are usually the most stable surface. Keep retries modest and prefer explicit caller deadlines:

```go
client, err := steam.NewClient(
	steam.WithAPIKey("your-key"),
	steam.WithTrafficPolicy(steam.TrafficClassOfficialAPI, steam.TrafficPolicy{
		Retry: &steam.TrafficRetryPolicy{
			Retry:   2,
			Backoff: steam.DefaultRetryBackoffConfig(),
		},
	}),
)
```

## Storefront Web

Public Store pages are unofficial and more rate-limit sensitive. Start conservatively:

```go
client, err := steam.NewClient(
	steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
		RateLimiter: &steam.TrafficRateLimiterPolicy{
			Limit: 0.5,
			Burst: 1,
		},
		HostControl: &steam.TrafficHostControlPolicy{
			MaxConcurrent: 1,
		},
		Cache: &steam.TrafficCachePolicy{
			TTL: 5 * time.Minute,
		},
	}),
)
```

## Community Web

Community inventory access depends on privacy settings and upstream behavior. Use low concurrency and explicit pagination bounds:

```go
client, err := steam.NewClient(
	steam.WithTrafficPolicy(steam.TrafficClassCommunityWeb, steam.TrafficPolicy{
		RateLimiter: &steam.TrafficRateLimiterPolicy{
			Limit: 1,
			Burst: 1,
		},
		HostControl: &steam.TrafficHostControlPolicy{
			MaxConcurrent: 1,
		},
	}),
)
```

## Market Web

Market JSON endpoints can be sensitive to repeated calls. Keep request rate and batch helper concurrency low:

```go
client, err := steam.NewClient(
	steam.WithTrafficPolicy(steam.TrafficClassMarketWeb, steam.TrafficPolicy{
		RateLimiter: &steam.TrafficRateLimiterPolicy{
			Limit: 1,
			Burst: 1,
		},
		HostControl: &steam.TrafficHostControlPolicy{
			MaxConcurrent: 1,
		},
	}),
)
```

## Proxy and Region

Use a proxy selector when your deployment region needs different egress behavior:

```go
selector, err := steam.NewHealthCheckedRoundRobinProxySelector(
	steam.DefaultProxyHealthConfig(),
	"http://127.0.0.1:7897",
)
if err != nil {
	return err
}

client, err := steam.NewClient(
	steam.WithProxySelector(selector),
)
```

For session-like workflows, attach a session key so sticky selectors can keep related requests on the same proxy:

```go
ctx := steam.WithRequestSessionKey(context.Background(), "user:76561198000000000")
```

Do not put API keys, access tokens, cookies, or proxy passwords in session keys.

## Boundaries

- Traffic policy is client-side pressure control, not a way to bypass Steam limits.
- Cache applies only to eligible `GET` requests.
- Block detection is intended for unofficial Web traffic.
- Retry is method-aware: non-idempotent methods retry only when the SDK or raw request explicitly opts in.
- Request observers receive sanitized events and should stay fast.
