# Traffic Policy

`TrafficPolicy` lets different request classes use different network behavior.

It is one of the main engineering abstractions in `steam-go`.

## Traffic Classes

| Class | Meaning |
|---|---|
| `TrafficClassOfficialAPI` | Normal typed Steam Web API traffic |
| `TrafficClassPublicStorePage` | Storefront web JSON or browser-like Store traffic |
| `TrafficClassCommunityWeb` | Steam Community inventory-style web JSON traffic |
| `TrafficClassMarketWeb` | Steam Community Market web JSON traffic |

## Why It Exists

Official API requests and public Store page requests should not always share the same strategy.

Official API traffic usually cares about:

- API key
- retry
- rate limit
- response decoding
- typed responses

Public Store page traffic may care about:

- browser-like headers
- cookies
- Referer
- block detection
- short cache
- lower concurrency
- proxy routing

Community or Market web JSON traffic may also care about:

- separate throttling from official API calls
- cookie jar reuse
- short cache for repeated reads
- block detection on `403`, `429`, or challenge HTML

## Basic Example

```go
client, err := steam.NewClient(
    steam.WithTrafficPolicy(
        steam.TrafficClassPublicStorePage,
        steam.TrafficPolicy{
            RateLimiter: &steam.TrafficRateLimiterPolicy{
                Limit: 10,
                Burst: 10,
            },
        },
    ),
)
```

## Store-Page-Oriented Example

```go
profile := steam.DefaultPublicStoreHeaderProfileEN()

client, err := steam.NewClient(
    steam.WithTrafficPolicy(
        steam.TrafficClassPublicStorePage,
        steam.TrafficPolicy{
            HeaderProfile: &profile,
            Cache: &steam.TrafficCachePolicy{
                TTL: time.Minute,
            },
            BlockPolicy: &steam.TrafficBlockPolicy{
                HTMLSniffBytes: 4096,
            },
        },
    ),
)
```

## Request-Level Class Selection

```go
ctx := steam.WithTrafficClass(
    context.Background(),
    steam.TrafficClassPublicStorePage,
)
```

In normal usage, the built-in `client.Web.*` methods already choose their default traffic class automatically:

- `client.Web.Storefront.*` -> `TrafficClassPublicStorePage`
- `client.Web.Community.*` -> `TrafficClassCommunityWeb`
- `client.Web.Market.*` -> `TrafficClassMarketWeb`

## Policy Surface

`TrafficPolicy` can override:

- proxy selector
- cookie jar
- rate limiter
- retry policy
- host control
- session control
- short cache
- block detection
- header profile
- Referer selector
- transport hook

## Design Rule

Keep traffic behavior close to the SDK boundary.

Avoid pushing rate limits, cookies, proxy decisions, and browser-like headers into business code.

