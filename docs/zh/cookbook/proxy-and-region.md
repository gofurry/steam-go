# Cookbook：代理与区域网络

当 Steam 访问需要固定代理、代理轮换或按 host/path 路由时，使用 proxy selector。

## 固定代理

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
defer client.Close()
```

## 按 Host 和 Path 路由

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

## 说明

- `ProxyURL: ""` 表示该 route 使用直连。
- session 需要尽量复用同一代理时，使用 `NewStickyProxySelector(...)`。
- 代理池需要失败冷却时，使用 `NewHealthCheckedRoundRobinProxySelector(...)`。
- 不要记录包含用户名或密码的 proxy URL。
