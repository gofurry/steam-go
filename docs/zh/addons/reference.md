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

## `addons/a2s`

`addons/a2s` 是对独立包 [`github.com/GoFurry/a2s-go`](https://github.com/GoFurry/a2s-go) 的轻量桥接。

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
