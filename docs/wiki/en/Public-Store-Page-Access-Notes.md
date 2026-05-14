# Public Store Page Access Notes

Public Store pages are not the same as official Steam Web API endpoints.

They may be useful, but they should be treated as browser-like public web traffic, not stable typed API traffic.

Starting with `v1.1.0`, `steam-go` exposes a small read-only `client.Web.*` layer for a few high-value Storefront, Community, and Market JSON endpoints. These methods are stable Go APIs, but their upstream payloads should still be treated as unofficial or volatile web surfaces.

## Official API vs Public Store Page

| Official Web API | Public Store Page |
|---|---|
| JSON-oriented | HTML-oriented |
| API key may be required | Browser-like request may be needed |
| Better suited for typed SDK methods | Better suited for best-effort extraction |
| More stable request shape | Page structure may change |
| Usually backend API style | Public website behavior |

## Recommended SDK Features

For Store-page-like or adjacent web access, consider:

- `TrafficClassPublicStorePage`
- `TrafficClassCommunityWeb`
- `TrafficClassMarketWeb`
- `HeaderProfile`
- `RefererSelector`
- `TrafficCachePolicy`
- `TrafficBlockPolicy`
- `HostControl`
- `SessionControl`
- route-specific proxy policy

## Header Profile

```go
profile := steam.DefaultPublicStoreHeaderProfileEN()
```

or:

```go
profile := steam.DefaultPublicStoreHeaderProfileZH()
```

## Referer

```go
referer, err := steam.NewStaticRefererSelector(
    "https://store.steampowered.com/search/",
)
```

## Short Cache

```go
Cache: &steam.TrafficCachePolicy{
    TTL: time.Minute,
}
```

## Block Detection

```go
BlockPolicy: &steam.TrafficBlockPolicy{
    HTMLSniffBytes: 4096,
}
```

## Practical Advice

- Prefer built-in `client.Web.Storefront.*`, `client.Web.Community.*`, and `client.Web.Market.*` methods when the SDK already exposes the JSON endpoint you need.
- Do not treat HTML pages as stable APIs.
- Keep parsing code isolated from the core Web API client.
- Use cache for repeated reads.
- Use low concurrency.
- Watch for `403`, `429`, and HTML challenge pages.
- Prefer official Web API methods when available.

