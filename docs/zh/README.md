# steam-go

[中文Wiki](https://github.com/GoFurry/steam-go/wiki/%E9%A6%96%E9%A1%B5) | 
[English](../../README.md)

`steam-go` 是一个专注于官方 Steam Web API 的轻量级 Go SDK。

`v1.0.0` 是 `steam-go` 的首个正式稳定版，定位为一个面向生产使用、专注于官方 Steam Web API 的 Go SDK。

## 特性

- 统一的根 `Client`，通过 `client.API.*` 分组访问各类服务
- 提供 API key、access token、timeout、retry、rate limit、proxy 的函数式配置
- 默认限制缓冲响应体大小，并可通过 `WithMaxResponseBodyBytes(...)` 调整
- `key` 和 `access_token` 被视为两种不同凭证，分别处理
- API key 可选，也可以通过轮换 key provider 提供
- 同时支持 typed/raw 两层返回
- 当 `WithAPIKeys(...)` 与 `WithRetry(...)` 一起使用时，`401/429` 可自动切换到下一个 key 重试
- 通过 addon 扩展能力，但不会把非 Web API 能力重新塞回核心 SDK

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

更详细的 API 分组说明请看 [api/reference.md](api/reference.md)。

项目治理文档：

- [兼容性策略](governance/compatibility.md)
- [Endpoint 稳定性](governance/endpoint-stability.md)
- [Endpoint 覆盖范围](governance/endpoint-coverage.md)
- [v1.0.0 Release Notes](releases/v1.0.0.md)

## WishlistService 覆盖范围

`client.API.WishlistService` 目前已经覆盖常用的愿望单查询流程：

- `GetWishlist`：获取某个 Steam 账号的愿望单条目列表
- `GetWishlistItemCount`：获取愿望单总数
- `GetWishlistItemsOnSale`：按国家/地区拉取正在打折的愿望单条目，并支持配置 `input_json` 的明细字段

其中 `GetWishlist` 和 `GetWishlistItemCount` 返回轻量的 typed 结构；`GetWishlistItemsOnSale` 则把每个 `store_item` 暴露为原始 JSON，以兼容 Steam 体积很大且变化频繁的商店数据结构。

## PlayerService 覆盖范围

`client.API.PlayerService` 目前已经覆盖一批比较实用的公开接口和需要身份凭证的资料/游戏行为接口，包括：

- 徽章、社区徽章进度、精选徽章、Steam 等级、Steam 等级分布
- 动态头像、头像边框、个人资料背景、小型个人资料背景、已装备资料物品、已拥有资料物品
- 资料展示、自定义项购买记录、已购买/已升级自定义项摘要、可用主题
- 昵称列表、玩家链接详情、好友游玩信息、最近游玩游戏、最近游玩时间、游戏热门成就

如果方法签名里显式要求传入 `accessToken` 或 `key`，就应该把调用者自己的凭证直接传给这个方法。`Client` 级别的全局凭证依然适合作为那些“不要求方法级显式凭证”的公共接口默认值。

## Addons

- `addons/a2s`：桥接到独立发布的 [`github.com/GoFurry/a2s-go`](https://github.com/GoFurry/a2s-go) `v1.0.1`
- `addons/openid`：用于 Steam OpenID 登录识别
- OpenID 只负责确认 Steam 身份并返回 `SteamID64`，不会替代 Web API 凭证
- 更详细的 addon 说明见 [addons/reference.md](addons/reference.md)

## Proxy

`steam-go` 继续把代理能力收敛在 `WithProxySelector(...)` 这一稳定扩展点上。

- `NewStaticProxySelector(...)`：固定代理
- `NewRoundRobinProxySelector(...)`：简单轮转
- `NewHealthCheckedRoundRobinProxySelector(...)`：基于失败评分和冷却时间的健康代理池
- `NewStickyProxySelector(...)`：按显式 session key 复用同一个代理
- `NewRoutingProxySelector(...)`：按 `host/path` 路由代理
- `NewHTTPClientWithProxySelector(...)`：给 addon 或独立 HTTP 流程复用
- `WithProxySessionKey(ctx, key)`：把粘性代理的会话键挂到请求上下文里
- `ProxyMetricsProvider`：读取健康代理池的内存快照指标
- 目前仍然不内建外部指标上报和重型代理池管理

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

## 流量类别

`steam-go` 现在支持按流量类别路由请求策略，让官方 Steam Web API 流量和后续公开商店页流量可以使用不同的请求配置。

- `TrafficClassOfficialAPI`：现有 typed `client.API.*` 方法的默认类别
- `TrafficClassPublicStorePage`：为后续公开商店页接入预留的类别
- `WithTrafficPolicy(...)`：按类别覆盖 proxy、cookie jar、retry、rate limit、短缓存、block 检测、header profile 和 Referer 策略
- `TransportHook` / `TransportHookFunc`：为某个流量类别预留 HTTP 执行栈扩展点，方便后续接自定义 TLS 或浏览器回退方案
- `WithTrafficClass(ctx, class)`：让单次请求显式切到非默认类别
- `DefaultPublicStoreHeaderProfileZH()` / `DefaultPublicStoreHeaderProfileEN()`：提供稳定的浏览器式请求头预设
- `WithRefererSource(ctx, rawURL)`、`NewStaticRefererSelector(...)`、`NewRoutingRefererSelector(...)`、`NewContextRefererSelector(...)`：提供固定、路由式和上下文来源式 Referer 策略
- `TrafficCachePolicy{TTL: ...}`：为某个流量类别启用进程内短 TTL 缓存，并在 `GET` 请求上自动使用 `ETag` / `Last-Modified` 做条件请求
- `TrafficBlockPolicy{HTMLSniffBytes: ...}`：为公开商店页流量启用 `429`、`403` 与 HTML challenge 的 block 检测与恢复链路

示例：

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

// typed Steam Web API 调用仍然建议走默认的 OfficialAPI 类别。
_, _ = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"76561198370695025"})

// TrafficClassPublicStorePage 预留给后续公开商店页请求入口使用。
storeCtx := steam.WithTrafficClass(context.Background(), steam.TrafficClassPublicStorePage)
_ = storeCtx
```

`TrafficClassPublicStorePage` 以及相关的 header / Referer / cache / block / transport hook 能力，目前的定位是“策略隔离与基础设施”，并不代表 `v1.0.0` 已经内置公开商店页抓取 API。

公开商店页请求画像示例：

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

公开商店页 transport hook 示例：

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

粘性代理示例：

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

健康代理池示例：

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

按 `host/path` 路由示例：

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

在中国网络环境下，浏览器里的 Steam OpenID 登录可能成功，但服务端的 `check_authentication` 校验请求仍然可能超时。`examples/openid` 支持通过 `--proxy http://127.0.0.1:7897` 处理这种情况，并且示例里已经加入了基于 cookie 的 `state` 校验流程。

## 示例

- `go run ./examples/a2s -server 1.2.3.4:27015 -query info`
- `go run ./examples/a2s -server 1.2.3.4:27015 -query players`
- `go run ./examples/a2s -server 1.2.3.4:27015 -query rules`
- `go run ./examples/openid`
- `go run ./examples/openid --proxy http://127.0.0.1:7897`
- `go run ./examples/proxy`
- `go run ./examples/steamuser`
- `go run ./examples/playerservice`
- `go run ./examples/steamuserstats`
- `go run ./examples/steamnews`
- `go run ./examples/live/steamuser`
- `go run ./examples/live/playerservice`
- `go run ./examples/live/wishlistservice`
- 完整 live smoke 列表见 [../../examples/live/README.md](../../examples/live/README.md)

## 错误处理

SDK 使用 `*steam.APIError`，主要错误类型有：

- `request_build`
- `transport`
- `http_status`
- `decode`
- `api_response`

可以通过 `errors.As(err, &apiErr)` 来检查错误类型、HTTP 状态码和原始响应体。
