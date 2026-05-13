# Partner API 说明

Steam 提供了面向发行商安全服务器场景的 partner-only Web API host。

本页主要用于 `steam-go` 未来设计参考，以及高级用户理解边界。

## Public Host 与 Partner Host

| Host | 使用场景 |
|---|---|
| `api.steampowered.com` | 公开 Steam Web API |
| `partner.steam-api.com` | Partner-only 安全服务端请求 |

## Partner Host 特性

Partner host 有不同的运行规则：

- 只支持 HTTPS。
- 面向安全的 publisher server。
- 每个请求都需要有效的 publisher Web API key。
- 没有有效 publisher key 的请求会返回 `403`。
- 反复产生 `403` 的请求可能触发连接 IP 的严格限流。
- 它不是普通 public Web API key 的使用流程。

## IP 白名单

Steam 支持为 Web API key 配置 IP 白名单。

一旦配置了白名单，非白名单地址的请求会被 `403` 拒绝。

这是额外安全层，不是 key 保护的替代品。

## SDK 设计影响

未来面向 partner 的功能可以考虑：

- 独立 base URL 支持
- 显式 publisher key 处理
- 对 `403` 更严格的错误处理
- 更安全的日志与脱敏
- 明确仅服务端使用的配置示例

## 实用建议

- 不要把 publisher key 暴露给客户端。
- 优先只在后端使用。
- 测试时注意不要使用错误 key 类型。
- 反复 `403` 应视为配置错误，不应该激进重试。

## 参考链接

- [Steam Web API Overview](https://partner.steamgames.com/doc/webapi_overview?language=english)
- [Steam Web API Terms of Use](https://steamcommunity.com/dev/apiterms)
