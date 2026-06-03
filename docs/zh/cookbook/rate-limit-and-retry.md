# Cookbook：限流与重试

真实外部流量建议先使用 `WithSafeDefaults()`，再按业务负载调整。

## 安全默认值

```go
client, err := steam.NewClient(
	steam.WithAPIKey("your-key"),
	steam.WithSafeDefaults(),
)
if err != nil {
	panic(err)
}
defer client.Close()
```

`WithSafeDefaults()` 当前会启用保守的 retry 与 rate limit 预设。

## 显式 Retry

```go
client, err := steam.NewClient(
	steam.WithAPIKey("your-key"),
	steam.WithTimeout(10*time.Second),
	steam.WithRetry(2),
)
```

## 按流量类别配置

```go
client, err := steam.NewClient(
	steam.WithAPIKey("your-key"),
	steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
		RateLimiter: &steam.TrafficRateLimiterPolicy{
			Limit: 3,
			Burst: 3,
		},
		Retry: &steam.TrafficRetryPolicy{
			Retry: 2,
		},
	}),
)
```

## 说明

- 非官方 Web surface 建议使用更低并发。
- `WithAPIKeys(...)` 与 `WithRetry(...)` 配合时，可以在 `401/429` 后尝试下一个 key。
- 使用 context cancellation 控制调用方级别的截止时间。
