# Steam VDF 与 `addons/vdf`

Steam 生态中有不少本地文本文件使用 Valve Data Format，通常也叫 VDF 或 KeyValues。

这些文件不是 Steam Web API，也不属于 `client.API.*` 或 `client.Web.*` 的网络请求层。它们更像是 Steam 客户端、Valve 工具、Source 游戏和部分游戏资源使用的本地配置或元数据文件。

`addons/vdf` 是对 `github.com/gofurry/vdf-go` 的轻量桥接。它让已经使用 `steam-go` 的项目，可以通过 addon 入口解析 VDF / KeyValues 文本文件。

## VDF 长什么样

VDF 是一种树形 key-value 文本格式：

```vdf
"AppState"
{
    "appid"      "730"
    "name"       "Counter-Strike 2"
    "installdir" "Counter-Strike Global Offensive"
}
```

常见特点：

- key 和 value 通常用双引号包裹
- `{}` 表示子节点
- value 本质上是字符串，数值或布尔含义由调用方解释
- 同级节点可能有重复 key
- 节点顺序可能有意义
- VDF 不是 JSON

因为重复 key 和顺序都可能有意义，VDF 不适合简单建模成普通 map。

## 常见 Steam 文件

常见文本 VDF / KeyValues 文件包括：

| 文件 | 常见用途 |
|---|---|
| `libraryfolders.vdf` | Steam 库目录信息 |
| `appmanifest_*.acf` | 已安装应用的本地安装清单 |
| `config.vdf` | Steam 客户端配置 |
| `loginusers.vdf` | 本机已知 Steam 用户 |
| `gameinfo.txt` | Source / Valve 游戏或工具配置 |

不是所有 `.cfg` 都是 VDF。很多 Source 或 console `.cfg` 是命令式文件，内容类似 `bind`、`alias` 或 cvar 配置，不属于 `addons/vdf` 的目标范围。

## `addons/vdf` 的定位

`addons/vdf` 刻意保持很薄。

它负责：

- 解析文本 VDF / KeyValues 文件
- 暴露 `vdf-go` 的核心 document / node 模型
- 保留重复 key 和节点顺序
- 复用 `vdf-go` 的读取、查询、编辑和 marshal helper

它不负责：

- 实现 binary VDF
- 解析 `shortcuts.vdf`
- 自动扫描 Steam 安装目录
- 在调用方未显式传入路径时读取用户目录
- 提取账号、token、cookie 或 session
- 提供强类型 `AppManifest` 或 `LibraryFolder` 业务模型

如果未来需要本机 Steam 扫描或强业务模型，应单独设计，而不是混入 `addons/vdf`。

## 导入

如果你只需要 VDF parser，直接使用 `vdf-go` 也可以：

```go
import vdf "github.com/gofurry/vdf-go"
```

如果项目已经依赖 `steam-go`，可以使用 addon 入口：

```go
import "github.com/gofurry/steam-go/addons/vdf"
```

## 解析 app manifest

```go
doc, err := vdf.ParseFile("steamapps/appmanifest_730.acf")
if err != nil {
    panic(err)
}

appid := doc.Path("AppState", "appid").AsString()
name := doc.Path("AppState", "name").AsString()

fmt.Println(appid, name)
```

`appmanifest_*.acf` 里通常可能有 AppID、应用名、安装目录、build ID、安装状态、depot 信息和更新元数据。

`addons/vdf` 只解析 VDF 树，不承诺 typed app manifest 模型。

## 解析 library folders

```go
doc, err := vdf.ParseFile("config/libraryfolders.vdf")
if err != nil {
    panic(err)
}

folders := doc.First("libraryfolders")
if folders == nil {
    return
}

for _, folder := range folders.Children {
    pathNode := folder.First("path")
    if pathNode == nil {
        continue
    }
    fmt.Println(folder.Key, pathNode.AsString())
}
```

真实 Steam 文件可能随客户端版本变化。业务层应容忍字段缺失或新增。

## 重复 key 与查询

`vdf-go` 使用 slice-based document model，因此重复 key 不会被静默覆盖。

常用 helper：

```go
first := node.First("item")
all := node.All("item")
leaf := doc.Path("AppState", "InstalledDepots", "731", "manifest")
```

- `First` 返回第一个匹配节点。
- `All` 返回全部匹配节点。
- `Path` 按嵌套 key 路径查找。

## 写出 VDF

```go
doc := vdf.NewDocument(
    vdf.NewNode("AppState",
        vdf.NewValue("appid", "730"),
        vdf.NewValue("name", "Counter-Strike 2"),
    ),
)

text, err := vdf.MarshalString(doc)
if err != nil {
    panic(err)
}

fmt.Println(text)
```

Marshal 输出稳定且可读，但不会保留原始注释、空行或格式风格。

## 安全说明

对于不可信或可能很大的输入，使用大小限制：

```go
doc, err := vdf.ParseReaderLimit(reader, 1<<20)
```

也可以设置 parser 资源限制：

```go
doc, err := vdf.ParseReaderLimit(reader, 1<<20,
    vdf.WithMaxDepth(128),
    vdf.WithMaxTokenBytes(1<<20),
    vdf.WithMaxNodes(1_000_000),
)
```

`addons/vdf` 不执行 `#include` 或 `#base`，也不会自动读取额外本地文件。已知 directive 默认会被忽略；如果需要保留，可使用 `WithPreserveDirectives(true)`。

## 和 `steam-go` 的关系

`addons/vdf` 不属于网络请求层。

它不使用：

- `Client`
- `client.API.*`
- `client.Web.*`
- traffic policy
- proxy
- retry
- rate limit
- cookie jar

推荐边界：

| 能力 | 位置 |
|---|---|
| Steam Web API 请求 | `client.API.*` |
| Storefront / Community / Market 只读 JSON | `client.Web.*` |
| A2S 查询 | `addons/a2s` |
| VDF / ACF / KeyValues 文本解析 | `addons/vdf` |
| 自动扫描本机 Steam 安装 | 后续单独设计 |

## 实用建议

- 只在调用方显式传入路径时读取本地 VDF 文件。
- 不要默认扫描用户系统目录。
- 不要把 `loginusers.vdf` 当作认证来源。
- 不要把真实用户路径、SteamID、token 或本机 Steam 配置提交到测试数据。
- 强本地 Steam 业务模型优先放在应用层；只有明确需要时再单独设计 addon。

## 参考文献

- [Valve Developer Community: KeyValues](https://developer.valvesoftware.com/wiki/KeyValues)
