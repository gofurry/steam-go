# SteamID 模型

SteamID 是 Steam 生态中使用的身份标识模型。

它并不只是“玩家 ID”。一个 SteamID 可以表示个人用户、Clan、游戏服务器、Chat 对象、匿名账号以及其他 Steam Account Type。

`steam-go` 通过下面的公开包提供本地 SteamID 解析与转换能力：

```go
import "github.com/gofurry/steam-go/steamid"
```

这个包完全在本地工作，不需要 Steam Web API Key，也不会主动发起网络请求。

## 64 位身份模型

SteamID 的 canonical identity 是一个打包后的 64 位整数：

```text
63                 56 55      52 51                 32 31                 0
┌────────────────────┬───────────┬─────────────────────┬────────────────────┐
│ Universe   8 bits  │ Type 4bit │ Instance    20 bits │ AccountID  32 bits │
└────────────────────┴───────────┴─────────────────────┴────────────────────┘
```

从概念上可以理解为：

```text
SteamID
=
Universe
+ AccountType
+ Instance
+ AccountID
```

因此 SteamID 并不是一个单纯的数字用户编号。

它同时描述了这个身份属于哪个 Steam Universe、是什么类型的账号以及使用哪个 Instance。

## 常见表示形式

Steam 生态中经常可以看到几种不同的 ID：

| 表示形式 | 示例 | 常见用途 |
|---|---|---|
| AccountID | `12345` | SteamID 的低 32 位账号编号 |
| SteamID64 | `76561197960278073` | Steam Web API、Community |
| Steam2 | `STEAM_1:1:6172` | Source / GoldSrc 服务器工具 |
| Steam3 | `[U:1:12345]` | 带 AccountType 的现代表示 |

它们并不是四种互不相关的 ID。

对于普通 Public Individual 用户，它们只是同一身份信息的不同表示形式：

```text
AccountID: 12345

        ↓

SteamID64:
76561197960278073

Steam2:
STEAM_1:1:6172

Steam3:
[U:1:12345]
```

## AccountID

`AccountID` 只占 SteamID 的低 32 位。

它本身不包含：

- Universe
- AccountType
- Instance

因此，仅凭 AccountID 无法完整描述一个 Steam 身份。

把 AccountID 转换成完整 SteamID 时，必须额外确定这些组件。

`steam-go` 的通用 `steamid.Parse` 会把单独出现的 AccountID 解释为 Public + Individual + Desktop。

如果希望明确表达这个行为，可以使用：

```go
id, err := steamid.NewIndividual(12345)
```

需要构造其他类型的 SteamID 时，则使用 `steamid.New` 明确提供全部组件。

## SteamID64

SteamID64 是完整 SteamID 打包后的 64 位整数表示，也是 Steam Web API 中最常见的形式。

例如：

```text
76561197960278073
```

Steam Community 的数字个人主页同样使用 SteamID64：

```text
https://steamcommunity.com/profiles/76561197960278073
```

在 `steam-go` 中可以使用：

```go
id, err := steamid.ParseSteamID64("76561197960278073")
```

如果输入的是 Community URL，则应使用 `ParseCommunityURL`。

## Steam2

Steam2 使用 Source 生态中非常常见的格式：

```text
STEAM_X:Y:Z
```

例如：

```text
STEAM_1:1:6172
```

它主要用于 Individual Account，在 Source / GoldSrc 服务器管理、Ban List 和一些旧工具中仍然很常见。

历史上的部分软件会把 Public Universe 表示成：

```text
STEAM_0:...
```

而 Public Universe 的实际编号是 `1`。

因此 `steam-go` 同时接受 `STEAM_0` 和 `STEAM_1`，并在输出 canonical Steam2 时统一使用 `STEAM_1`。

Steam2 无法无损表示所有 Steam Account Type 和 Instance，因此 `steam-go` 不会为了得到一个 Steam2 字符串而静默丢失身份信息。

## Steam3

Steam3 会明确携带 Account Type：

```text
[U:1:12345]
```

常见类型包括：

| 字符 | 含义 |
|---|---|
| `U` | Individual |
| `G` | Game Server |
| `A` | Anonymous Game Server |
| `g` | Clan |
| `T` | Chat |
| `c` | Clan Chat |
| `L` | Lobby |
| `a` | Anonymous User |

Steam3 还可以显式携带 Instance：

```text
[U:1:12345:2]
```

Account Type 是 SteamID 语义中的重要部分。

例如：

```text
[U:1:12345]
```

和：

```text
[g:1:12345]
```

即使 AccountID 相同，也表示完全不同类型的 Steam 身份。

## Community URL 与 Vanity

Steam Community 常见两种个人主页：

```text
/profiles/<SteamID64>
```

以及：

```text
/id/<vanity>
```

数字形式的 `/profiles/` URL 本身已经包含 SteamID64，因此可以完全在本地解析：

```go
id, err := steamid.ParseCommunityURL(
    "https://steamcommunity.com/profiles/76561197960278073",
)
```

而：

```text
https://steamcommunity.com/id/example-user
```

只包含 Vanity Name，并没有直接携带 SteamID。

要得到真实 SteamID 必须请求 Steam。

因此 `steam-go` 有意把两个行为分开：

```text
steamid.ParseCommunityURL
→ 本地识别 URL

client.API.SteamUser.ResolveVanityURL
→ 显式进行网络解析
```

这样 `steamid` 基础包始终保持纯本地能力。

## 使用 steam-go 解析

对于一般输入，可以直接使用：

```go
id, err := steamid.Parse("STEAM_1:1:6172")
if err != nil {
    return err
}

fmt.Println(id.String())
fmt.Println(id.AccountID())
fmt.Println(id.Universe())
fmt.Println(id.AccountType())
fmt.Println(id.Instance())
```

`Parse` 可以识别：

- AccountID
- SteamID64
- Steam2
- Steam3

它不会自动解析 Vanity，也不会获取玩家资料。

## 实用建议

- 调用大部分 Steam Web API 时优先使用 SteamID64。
- 和 Source / GoldSrc 服务器工具交互时，根据需要使用 Steam2。
- 需要保留 Account Type 和 Instance 语义时，Steam3 更直观。
- 不要把 AccountID 当成完整 SteamID 使用。
- 不要认为任何能放进 `uint64` 的数字都是合法 SteamID。
- 不要为了转换格式而静默丢失 Account Type 或 Instance。
- Vanity Name 只是引用，需要显式通过 Steam API 解析。

完整的 `steam-go` API、校验规则、转换行为与错误模型请参阅 [`docs/steamid.md`](https://github.com/gofurry/steam-go/blob/main/docs/steamid.md)。