# steam-go

![License](https://img.shields.io/badge/License-MIT-6C757D?style=flat&color=3B82F6)
![Release](https://img.shields.io/github/v/release/GoFurry/steam-go?style=flat&color=blue)
![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go&logoColor=white)
[![Go Report Card](https://goreportcard.com/badge/github.com/GoFurry/steam-go)](https://goreportcard.com/report/github.com/GoFurry/steam-go)

![Weekend Project](https://img.shields.io/badge/weekend-project-8B5CF6?style=flat)
![Made with Love](https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange)

[中文文档](docs/README_zh.md)

`steam-go` is a lightweight Go SDK focused on the Steam Web API.

## Features

- Root `Client` with grouped service access under `client.API.*`
- Functional options for API key, access token, timeout, retry, rate limit, and proxy selection
- `key` and `access_token` are treated as different credentials and can be configured independently
- API key is optional and can be supplied through a rotating key provider
- Typed responses by default with matching raw response methods
- `401/429` can automatically retry with the next API key when `WithAPIKeys(...)` and `WithRetry(...)` are used together
- Independent addons can extend the SDK without bloating the core Web API client

## Installation

```bash
go get github.com/GoFurry/steam-go@latest
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"time"

	steam "github.com/GoFurry/steam-go"
)

func main() {
	client, err := steam.NewClient(
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

Detailed API group references live in [docs/api.md](docs/api.md).

## Addons

- `addons/a2s` is a lightweight bridge to [`github.com/GoFurry/a2s-go`](https://github.com/GoFurry/a2s-go) `v1.0.1`
- `addons/openid` provides Steam OpenID login verification for browser-based sign-in flows
- OpenID only confirms Steam identity and returns `SteamID64`; it does not replace Web API credentials
- detailed addon notes live in [docs/addons.md](docs/addons.md)

## Proxy

`steam-go` keeps proxy support centered on `WithProxySelector(...)`.

- `NewStaticProxySelector(...)` for one fixed proxy
- `NewRoundRobinProxySelector(...)` for simple rotation
- `NewRoutingProxySelector(...)` for host/path-based routing
- `NewHTTPClientWithProxySelector(...)` for addon or standalone HTTP flows
- no built-in health checks, circuit breaking, or heavy proxy-pool management

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

On China-region networks, browser login may succeed while the server-side Steam OpenID `check_authentication` request still times out. The OpenID example supports `--proxy http://127.0.0.1:7897` for that case and also demonstrates cookie-backed `state` verification on the callback.

## Examples

- `go run ./examples/a2s -server 1.2.3.4:27015 -query info`
- `go run ./examples/a2s -server 1.2.3.4:27015 -query players`
- `go run ./examples/a2s -server 1.2.3.4:27015 -query rules`
- `go run ./examples/openid`
- `go run ./examples/openid --proxy http://127.0.0.1:7897`
- `go run ./examples/proxy`
- `go run ./test`

## Error Handling

SDK errors use `*steam.APIError` with these kinds:

- `request_build`
- `transport`
- `http_status`
- `decode`
- `api_response`

Use `errors.As(err, &apiErr)` to inspect kind, status code, and raw body.
