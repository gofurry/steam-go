# Rate Limiting Strategy

`steam-go` uses conservative request controls because Steam Web API usage has both documented and practical limits.

## Why Rate Limiting Exists

The Steam Web API Terms of Use limit applications to 100,000 calls per day.

In addition to documented daily limits, production callers should avoid aggressive bursts. External services may return `429`, degrade temporarily, or block abusive-looking traffic.

## SDK Strategy

`steam-go` provides multiple layers instead of one hardcoded limit:

- `WithSafeDefaults()`
- `WithRateLimit(...)`
- `WithRateLimiter(...)`
- `WithRetry(...)`
- `WithRetryBackoff(...)`
- `WithRetryRespectRetryAfter(...)`
- `WithHealthCheckedAPIKeys(...)`
- `TrafficRateLimiterPolicy`
- `HostControl`
- `SessionControl`

## Safe Default

For normal external traffic:

```go
client, err := steam.NewClient(
    steam.WithSafeDefaults(),
)
```

This is intentionally conservative and easy to override later.

## Custom Rate Limiter

For heavier workloads:

```go
client, err := steam.NewClient(
    steam.WithRetry(2),
    steam.WithRateLimiter(rate.Limit(5), 5),
    steam.WithRetryRespectRetryAfter(true),
)
```

## Traffic-Class Rate Limit

For public Store-page-like traffic:

```go
client, err := steam.NewClient(
    steam.WithTrafficPolicy(
        steam.TrafficClassPublicStorePage,
        steam.TrafficPolicy{
            RateLimiter: &steam.TrafficRateLimiterPolicy{
                Limit: 2,
                Burst: 2,
            },
        },
    ),
)
```

## Practical Advice

- Do not run unbounded goroutines against Steam APIs.
- Treat `429` as a signal to slow down.
- Prefer cache for repeated reads.
- Use lower concurrency for public Store pages.
- Keep official API traffic and Storefront page-like traffic in separate traffic classes and budgets.
- Use key rotation for resilience, not for bypassing policy.
- Avoid documenting unofficial fixed per-second limits as facts.
- Put limits near the SDK boundary, not deep inside business logic.

## Empirical Storefront AppDetails Observation

An internal experiment on 2026-06-07 tested Storefront `appdetails` with:

- `360` logical requests per run.
- `10` workers.
- no local Store interval.
- no local cooldown.
- no retry.
- no proxy.

Two runs, separated by roughly five minutes, both completed `228` successful requests and then failed the remaining `132` requests.

| Run ID | OK | Failed | HTTP 429 | HTTP 403 | Transport |
|---|---:|---:|---:|---:|---:|
| `20260607-183228` | 228 | 132 | 31 | 74 | 27 |
| `20260607-184100` | 228 | 132 | 34 | 98 | 0 |

Interpretation:

- `429` began around 220-230 Store appdetails requests.
- Continuing after `429` quickly led to `403` / block-detected responses.
- Waiting about five minutes was enough for a repeat test to reach a similar request-count boundary again.
- This supports a conservative Store budget around `150-250` requests per five minutes per egress identity.

For production Storefront appdetails, start with a conservative budget such as `1 request / 2 seconds` and `burst=1`, then tune only with cache, queues, and observed `429` / `403` behavior.

Detailed report: [docs/experiments/store-rate-limit-20260607.md](../../experiments/store-rate-limit-20260607.md).

## Suggested Production Defaults

| Scenario | Suggested Strategy |
|---|---|
| Small tool / CLI | `WithSafeDefaults()` |
| Backend service | `WithRetry(2)` + explicit rate limiter |
| Public Store page access | Separate Store traffic class + low RPS + cache + block detection |
| Large batch job | Queue + worker pool + global limiter |
| Multi-key setup | Health-checked key provider + retry |

## References

- [Steam Web API Overview](https://partner.steamgames.com/doc/webapi_overview?language=english)
- [Steam Web API Terms of Use](https://steamcommunity.com/dev/apiterms)
