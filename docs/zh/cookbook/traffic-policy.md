# Cookbook：Traffic Policy

当不同 Steam surface 需要不同 retry、rate limit、cache、proxy、cookie、host 或 session 行为时，使用 `WithTrafficPolicy(...)`。

## 基线

```go
client, err := steam.NewClient(
	steam.WithSafeDefaults(),
)
```

`WithSafeDefaults()` 适合作为真实外部流量的起点，但生产系统仍应按 surface 调整策略。

## 官方 API

官方 API 通常是最稳定的 surface。retry 保持克制，并优先使用调用方 deadline：

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

公开商店页面属于 unofficial surface，更容易受 rate limit 影响。建议从保守配置开始：

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

Community inventory 访问受隐私设置和上游行为影响。保持低并发，并显式限制分页规模：

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

Market JSON endpoint 对重复请求更敏感。请求速率和 batch helper 并发都应保守：

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

## 代理与区域网络

部署区域需要不同出口行为时，使用 proxy selector：

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

对于 session-like 工作流，可以附加 session key，让 sticky selector 尽量保持同一组请求使用同一代理：

```go
ctx := steam.WithRequestSessionKey(context.Background(), "user:76561198000000000")
```

不要把 API key、access token、cookie 或 proxy password 放进 session key。

## 边界

- Traffic policy 是客户端侧压力控制，不是绕过 Steam 限制的机制。
- Cache 只作用于符合条件的 `GET` 请求。
- Block detection 面向 unofficial Web traffic。
- Retry 会识别请求方法：非幂等方法只有在 SDK 或 raw request 显式 opt-in 后才会自动重试。
- Request observer 收到的是脱敏事件，回调应保持轻量。
