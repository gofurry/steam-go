# Steam Web API 特性说明

本页记录 Steam Web API 在真实应用中的一些实用特性。

它不是官方 Steamworks 文档的完整复制，而是摘出对 SDK 设计和生产接入更重要的行为。

## Public 与 Protected API

Steam Web API 同时包含 public methods 和 protected methods。

Public methods 可以由能够发起 HTTP 请求的应用调用。Protected methods 需要认证，更适合可信后端应用调用。

## Public API Host

公开 Steam Web API 通常访问：

```text
api.steampowered.com
```

常见请求形式：

```text
https://api.steampowered.com/<interface>/<method>/v<version>/
```

## 请求格式

Steam Web API 方法通常接收 GET 或 POST 参数。

对于 POST 请求，官方文档要求使用 form URL encoded body，并使用 UTF-8 文本。

## Service 接口与 input_json

部分 Steam Web API 接口属于 service interface。

如果接口名以 `Service` 结尾，例如 `IPlayerService`，它可能支持通过 `input_json` 传递一个 JSON blob 作为参数。

请求形式示例：

```text
?key=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX&input_json={...}
```

其中 JSON 值需要 URL encode。

## SteamID64

Steam Web API 使用 64 位 Steam ID 标识用户。

如果是浏览器登录场景，应该通过 Steam OpenID 验证获取用户身份，不应该直接收集用户的 Steam 账号密码。

## SDK 实用建议

- typed API 方法应该保持清晰和稳定。
- 对大型或不稳定 payload 保留 raw response 方法。
- 凭证注入应优先放在后端。
- 公开 Store 页面和官方 Web API JSON 端点应视为不同类型流量。
- 不要假设没有官方说明的每秒固定限流值。
- 生产环境应使用保守限流与缓存。

## 参考链接

- [Steam Web API Overview](https://partner.steamgames.com/doc/webapi_overview?language=english)
- [Steam Web API Terms of Use](https://steamcommunity.com/dev/apiterms)
