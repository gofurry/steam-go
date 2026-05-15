# steam-go

![License](https://img.shields.io/badge/License-MIT-6C757D?style=flat&color=3B82F6)
![Release](https://img.shields.io/github/v/release/gofurry/steam-go?style=flat&color=blue)
![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go&logoColor=white)
[![Go Report Card](https://goreportcard.com/badge/github.com/gofurry/steam-go)](https://goreportcard.com/report/github.com/gofurry/steam-go)

![Weekend Project](https://img.shields.io/badge/weekend-project-8B5CF6?style=flat)
![Made with Love](https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange)

[steam-go Wiki](https://github.com/gofurry/steam-go/wiki) | 
[Steam Keys and Access Tokens](https://github.com/gofurry/steam-go/wiki/Steam-Keys-and-Access-Tokens) | 
[中文文档](docs/zh/README.md) | 
[Steam Key 与 Access Token](https://github.com/gofurry/steam-go/wiki/Steam-Key-%E4%B8%8E-Access-Token)

`steam-go` is a lightweight Go SDK focused on the official Steam Web API.

`v1.0.0` is the first stable release of `steam-go`, positioned as a production-oriented Go SDK for the official Steam Web API.

`v1.1.0` adds a read-only `client.Web.*` layer for high-value Steam Storefront, Community, and Market JSON endpoints outside `api.steampowered.com`.

## Features

- Root `Client` with grouped service access under `client.API.*`
- Read-only `client.Web.*` access for Storefront app/package details, reviews, community inventory, and market price overview
- Functional options for API key, access token, timeout, retry, rate limit, and proxy selection
- Buffered response bodies are capped by default and can be tuned with `WithMaxResponseBodyBytes(...)`
- `key` and `access_token` are treated as different credentials and can be configured independently
- API key is optional and can be supplied through a rotating key provider
- `WithSafeDefaults()` enables a conservative retry + rate-limit preset for real external traffic
- `WithHealthCheckedAPIKeys(...)` adds temporary cooldown for keys that repeatedly hit `401/429`
- Typed responses by default with matching raw response methods
- `401/429` can automatically retry with the next API key when `WithAPIKeys(...)` and `WithRetry(...)` are used together
- Independent addons can extend the SDK without bloating the core Web API client

## Installation

```bash
go get github.com/gofurry/steam-go@latest
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"time"

	steam "github.com/gofurry/steam-go"
)

func main() {
	client, err := steam.NewClient(
		steam.WithAPIKey("your-key"),
		steam.WithTimeout(10*time.Second),
		steam.WithRetry(2),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	resp, err := client.API.SteamUser.GetPlayerSummaries(
		context.Background(),
		[]string{"76561198370695025"},
	)
	if err != nil {
		panic(err)
	}

	for _, player := range resp.Response.Players {
		fmt.Printf("%s: %s\n", player.SteamID, player.PersonaName)
	}
}
```

Detailed API group references live in [docs/api/reference.md](docs/api/reference.md).
Read-only web endpoint notes live in [docs/web/reference.md](docs/web/reference.md).
Project governance documents:

- [Documentation Index](docs/README.md)
- [Compatibility Policy](docs/governance/compatibility.md)
- [Endpoint Stability](docs/governance/endpoint-stability.md)
- [Endpoint Coverage](docs/governance/endpoint-coverage.md)
- [v1.0.0 Release Notes](docs/releases/v1.0.0.md)

## WishlistService Coverage

`client.API.WishlistService` currently covers the main wishlist lookup flows:

- `GetWishlist` for the wishlist item list of a Steam account
- `GetWishlistItemCount` for the total wishlist count
- `GetWishlistItemsOnSale` for detailed on-sale wishlist items with configurable `input_json` fields

`GetWishlist` and `GetWishlistItemCount` use typed lightweight responses. `GetWishlistItemsOnSale` keeps each `store_item` as raw JSON so the SDK can tolerate Steam's very large and frequently changing store payload.

## PlayerService Coverage

`client.API.PlayerService` already covers a useful mix of public and authenticated profile/gameplay endpoints, including:

- badges, community badge progress, favorite badge, Steam level, and Steam level distribution
- animated avatars, avatar frames, profile backgrounds, mini-profile backgrounds, equipped items, and owned profile items
- profile customization, purchased customizations, purchased/upgraded customization summaries, and available profile themes
- nickname lists, player link details, friends gameplay info, recently played games, last played times, and top achievements for games

When a method signature explicitly asks for `accessToken` or `key`, that credential must be passed to the method itself. Client-level credentials remain useful as defaults for endpoints that do not require caller-specific credentials in the method signature.

## Addons

- `addons/a2s` is a lightweight bridge to [`github.com/gofurry/a2s-go`](https://github.com/gofurry/a2s-go) `v1.0.1`
- `addons/openid` provides Steam OpenID login verification for browser-based sign-in flows
- `addons/websession` composes Steam authentication atoms into one manual web-login session flow
- `addons/freeclaim` searches Store promotions and can claim one free license with caller-supplied web cookies
- OpenID only confirms Steam identity and returns `SteamID64`; it does not replace Web API credentials
- `addons/freeclaim` stays read-only unless you pass the explicit claim flag in the example or your own code
- detailed addon notes live in [docs/addons/reference.md](docs/addons/reference.md)

## Web

`client.Web.*` covers a small set of high-value read-only JSON endpoints outside the official `api.steampowered.com` surface:

- `client.Web.Storefront.GetAppDetails`
- `client.Web.Storefront.GetPackageDetails`
- `client.Web.Storefront.GetAppReviews`
- `client.Web.Community.GetInventory`
- `client.Web.Market.GetPriceOverview`

These methods are part of the stable Go API surface in `v1.x`, but the upstream Store / Community / Market payloads remain unofficial or volatile web surfaces. `client.Web.*` never injects Steam Web API `key` or `access_token`; inventory access relies on caller-supplied cookies when required.

## Proxy

`steam-go` keeps proxy support centered on `WithProxySelector(...)`.

- `NewStaticProxySelector(...)` for one fixed proxy
- `NewRoundRobinProxySelector(...)` for simple rotation
- `NewHealthCheckedRoundRobinProxySelector(...)` for failure-based cooldown in one proxy pool
- `NewStickyProxySelector(...)` for explicit session-key based sticky proxy selection
- `NewRoutingProxySelector(...)` for host/path-based routing
- `NewHTTPClientWithProxySelector(...)` for addon or standalone HTTP flows
- `WithProxySessionKey(ctx, key)` for attaching one sticky session key to request context
- `ProxyMetricsProvider` for one in-memory health snapshot of a health-checked proxy pool
- no external metrics integration or heavy proxy-pool management

Static example:

```go
selector, err := steam.NewStaticProxySelector("http://127.0.0.1:7897")
if err != nil {
	panic(err)
}

client, err := steam.NewClient(
	steam.WithAPIKey("your-key"),
	steam.WithProxySelector(selector),
)
if err != nil {
	panic(err)
}
```

Routing example:

```go
selector, err := steam.NewRoutingProxySelector(
	steam.ProxyRoute{
		Host:       "api.steampowered.com",
		PathPrefix: "/ISteamUser/",
		ProxyURL:   "http://127.0.0.1:7897",
	},
	steam.ProxyRoute{
		Host:       "steamcommunity.com",
		PathPrefix: "/openid/",
		ProxyURL:   "",
	},
)
if err != nil {
	panic(err)
}
```

Sticky example:

```go
baseSelector, err := steam.NewRoundRobinProxySelector(
	"http://127.0.0.1:7897",
	"http://127.0.0.1:7898",
)
if err != nil {
	panic(err)
}

client, err := steam.NewClient(
	steam.WithAPIKey("your-key"),
	steam.WithProxySelector(steam.NewStickyProxySelector(baseSelector)),
)
if err != nil {
	panic(err)
}

ctx := steam.WithProxySessionKey(context.Background(), "browser-session-1")
_, err = client.API.SteamUser.GetPlayerSummaries(ctx, []string{"76561198370695025"})
if err != nil {
	panic(err)
}
```

Health-checked round-robin example:

```go
selector, err := steam.NewHealthCheckedRoundRobinProxySelector(
	steam.DefaultProxyHealthConfig(),
	"http://127.0.0.1:7897",
	"http://127.0.0.1:7898",
)
if err != nil {
	panic(err)
}

client, err := steam.NewClient(
	steam.WithAPIKey("your-key"),
	steam.WithProxySelector(selector),
)
if err != nil {
	panic(err)
}

metrics := selector.(steam.ProxyMetricsProvider).ProxyMetricsSnapshot()
fmt.Printf("healthy=%d cooling=%d\n", metrics.HealthyProxies, metrics.CoolingProxies)
```

## Traffic Classes

`steam-go` now supports per-class request policy routing so official Steam Web API traffic and the built-in `client.Web.*` surfaces can use different request strategies.

- `TrafficClassOfficialAPI` is the default for existing typed `client.API.*` methods
- `TrafficClassPublicStorePage` is used by `client.Web.Storefront.*`
- `TrafficClassCommunityWeb` is used by `client.Web.Community.*`
- `TrafficClassMarketWeb` is used by `client.Web.Market.*`
- `WithTrafficPolicy(...)` overrides proxy, cookie jar, retry, rate limit, short-cache, block detection, header profile, and Referer strategy per class
- `TransportHook` and `TransportHookFunc` reserve one per-class HTTP execution extension point for future TLS customization or browser-backed fallback
- `WithTrafficClass(ctx, class)` lets one request opt into a non-default class
- `DefaultPublicStoreHeaderProfileZH()` and `DefaultPublicStoreHeaderProfileEN()` provide stable browser-like header presets
- `WithRefererSource(ctx, rawURL)` plus `NewStaticRefererSelector(...)`, `NewRoutingRefererSelector(...)`, and `NewContextRefererSelector(...)` support fixed, routed, and context-driven Referer policies
- `TrafficCachePolicy{TTL: ...}` enables per-class in-memory short caching with `ETag` / `Last-Modified` revalidation for `GET` requests
- `TrafficBlockPolicy{HTMLSniffBytes: ...}` enables public store-page block detection for `429`, `403`, and HTML challenge responses

Example:

```go
client, err := steam.NewClient(
	steam.WithAPIKey("your-key"),
	steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
		RateLimiter: &steam.TrafficRateLimiterPolicy{
			Limit: 10,
			Burst: 10,
		},
	}),
)
if err != nil {
	panic(err)
}

// Keep typed Steam Web API calls on the default OfficialAPI class.
_, _ = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"76561198370695025"})

// Web requests route automatically by method.
_, _ = client.Web.Market.GetPriceOverview(context.Background(), 440, "Mann Co. Supply Crate Key", nil)
```

Public store-page profile example:

```go
profile := steam.DefaultPublicStoreHeaderProfileZH()
refererSelector, err := steam.NewStaticRefererSelector("https://store.steampowered.com/search/")
if err != nil {
	panic(err)
}

client, err := steam.NewClient(
	steam.WithAPIKey("your-key"),
	steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
		Cache:           &steam.TrafficCachePolicy{TTL: time.Minute},
		BlockPolicy:     &steam.TrafficBlockPolicy{},
		HeaderProfile:   &profile,
		RefererSelector: refererSelector,
	}),
)
if err != nil {
	panic(err)
}
```

Public store-page transport hook example:

```go
client, err := steam.NewClient(
	steam.WithAPIKey("your-key"),
	steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
		TransportHook: steam.TransportHookFunc(func(class steam.TrafficClass, base *http.Client) (*http.Client, error) {
			cloned := *base
			if transport, ok := base.Transport.(*http.Transport); ok {
				custom := transport.Clone()
				custom.TLSHandshakeTimeout = 5 * time.Second
				cloned.Transport = custom
			}
			return &cloned, nil
		}),
	}),
)
if err != nil {
	panic(err)
}
```

On China-region networks, browser login may succeed while the server-side Steam OpenID `check_authentication` request still times out. The OpenID example supports `--proxy http://127.0.0.1:7897` for that case and also demonstrates cookie-backed `state` verification on the callback.

## Examples

- `go run ./examples/a2s -server 1.2.3.4:27015 -query info`
- `go run ./examples/a2s -server 1.2.3.4:27015 -query players`
- `go run ./examples/a2s -server 1.2.3.4:27015 -query rules`
- `go run ./examples/openid`
- `go run ./examples/openid --proxy http://127.0.0.1:7897`
- `go run ./examples/websession`
- `go run ./examples/freeclaim`
- `go run ./examples/proxy`
- `go run ./examples/traffic`
- `go run ./examples/steamuser`
- `go run ./examples/playerservice`
- `go run ./examples/steamuserstats`
- `go run ./examples/steamnews`
- `go run ./examples/live/steamuser`
- `go run ./examples/live/playerservice`
- `go run ./examples/live/wishlistservice`
- full live smoke list: [examples/live/README.md](examples/live/README.md)

## Error Handling

SDK errors use `*steam.APIError` with these kinds:

- `request_build`
- `transport`
- `http_status`
- `decode`
- `api_response`

Use `errors.As(err, &apiErr)` to inspect kind, status code, and raw body.

Steam Web API credentials are injected through query parameters by default because that matches Steam's HTTP interface.
Avoid logging raw request URLs in production. Use `steam.RedactSensitiveURL(...)` before sending URLs to logs, traces, or monitoring systems.
