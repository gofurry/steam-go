# Cookbook：请求观测

当你需要轻量请求日志或指标，但不想把 tracing / metrics 依赖放进 SDK 核心时，可以使用 `WithRequestObserver(...)`。

## 同步 Observer

```go
client, err := steam.NewClient(
	steam.WithRequestObserver(steam.RequestObserverFunc(func(event steam.RequestEvent) {
		fmt.Println(event.TrafficClass, event.Method, event.Path, event.StatusCode, event.Duration)
	})),
)
```

callback 会在 SDK 请求完成后同步执行。这里应保持轻量，不要做慢文件日志、网络 I/O、阻塞式 metrics push 或昂贵格式化。

## 异步 Channel Observer

当事件处理可能比请求慢时，可以使用带 buffer 的 channel：

```go
events := make(chan steam.RequestEvent, 128)

go func() {
	for event := range events {
		recordMetric(event)
	}
}()

client, err := steam.NewClient(
	steam.WithRequestObserver(steam.RequestObserverFunc(func(event steam.RequestEvent) {
		select {
		case events <- event:
		default:
			// 丢弃事件，避免阻塞 SDK 请求。
		}
	})),
)
```

是否丢弃、阻塞或采样事件，应由应用显式决定。生产请求路径里，丢弃或采样通常比让 SDK 被 metrics backpressure 阻塞更安全。

## Panic-safe Wrapper

如果 observer 代码属于应用层，可以包一层防护，避免指标代码 panic 影响请求路径：

```go
func panicSafeObserver(next steam.RequestObserver) steam.RequestObserver {
	return steam.RequestObserverFunc(func(event steam.RequestEvent) {
		defer func() {
			_ = recover()
		}()
		next.ObserveRequest(event)
	})
}
```

## Metrics Label 建议

推荐的低基数字段：

- `traffic_class`
- `method`
- `host`
- `path`
- `status_bucket`，例如 `2xx`、`4xx`、`5xx`
- `error_kind`

不要把 raw query、用户 ID、SteamID、API key、token、cookie、proxy 凭据或响应 body 放进 label。`RequestEvent.Path` 不包含 raw query，但自定义 raw HTTP path 里仍可能包含 ID；作为 metrics label 前应先归一化或分桶。

## 安全边界

`RequestEvent` 包含 traffic class、method、host、path、status code、error kind、retry attempts、cache hit、conditional cache refresh hit、block detection 和 duration。

过期缓存通过 `304 Not Modified` 完成 conditional revalidation 时，observer 会收到 `CacheHit=true`、`ConditionalHit=true` 的事件。

它不包含 raw query、header、body、API key、access token、cookie 或 proxy password。
