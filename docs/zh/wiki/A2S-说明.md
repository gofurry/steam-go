# A2S 说明

`addons/a2s` 用于直接查询游戏服务器。

它和 Steam Web API 调用不是一回事。

## Web API 与 A2S

| Steam Web API | A2S |
|---|---|
| 基于 HTTP | 基于 UDP 的游戏服务器查询 |
| 查询 Steam 平台数据 | 查询某个具体游戏服务器 |
| 需要时使用 SDK client 和凭证 | 不使用 Steam Web API key |
| 适合玩家资料、成就、应用数据 | 适合服务器 info、players、rules |

## 支持的查询类型

A2S addon 暴露常见服务器查询能力：

- info
- players
- rules

## 示例

服务器信息：

```bash
go run ./examples/a2s -server 1.2.3.4:27015 -query info
```

玩家：

```bash
go run ./examples/a2s -server 1.2.3.4:27015 -query players
```

规则：

```bash
go run ./examples/a2s -server 1.2.3.4:27015 -query rules
```

## 相关包

- `addons/a2s`
- `addons/a2s/master`
- `addons/a2s/scanner`

## 实用建议

- 把 A2S 看作对游戏服务器的网络探测。
- 预期会遇到超时和部分失败。
- 扫描时使用有界并发。
- 不要把 A2S 错误处理和 Web API 错误处理混在一起。
- 保持 A2S addon 化，让核心 Web API client 保持轻量。

