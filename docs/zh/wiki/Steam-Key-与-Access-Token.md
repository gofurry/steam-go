# Steam Key 与 Access Token

本页用于说明在使用 `steam-go` 或 Steam Web API 时常见的几类凭证：Steam Web API Key、Access Token、Publisher Web API Key / Partner Key，以及 OpenID 登录身份。

> Access Token 获取方式依赖 Steam Store / Community 网页登录态，属于网页侧行为，可能变化；不要把它当作和官方 Web API Key 一样的长期稳定凭证。

## 快速结论

| 凭证 | 用途 | 是否长期稳定 | 适合放在哪里 |
|---|---|---|---|
| Steam Web API Key | 调用需要 `key` 的普通 Steam Web API | 通常较长期，直到重置或失效 | 后端配置 / 服务端环境变量 |
| Access Token | Steam Store / Community 登录态下的部分接口认证 | 会过期 | 临时使用，不建议长期保存 |
| Publisher Web API Key / Partner Key | Steamworks 合作伙伴 / 发行商后端接口 | 高敏感，权限更高 | 只放可信服务器 |
| OpenID | 让用户通过 Steam 登录你的网站 | 不是 API key | 登录流程中使用 |

```text
Web API Key   = 调官方 Steam Web API 的 key
Access Token  = Steam 网页登录态下的短期 token
Publisher Key = Steamworks 合作伙伴后端使用的高级 key
OpenID        = 只用于确认“这个用户是谁”
```

## 1. Steam Web API Key 是什么？

Steam Web API Key 是调用 Steam Web API 时最常见的凭证。很多接口会要求传入：

```text
key=YOUR_STEAM_WEB_API_KEY
```

在 `steam-go` 中可以这样配置：

```go
client, err := steam.NewClient(
    steam.WithAPIKey("your-steam-web-api-key"),
)
```

它适合用于查询玩家公开资料、好友列表、游戏成就、拥有游戏，以及调用部分需要普通 Web API key 的接口。

注意：

```text
Steam Web API Key 不是用户登录态。
Steam Web API Key 也不是 Publisher Key。
```

## 2. Steam Web API Key 怎么获取？

普通 Steam Web API Key 通常在这里获取：

```text
https://steamcommunity.com/dev/apikey
```

一般流程：

1. 登录你的 Steam 账号；
2. 打开 `https://steamcommunity.com/dev/apikey`；
3. 注册一个 Web API Key；
4. 填写 domain name；
5. 同意条款；
6. 复制生成的 key。

如果你的项目是后端服务，建议把 key 放在 `.env`、服务器环境变量、密钥管理服务或 CI/CD Secret 中，不要写死在源码里。

```bash
STEAM_WEB_API_KEY=your-key-here
```

```go
key := os.Getenv("STEAM_WEB_API_KEY")

client, err := steam.NewClient(
    steam.WithAPIKey(key),
)
```

## 3. Steam Web API 有哪些认证级别？

从使用者角度，可以把 Steam Web API 认证分成三类：

| 类型 | 说明 |
|---|---|
| Public methods | 返回公开数据，通常不需要认证 |
| User Web API key methods | 需要普通用户 Web API key |
| Protected / publisher methods | 需要 Publisher Web API Key，通常面向可信后端 |

官方 Steamworks Web API 文档说明，Steam Web API 包含 public methods 和 protected methods；protected methods 需要认证，通常应该由可信后端调用。

## 4. API Key 如何传给 Steam？

Steam Web API 常见传参方式是 query 参数：

```text
https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/?key=YOUR_KEY&steamids=7656119...
```

部分资料也提到可以通过 HTTP header 传：

```text
x-webapi-key: YOUR_KEY
```

但对普通使用者来说，最常见、最兼容的方式仍然是 `key` query 参数。在 `steam-go` 中通常不需要手动拼接 `key`，直接使用：

```go
steam.WithAPIKey("your-key")
```

## 5. Access Token 是什么？

Access Token 和 Web API Key 不是一回事。它更接近：

```text
Steam 网页登录态下，Store / Community 页面使用的用户认证 token。
```

它通常用于 Steam 自己的 Store / Community 网页接口。特点：

- 通常和当前登录用户相关；
- 会过期；
- 不是所有 API 都支持；
- 有些 API 只支持 key；
- 有些 API 只支持 access token；
- 一般通过 `access_token` 参数传递。

在 `steam-go` 中可以这样配置：

```go
client, err := steam.NewClient(
    steam.WithAccessToken("your-access-token"),
)
```

也可以使用多个 token：

```go
client, err := steam.NewClient(
    steam.WithAccessTokens("token-a", "token-b"),
)
```

## 6. Access Token 怎么获取？

> Access Token 属于登录态敏感信息。只应该在你自己的浏览器、自己的 Steam 账号、自己的测试环境中查看。不要把 token 粘贴到不可信网站或分享给别人。

### 6.1 获取 Store Token

登录 Steam 网页后，打开：

```text
https://store.steampowered.com/pointssummary/ajaxgetasyncconfig
```

复制返回 JSON 里的：

```text
webapi_token
```

这个 token 通常用于 Store 相关 authenticated web API。

### 6.2 获取 Community Token

登录 Steam Community 后，打开：

```text
https://steamcommunity.com/my/edit/info
```

然后在浏览器 DevTools Console 中执行：

```js
JSON.parse(application_config.dataset.loyalty_webapi_token)
```

也可以手动从页面里的 `application_config` 元素上复制：

```text
data-loyalty_webapi_token
```

这个 token 通常用于 Community 相关 authenticated web API。

## 7. Access Token 的重要特性

Access Token 和普通 API Key 最大的区别是：

```text
Access Token 会过期。
```

你可能会看到类似信息：

```text
Currently entered token is for web:store with steamid 76561198370695025
and expires on May 14, 2026, 14:32.
```

这说明 token 可能绑定用途、SteamID 和过期时间。过期后需要重新获取；泄露后，在过期前可能被滥用。

更适合：

```text
临时测试
手动 smoke test
调试 Store / Community 登录态接口
```

不适合：

```text
长期写进配置文件
提交到 Git 仓库
放到前端代码
分享给第三方工具
```

## 8. Publisher Web API Key / Partner Key 是什么？

Publisher Web API Key 是 Steamworks 合作伙伴、发行商、开发者后台使用的高级 key。

它通常用于需要发行商权限的接口、敏感数据访问、受保护操作和 Steamworks 后端服务器调用。

普通 key：

```text
适合普通 Steam Web API
```

Publisher key：

```text
适合 Steamworks Partner / Publisher 后端接口
```

## 9. Public Host 和 Partner Host

普通 Steam Web API 常见 host：

```text
https://api.steampowered.com
```

部分 public Web API 也可能使用：

```text
https://community.steam-api.com
```

Partner-only host：

```text
https://partner.steam-api.com
```

Partner host 有更严格的要求：

- 只支持 HTTPS；
- 每个请求都需要有效的 publisher Web API key；
- 即使某些接口在 public host 上不需要 key，在 partner host 上也需要 publisher key；
- 使用普通 key 或无效 publisher key 可能返回 `403`；
- 反复产生 `403` 可能导致连接 IP 被严格限流或临时 deny；
- 不应该直接连 IP，应使用 DNS 名称。

```text
不要拿普通 Web API Key 去调用 partner.steam-api.com。
```

## 10. OpenID 和 Key 有什么区别？

Steam 也可以作为 OpenID provider。OpenID 用于让用户通过 Steam 登录你的网站，它的作用是确认：

```text
这个用户确实拥有某个 SteamID64。
```

它不等于 Web API Key、Access Token、Publisher Key，也不等于你的业务系统 session。

典型流程：

```text
1. 用户点击“使用 Steam 登录”
2. 跳转到 Steam OpenID 登录页面
3. Steam 回调你的网站
4. 你验证回调
5. 获得用户的 SteamID64
6. 你自己创建业务 session
```

在 `steam-go` 中，对应能力是：

```text
addons/openid
```

## 11. Key 和 Token 的安全建议

### 不要提交到 Git

不要这样写：

```go
client, _ := steam.NewClient(
    steam.WithAPIKey("ABCDEF123456"),
)
```

推荐：

```go
client, _ := steam.NewClient(
    steam.WithAPIKey(os.Getenv("STEAM_WEB_API_KEY")),
)
```

### 不要放到前端

不要把 key 放在 Vue / React / Nuxt 前端代码、浏览器 localStorage、公开 JS bundle 或 GitHub Pages 静态页面中。

推荐架构：

```text
Browser -> Your Backend -> steam-go -> Steam Web API
```

不推荐：

```text
Browser -> Steam Web API + raw key
```

### 不要打印原始 URL

Steam Web API key 和 access token 经常出现在 query 参数里。

危险日志：

```text
https://api.steampowered.com/xxx?key=SECRET&access_token=TOKEN
```

在 `steam-go` 中应使用：

```go
safeURL := steam.RedactSensitiveURL(rawURL)
```

### 不要把 Access Token 当长期凭证

Access Token 会过期，也和网页登录态强相关。不要长期写入 `.env`、数据库或服务端永久配置。

### Publisher Key 只放可信服务器

Publisher Web API Key 权限更高，必须只保存在可信后端服务器、安全的 CI/CD Secret 或专用密钥管理系统中。不要分发到游戏客户端、桌面客户端、移动端 App、前端页面或公开仓库。

## 12. steam-go 中应该怎么选？

### 普通 Web API 请求

```go
client, err := steam.NewClient(
    steam.WithAPIKey(os.Getenv("STEAM_WEB_API_KEY")),
)
```

适合：

```text
GetPlayerSummaries
GetFriendList
GetOwnedGames
GetPlayerAchievements
GetNewsForApp
```

### 需要 Access Token 的接口

```go
client, err := steam.NewClient(
    steam.WithAccessToken(os.Getenv("STEAM_ACCESS_TOKEN")),
)
```

适合部分 Store / Community 登录态接口，或部分只接受 `access_token` 的接口。

### 多 key 场景

```go
client, err := steam.NewClient(
    steam.WithAPIKeys(
        os.Getenv("STEAM_WEB_API_KEY_A"),
        os.Getenv("STEAM_WEB_API_KEY_B"),
    ),
)
```

适合你合法拥有多个 key，并希望分摊请求。

### 多 key + 失败冷却

```go
client, err := steam.NewClient(
    steam.WithHealthCheckedAPIKeys(
        steam.DefaultAPIKeyHealthConfig(),
        os.Getenv("STEAM_WEB_API_KEY_A"),
        os.Getenv("STEAM_WEB_API_KEY_B"),
    ),
    steam.WithRetry(2),
)
```

适合某个 key 遇到 `401` / `429` 时，临时切换到其他 key。它用于提高可用性，不应该用于绕过 Steam 的限制。

## 13. 常见问题

### Q: Web API Key 会过期吗？

普通 Web API Key 通常更像长期凭证，直到你手动重置、撤销、账号或权限发生变化。但它仍然应该被当成敏感凭证保护。

### Q: Access Token 会过期吗？

会。Access Token 通常包含过期时间，过期后需要重新获取。

### Q: Access Token 可以替代 Web API Key 吗？

不能简单替代。有些 API 支持 access token，有些 API 只支持 Web API key，有些 API 只支持特定类型的凭证。

### Q: Publisher Key 可以给客户端用吗？

不可以。Publisher Web API Key 必须保存在可信服务端，不能分发给游戏客户端、桌面客户端、移动端或前端页面。

### Q: OpenID 登录后就能调用所有 Web API 吗？

不能。OpenID 只证明用户身份，通常只能让你拿到用户的 SteamID64。调用 Web API 仍然需要根据接口要求提供 key、access token 或 publisher key。

### Q: 我应该在 README 里公开自己的 key 吗？

绝对不要。如果 key 泄露，应立即重置或撤销，并检查日志、CI、部署配置中是否仍然残留。

## 14. 推荐实践

| 场景 | 推荐做法 |
|---|---|
| 普通公开数据查询 | 使用普通 Steam Web API Key |
| 后端服务 | key 放在环境变量或 Secret 中 |
| 用户登录 | 使用 Steam OpenID |
| 临时测试 Store / Community 登录态接口 | 使用 Access Token，但不要长期保存 |
| 发行商后端接口 | 使用 Publisher Web API Key |
| 生产日志 | 使用 `steam.RedactSensitiveURL(...)` 脱敏 |
| 高并发请求 | 配合 `WithSafeDefaults()`、限流、重试 |
| 多 key 使用 | 使用 health-checked key provider，但不要绕过限制 |

## 15. 参考链接

- Steam Web API Key 获取入口: `https://steamcommunity.com/dev/apikey`
- Steam Web API Terms of Use: `https://steamcommunity.com/dev/apiterms`
- Steamworks Web API Overview: `https://partner.steamgames.com/doc/webapi_overview`
- Steam Web API Explorer / xPaw: `https://steamapi.xpaw.me/`
