# steam-go

[English](../README.md)

`steam-go` 是一个聚焦 Steam Web API 的轻量 Go SDK。

## 特性

- 统一的根 `Client`，通过 `client.API.*` 分组访问服务
- 提供 API key、access token、timeout、retry、rate limit、proxy 的函数式配置
- `key` 和 `access_token` 作为两种不同凭证独立处理
- 支持 typed/raw 双层返回
- 在 `WithAPIKeys(...)` 和 `WithRetry(...)` 一起使用时，`401/429` 可以自动轮换到下一个 key 重试
- 通过 addon 扩展能力，但不把非 Web API 能力重新塞回核心 SDK

## 安装

```bash
go get github.com/GoFurry/steam-go@latest
```

## 快速开始

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

详细 API 分组请看 [api.md](api.md)。

## Addons

- `addons/a2s`：桥接到独立发布的 [`github.com/GoFurry/a2s-go`](https://github.com/GoFurry/a2s-go) `v1.0.1`
- `addons/openid`：用于 Steam OpenID 登录识别
- OpenID 只负责确认 Steam 身份并返回 `SteamID64`，不会替代 Web API 凭证
- 更详细的 addon 说明见 [addons.md](addons.md)

## Proxy

`steam-go` 继续把代理能力收敛在 `WithProxySelector(...)` 这一个稳定扩展点上。

- `NewStaticProxySelector(...)`：固定代理
- `NewRoundRobinProxySelector(...)`：简单轮转
- `NewRoutingProxySelector(...)`：按 `host/path` 路由代理
- `NewHTTPClientWithProxySelector(...)`：给 addon 或独立 HTTP 流程复用
- 目前仍然不内建健康检查、熔断和重型代理池管理

固定代理示例：

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

按 host/path 路由示例：

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

中国网络环境下，浏览器里的 Steam OpenID 登录可能成功，但服务端的 `check_authentication` 校验请求仍然可能超时。`examples/openid` 支持通过 `--proxy http://127.0.0.1:7897` 来处理这种情况，并且示例里已经加入了基于 cookie 的 `state` 校验流程。

## 示例

- `go run ./examples/a2s -server 1.2.3.4:27015 -query info`
- `go run ./examples/a2s -server 1.2.3.4:27015 -query players`
- `go run ./examples/a2s -server 1.2.3.4:27015 -query rules`
- `go run ./examples/openid`
- `go run ./examples/openid --proxy http://127.0.0.1:7897`
- `go run ./examples/proxy`
- `go run ./test`

## 错误处理

SDK 使用 `*steam.APIError`，主要错误类型有：

- `request_build`
- `transport`
- `http_status`
- `decode`
- `api_response`

可以通过 `errors.As(err, &apiErr)` 来检查错误类型、HTTP 状态码和原始响应体。
