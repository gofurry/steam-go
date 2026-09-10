# SteamID 解析与转换

导入 `github.com/gofurry/steam-go/steamid`，即可在本地解析和转换 Steam 身份标识。该包仅使用 Go 标准库，不需要凭据，也不发起网络请求。`Valid()` 检查 ID 的结构，不验证账号是否真实存在。

[English](../steamid.md) · [离线示例](../../examples/steamid/main.go)

## 快速开始

```go
import (
    "fmt"

    "github.com/gofurry/steam-go/steamid"
)

id, err := steamid.Parse("12345")
if err != nil {
    return err
}
fmt.Println(id.String())    // 76561197960278073
fmt.Println(id.AccountID()) // 12345

steam2, err := id.Steam2()
if err != nil {
    return err
}
steam3, err := id.Steam3()
if err != nil {
    return err
}
fmt.Println(steam2) // STEAM_1:1:6172
fmt.Println(steam3) // [U:1:12345]
```

完整可运行示例：`go run ./examples/steamid`。

## 公开 API

| API | 含义 |
|---|---|
| `type ID uint64` | canonical packed SteamID64 值；零值无效。 |
| `type Universe uint8`、`type AccountType uint8` | Steamworks 组件的 typed enum。 |
| `Parse(string) (ID, error)` | 解析 AccountID、SteamID64、Steam2、Steam3；不接受 URL。 |
| `ParseSteamID64(string) (ID, error)` | 解析十进制 packed SteamID64 并检查结构。 |
| `ParseSteam2(string) (ID, error)` | 解析 `STEAM_X:Y:Z`。 |
| `ParseSteam3(string) (ID, error)` | 解析 `[type:universe:account[:instance]]`。 |
| `ParseCommunityURL(string) (ID, error)` | 解析数字形式的 Steam Community Individual 个人页 URL。 |
| `New(uint32, Universe, AccountType, uint32) (ID, error)` | 按 AccountID、Universe、AccountType、Instance 构造 ID。 |
| `NewIndividual(uint32) (ID, error)` | 为 AccountID 补齐 Public、Individual、Desktop。 |
| `ID.Uint64() uint64`、`ID.String() string` | 原始 packed 值及 canonical 十进制文本。 |
| `ID.AccountID() uint32`、`ID.Instance() uint32` | 低 32 位 AccountID 与 20 位 Instance。 |
| `ID.Universe() Universe`、`ID.AccountType() AccountType` | 8 位 Universe 与 4 位 AccountType。 |
| `ID.Valid() bool` | 独立检查结构，也适用于直接强制转换得到的 ID。 |
| `ID.Steam2() (string, error)`、`ID.Steam3() (string, error)` | 目标格式能够表达时进行无损转换。 |

位布局为：AccountID 占 0–31 位，Instance 占 32–51 位，AccountType 占 52–55 位，Universe 占 56–63 位。`New` 显式拒绝大于 `0xFFFFF` 的 Instance，不通过掩码静默截断。

Universe 常量为 `UniverseInvalid=0`、`UniversePublic=1`、`UniverseBeta=2`、`UniverseInternal=3`、`UniverseDev=4`。AccountType 常量使用 `AccountType` 前缀：

| 后缀 | 数值 | Steam3 字符 |
|---|---:|---|
| Invalid | 0 | 不支持 |
| Individual | 1 | `U` |
| Multiseat | 2 | `M` |
| GameServer | 3 | `G` |
| AnonGameServer | 4 | `A` |
| Pending | 5 | `P` |
| ContentServer | 6 | `C` |
| Clan | 7 | `g` |
| Chat | 8 | `T`、`c`（Clan Chat）、`L`（Lobby Chat） |
| ConsoleUser | 9 | 不支持，不自行发明字符 |
| AnonUser | 10 | `a` |

常用 Instance 为 `InstanceAll=0`、`InstanceDesktop=1`、`InstanceConsole=2`、`InstanceWeb=4`，类型均为 `uint32`。

## 解析与规范化

所有 parser 都先移除输入两端空白。十进制文本只接受 ASCII 数字，不接受正负号、十六进制、浮点数或指数形式。允许前导零；`String()` 输出 canonical 十进制文本。

`Parse` 将 1 至 `4294967295` 的十进制值视为 AccountID，补齐 Public + Individual + Desktop；更大的数值按 packed SteamID64 解析。因此 `Parse("12345")` 成功，而 `ParseSteamID64("12345")` 返回 `ErrInvalidID`。Individual AccountID 不能为 0。

Steam2 前缀不区分大小写。`STEAM_0` 与 `STEAM_1` 都解析为 Public，输出统一为 `STEAM_1`。奇偶位 `Y` 必须是单个 `0` 或 `1`，`Z*2+Y` 不得溢出 uint32。Universe 2–4 保留原义。解析后的类型和实例固定为 Individual + Desktop；其他类型或实例转换到 Steam2 时返回 `ErrUnsupportedConversion`，不会丢弃组件。

Steam3 字符区分大小写：`g` 是 Clan，`G` 是 GameServer。支持 `[U:1:12345:2]`，也兼容旧形式 `[U:1:12345(2)]`；输出统一使用冒号。未指定 Instance 时，`g/T/c/L` 的 base instance 为 0，其他支持的字符为 1。Steam3 的 Universe 0 无效，只有 Steam2 支持历史 universe-0 别名。

`c` 和 `L` 分别把 Instance 与 Clan flag `0x80000`、Lobby flag `0x40000` 做按位 OR，并保留其他位。输出只有在 Instance 等于该字符的默认值（含 flag）时才省略第四段；Multiseat 和 AnonGameServer 始终输出 Instance。额外位使用完整 Instance，例如 `[c:1:12345:524291]`；同时存在两种 flag 时，使用 `c` 加完整 Instance。对可表达的合法 ID，Steam2/Steam3 格式化后重新解析必须得到同一 ID。

## 结构验证与错误

`Valid()` 接受 Universe 1–4、AccountType 1–10。Individual 要求 AccountID 非零、Instance 在 0–4 之间（包括 3）；Clan 要求 AccountID 非零、Instance 为 0；GameServer 要求 AccountID 非零。其他已知类型仅检查共同结构约束，因此部分非 Individual ID 的 AccountID 可以为 0。

所有成功的 parser 和 constructor 都返回合法 ID。直接 `ID(raw)` 转换可能产生无效值，其 `String()` 和组件访问仍可使用；对无效 ID 调用 `Steam2()` 或 `Steam3()` 返回 `ErrInvalidID`。

使用 `errors.Is` 判断错误类别，不依赖错误正文：

| Sentinel | 含义 |
|---|---|
| `ErrInvalidFormat` | 文本格式错误、未知 Steam3 字符，或不支持的 URL 结构/主机。 |
| `ErrInvalidID` | 数值越界或组件组合无效。 |
| `ErrUnsupportedConversion` | ID 有效，但目标格式无法无损表达。 |
| `ErrVanityReference` | Vanity URL 需要调用者显式发起网络解析。 |

失败时返回零 ID（格式化则返回空字符串）。错误正文不会回显输入 URL 或其中的凭据。

## Community URL 与 Vanity 边界

`ParseCommunityURL` 仅接受 `http`/`https`，Hostname 必须精确匹配 `steamcommunity.com` 或 `www.steamcommunity.com`。`/profiles/<SteamID64>` 必须对应结构有效的 Individual。允许一个尾部 `/`、query 和 fragment；拒绝仿冒相似域名、userinfo、多余路径段、群组/Workshop URL 及 `steam://` URI。

`/id/<vanity>` 返回 `ErrVanityReference`，本地包不会自动访问 Steam。需要网络解析时，将 vanity **token**（如 `example-user`）传给现有 `client.API.SteamUser.ResolveVanityURL(ctx, token, nil)`，先处理 API 返回状态，再解析其返回的 SteamID64。参见 [API 参考](api/reference.md)。

公开 `steamid` 包由调用者选择使用。既有 SDK endpoint 和 `internal/steamid.ValidateSteamID64` 继续保持宽松的 decimal uint64 request validation，不会自动套用新包的严格身份语义。本轮不包含个人资料抓取、friend/invite code、认证或 PICS 能力。

## 语义参考

- [Steamworks CSteamID、EUniverse、EAccountType](https://partner.steamgames.com/doc/api/steam_api)：组件枚举，包括 `ConsoleUser=9`。
- [SteamKit SteamID](https://github.com/SteamRE/SteamKit/blob/master/SteamKit2/SteamKit2/Types/SteamID.cs)：结构验证及 Steam3 chat flags。本包额外保证格式化时保留非默认 Instance，并使用 `STEAM_1` 作为 Public Steam2 canonical 输出。
