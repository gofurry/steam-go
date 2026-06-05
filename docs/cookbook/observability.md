# Cookbook: Request Observability

Use `WithRequestObserver(...)` when you need lightweight request metadata for logs or metrics without adding tracing or metrics dependencies to the SDK core.

## Synchronous Observer

```go
client, err := steam.NewClient(
	steam.WithRequestObserver(steam.RequestObserverFunc(func(event steam.RequestEvent) {
		fmt.Println(event.TrafficClass, event.Method, event.Path, event.StatusCode, event.Duration)
	})),
)
```

The callback runs synchronously after an SDK request completes. Keep it fast; avoid slow file logging, network I/O, blocking metrics pushes, or expensive formatting inside the callback.

## Async Channel Observer

Use a buffered channel when event processing may be slower than requests:

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
			// Drop instead of blocking SDK requests.
		}
	})),
)
```

Choose whether to drop, block, or sample events explicitly. For production request paths, dropping or sampling is usually safer than blocking the SDK on metrics backpressure.

## Panic-safe Wrapper

If observer code is owned by application code, wrap it so a metrics bug cannot panic the request path:

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

## Metrics Labels

Good low-cardinality labels:

- `traffic_class`
- `method`
- `host`
- `path`
- `status_bucket`, such as `2xx`, `4xx`, `5xx`
- `error_kind`

Avoid labels from raw query strings, user IDs, SteamIDs, API keys, tokens, cookies, proxy credentials, or response bodies. `RequestEvent.Path` does not include the raw query string, but custom raw HTTP paths can still contain IDs; bucket or normalize those paths before using them as metrics labels.

## Safety Boundary

`RequestEvent` includes traffic class, method, host, path, status code, error kind, retry attempts, cache hit, block detection, and duration.

It does not include raw query strings, headers, bodies, API keys, access tokens, cookies, or proxy passwords.
