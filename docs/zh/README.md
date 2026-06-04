# steam-go

<p align="center">
  <img src="https://img.shields.io/badge/License-MIT-6C757D?style=flat&color=3B82F6" alt="License">&nbsp&nbsp&nbsp
  <img src="https://img.shields.io/github/v/release/gofurry/steam-go?style=flat&color=blue" alt="Release">&nbsp&nbsp&nbsp
  <img src="https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=flat&logo=go&logoColor=white" alt="Go Version">&nbsp&nbsp&nbsp
  <a href="https://goreportcard.com/report/github.com/gofurry/steam-go">
    <img src="https://goreportcard.com/badge/github.com/gofurry/steam-go" alt="Go Report Card">
  </a>&nbsp&nbsp&nbsp
  <img src="https://img.shields.io/badge/weekend-project-8B5CF6?style=flat" alt="Weekend Project">&nbsp&nbsp&nbsp
  <img src="https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange" alt="Made with Love">
</p>

<p align="center">
  ⭐
  <a href="https://github.com/gofurry/steam-go/wiki/%E9%A6%96%E9%A1%B5">中文Wiki</a>
  &nbsp;|&nbsp;
  <a href="https://github.com/gofurry/steam-go/wiki/Steam-Key-%E4%B8%8E-Access-Token">Steam Key 与 Access Token</a>
  &nbsp;|&nbsp;
  <a href="https://github.com/gofurry/steam-go">English</a>
  &nbsp;|&nbsp;
  <a href="https://github.com/gofurry/steam-go/wiki/Steam-Keys-and-Access-Tokens">Steam Keys and Access Tokens</a>
  ⭐
</p>

```text
				____ ____ ____ _  _ ____ ____ _   _   / ____ ___ ____ ____ _  _    ____ ____ 
				| __ |  | |___ |  | |__/ |__/  \_/   /  [__   |  |___ |__| |\/| __ | __ |  | 
				|__] |__| |    |__| |  \ |  \   |   /   ___]  |  |___ |  | |  |    |__] |__| 
                                                                             
```

`steam-go` 是一个面向官方 Steam Web API 的稳定 Go SDK，提供生产可用的请求控制能力，并谨慎扩展少量高价值、只读的 Steam Web 辅助能力。

## 为什么使用

- 稳定的根 `Client`，通过 `client.API.*` 分组访问官方 API
- 只读 `client.Web.*` 辅助能力，覆盖 Storefront、Community inventory 和 Market JSON 接口
- 用 functional options 配置 API key、access token、timeout、retry、rate limit、proxy、cookie 与 traffic policy
- 提供适合外部请求的安全默认值、响应体大小上限和 URL 脱敏 helper
- 稳定 payload 使用 typed response，高波动 subtree 使用 raw method 或 `json.RawMessage`
- 支持 API key 轮换与健康检查，降低 `401/429` 对重试链路的影响
- OpenID、websession、freeclaim、assets、A2S 等能力通过 addon 扩展，不膨胀核心 SDK

## 安装

```bash
go get github.com/gofurry/steam-go@latest
```

`steam-go` 当前要求 Go 1.25+。

## 快速开始：官方 API

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

更多官方 API 示例：[cookbook/basic-api.md](cookbook/basic-api.md)。

## 快速开始：Steam OpenID

```go
verifier, err := openid.NewVerifier(openid.Config{
	Realm:    "https://example.com/",
	ReturnTo: "https://example.com/auth/steam/callback",
})
if err != nil {
	panic(err)
}

loginURL, err := verifier.LoginURL("csrf-state")
if err != nil {
	panic(err)
}

fmt.Println(loginURL)
```

真实应用中，应在跳转前把 `state` 存到安全 cookie 或服务端 session。Cookbook：[cookbook/auth-openid.md](cookbook/auth-openid.md)。

## 快速开始：只读 Web 查询

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

`client.Web.*` 只读，并且不会注入 Steam Web API `key` 或 `access_token`。Cookbook：[cookbook/web-readonly.md](cookbook/web-readonly.md)。

## 生产建议

- 真实外部流量建议先用 `WithSafeDefaults()`，再根据场景调整 timeout、retry 和 rate limit。
- 区域、host 或 session 行为需要不同网络路径时，使用 `WithProxySelector(...)` 或 `WithTrafficPolicy(...)`。
- 不要记录可能包含 `key`、`access_token`、cookie 或 proxy 凭据的原始 URL。优先使用 `steam.RedactSensitiveURL(...)`。
- 需要更严格响应体上限时，使用 `WithMaxResponseBodyBytes(...)`。
- live smoke 凭据和 Web cookie 不应进入 Git；示例通过环境变量或隐藏输入读取敏感值。

## 稳定性边界

- `client.API.*` 是官方 Steam Web API surface，也是 `v1` 稳定承诺的主线。
- `client.Web.*` 的 Go 方法签名稳定，但上游 Store / Community / Market payload 属于非官方、高波动 surface。
- 高波动 nested payload 可以继续保留为 `json.RawMessage`，避免强行 typed 化导致脆弱接口。
- 核心包不内置浏览器 fallback、购买、出售、交易或批量账号自动化。

## Addons

| Addon | 适用场景 |
|---|---|
| `addons/openid` | Steam OpenID 登录验证 |
| `addons/websession` | 手动 Steam网页登录态流程 |
| `addons/freeclaim` | 只读限免发现，以及显式单 package 领取 |
| `addons/assets` | Store / Library 资源 URL 构造、验证、读取和下载 |
| `addons/a2s` | 通过 `github.com/gofurry/a2s-go` 查询 A2S 服务器 |

详细 addon 说明：[addons/reference.md](addons/reference.md)。

## 文档

- [文档索引](../README.md)
- [API 参考](api/reference.md)
- [Generated API 覆盖报告](../api/coverage.generated.md)
- [API 覆盖差异](../api/coverage-diff.md)
- [Web 参考](web/reference.md)
- [Addon 参考](addons/reference.md)
- [Cookbook](cookbook/basic-api.md)
- [兼容性策略](governance/compatibility.md)
- [Endpoint 稳定性](governance/endpoint-stability.md)
- [Endpoint 覆盖范围](governance/endpoint-coverage.md)
- [新增官方 Endpoint](governance/official-endpoints.md)
- [Fixture 与 Smoke 维护](governance/fixtures.md)
- [凭据安全](security/credentials.md)
- [Release Checklist](../releases/checklist.md)
