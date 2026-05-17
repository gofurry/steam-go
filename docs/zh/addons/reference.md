# Addon 参考

本文档记录 `steam-go` 当前 addon 的定位和使用边界。

## `addons/openid`

当你需要基于浏览器的 Steam 登录时，可以使用 `addons/openid`。

它负责：

- 构造 Steam OpenID 登录 URL
- 使用 Steam `check_authentication` 验证回调
- 返回 `SteamID64`、`ClaimedID` 和回调中的 `state`

它不负责：

- 替代 Web API 凭证
- 自动拉取玩家资料
- 管理业务登录 session

示例：

```bash
go run ./examples/openid
go run ./examples/openid --proxy http://127.0.0.1:7897
```

在部分网络环境下，浏览器可以打开 Steam 登录页，但服务端验证请求仍可能需要代理。

## `addons/websession`

当你想基于 `client.API.AuthenticationService` 走一条手动的 Steam 网页登录链路时，可以使用 `addons/websession`。

它负责：

- 启动账号密码认证会话
- 可选地提交一次 Steam Guard 验证码
- 轮询直到拿到 Steam token
- 把 refresh token 换成 Store / Community Web Cookie
- 校验 Store 和 Community 两边的 session

它不负责：

- 替你持久化密码、refresh token 或 Cookie
- 读取浏览器 Cookie 或 Steam 客户端本地登录态
- 在示例输出里直接打印敏感 token
- 替你长期托管 refresh token 或登录生命周期

构造方式：

- `websession.NewClientFromSteamClient(...)`：推荐方式，复用根 SDK 的按类别 traffic-policy 执行链
- `websession.NewClient(...)`：手动模式，适合你自己提供 `http.Client`

示例支持 `-account`、`-proxy`，并会在环境变量缺失时通过隐藏输入读取高敏感 secret：

- `STEAM_ACCOUNT_NAME`
- `STEAM_PASSWORD`
- `STEAM_GUARD_CODE`

示例：

```bash
go run ./examples/websession
```

## `addons/freeclaim`

当你想做限免搜索、免费 package 解析，或者显式领取一个免费 license 时，可以使用 `addons/freeclaim`。

它负责：

- 搜索当前 Store 限免候选
- 复用 `client.Web.Storefront.GetAppDetails` 解析免费 package
- 通过 `dynamicstore/userdata` 校验是否已拥有
- 只有在你显式要求时才发送一次领取请求

它不负责：

- 管理账号密码或浏览器 Cookie
- 读取 Steam 客户端本地 token 或任何本地账号数据库
- 自动全部领取
- 无限重试
- 悄悄扩展成批量自动领取流程

构造方式：

- `freeclaim.NewClientFromSteamClient(...)`：推荐方式，复用根 SDK 的按类别 traffic-policy 执行链
- `freeclaim.NewClient(...)`：手动模式，适合你自己提供 `http.Client`

示例默认只读。只有在显式 claim 模式下，才需要通过 `STEAM_REFRESH_TOKEN` 或一次隐藏输入提供 refresh token。

只读搜索 / 解析示例：

```bash
go run ./examples/freeclaim
```

显式领取示例：

```bash
go run ./examples/freeclaim -app-id 480 -package-id 12345 -claim
```

## `addons/a2s`

`addons/a2s` 是对独立包 [`github.com/gofurry/a2s-go`](https://github.com/gofurry/a2s-go) 的轻量桥接。

它适合直接查询游戏服务器：

- `QueryInfo`
- `QueryPlayers`
- `QueryRules`

示例：

```bash
go run ./examples/a2s -server 1.2.3.4:27015 -query info
go run ./examples/a2s -server 1.2.3.4:27015 -query players
go run ./examples/a2s -server 1.2.3.4:27015 -query rules
```

## `addons/a2s/master`

用于 A2S master server discovery。

主要能力：

- 单页 discovery
- 流式 discovery

## `addons/a2s/scanner`

用于批量探测 discovery 结果。

主要能力：

- 批量 probing
- 消费 discovery stream
- 批量执行 `info`、`players`、`rules` 查询

当 `a2s-go` 发布新的稳定版本时，`steam-go` 应同步桥接版本和示例，而不是在本仓库重新实现 A2S 协议。
