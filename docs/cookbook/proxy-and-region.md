# Cookbook: Proxy and Region

Use proxy selectors when Steam access needs a fixed proxy, proxy rotation, or host/path routing.

## Fixed Proxy

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

## Route by Host and Path

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

## Notes

- `ProxyURL: ""` means direct connection for that route.
- Use `NewStickyProxySelector(...)` when a session should keep preferring the same proxy.
- Use `NewHealthCheckedRoundRobinProxySelector(...)` when one proxy pool needs temporary cooldown after failures.
- Do not log proxy URLs with usernames or passwords.
