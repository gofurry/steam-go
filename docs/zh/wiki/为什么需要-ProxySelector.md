# 为什么需要 ProxySelector

`ProxySelector` 存在的原因是：真实 Steam 集成场景里，一个固定代理 URL 往往不够用。

## 简单方案

基础客户端可能只支持：

```go
http.ProxyURL(proxyURL)
```

这适合小场景，但无法表达：

- 代理轮换
- 健康检查
- 失败后 cooldown
- session 亲和
- 按 host/path 路由
- 直连 fallback
- 按流量类型配置代理策略

## steam-go 的方案

`steam-go` 使用：

```go
type ProxySelector interface {
    Next(req *http.Request) (*url.URL, error)
}
```

selector 能拿到 request，因此可以根据这些信息决策：

- request host
- request path
- request context
- session key
- traffic class
- proxy health state

## 它支持了什么

| 功能 | 为什么需要 Selector |
|---|---|
| Round-robin | 需要内部状态 |
| Health check | 需要记住失败情况 |
| Sticky proxy | 需要 session key |
| Routing proxy | 需要 request host/path |
| Direct fallback | 需要路由级决策 |
| Metrics | 需要记录选择和结果 |

## 示例

```go
base, _ := steam.NewRoundRobinProxySelector(
    "http://127.0.0.1:7897",
    "http://127.0.0.1:7898",
)

selector := steam.NewStickyProxySelector(base)

client, err := steam.NewClient(
    steam.WithProxySelector(selector),
)
```

## 设计原则

代理行为是请求策略，不只是静态配置。

