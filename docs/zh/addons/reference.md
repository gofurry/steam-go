# Addon 参考

本文档记录 `steam-go` 当前 addon 的定位和使用边界。

跨 addon 的安全边界见：[Addon 安全边界](safety.md)。

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

## `addons/assets`

当你需要根据一个或多个 Steam AppID 构造高价值公开 Store / Library 图片资源 URL 时，可以使用 `addons/assets`。

它负责：

- 本地构造 `header.jpg`、`capsule_616x353.jpg`、`library_600x900_2x.jpg`、`library_hero.jpg`、`logo_2x.png` 等静态 URL
- 为每类资源提供一个直接 helper，例如 `assets.HeaderURLs(...)`、`assets.LibraryLogoURLs(...)`
- 提供 `assets.StoreKinds()`、`assets.LibraryKinds()`、`assets.AllKinds()` 等预设 kind 组
- 提供 `assets.ListWithLanguage(...)` 等资源清单 helper
- 在资源类型运行时才确定时，通过 `assets.URLs(kind, appIDs...)` 统一构造
- 通过 `assets.All(...)` / `assets.AllWithLanguage(...)` 返回每个 AppID 一组完整结构体
- 通过 `assets.VerifyURLs(...)` 和 `assets.VerifyAppAssets(...)` 验证 URL 是否存在
- 通过 `assets.ReadURLs(...)` 和 `assets.ReadAppAssets(...)` 将资源读取到内存
- 通过 `assets.DownloadURLs(...)` 下载任意 URL
- 通过 `assets.DownloadAppAssets(...)` 下载根据 AppID 构造出的资源
- 通过 `assets.FetchStoreMediaURLs(...)` 请求商店页截图、视频和背景资源 URL
- 通过 `assets.VerifyStoreMedia(...)` / `assets.ReadStoreMedia(...)` / `assets.DownloadStoreMedia(...)` 验证、读取或下载这些商店媒体资源
- 通过 `assets.WriteManifestJSON(...)` 写出 URL / 下载 manifest

它不负责：

- 创建 client
- 静态 URL 构造不会请求 Steam；商店媒体发现、验证和下载 helper 会显式发起网络请求
- 接入 SteamGridDB

示例：

```go
headers := assets.HeaderURLs(550, 107100)
heroes := assets.URLs(assets.KindLibraryHero, 550, 107100)
all := assets.AllWithLanguage("schinese", 550, 107100)
```

导出 helper：

- `HeaderURLs`、`HeaderLocalizedURLs`
- `CapsuleSmallURLs`、`CapsuleMainURLs`
- `LibraryCapsuleURLs`、`LibraryCapsule2xURLs`、`LibraryHeroURLs`、`LibraryLogoURLs`、`LibraryLogo2xURLs`
- `StoreKinds`、`StoreKindsWithLocalized`、`LibraryKinds`、`DefaultKinds`、`AllKinds`
- `StoreMediaKinds`、`StoreBackgroundKinds`、`StoreScreenshotKinds`、`StoreMovieKinds`
- `List`、`ListWithLanguage`、`ListKinds`、`ListKindsWithLanguage`
- `URLs(kind, appIDs...)`、`LocalizedURLs(kind, language, appIDs...)`
- `All(appIDs...)`、`AllWithLanguage(language, appIDs...)`
- `CommunityIconURL`、`CommunityIconURLs`、`CommunityLogoURL`、`CommunityLogoURLs`、`ClientIconURL`、`ClientIconURLs`：用于已知 AppID/hash 的图标 URL
- `VerifyURLs(ctx, urls...)`、`VerifyURLsWithClient(ctx, httpClient, urls...)`、`VerifyURLsWithOptions(ctx, VerifyOptions{...}, urls...)`、`VerifyAppAssets(ctx, opts, appIDs...)`
- `ReadURLs(ctx, urls...)`、`ReadURLsWithClient(ctx, httpClient, urls...)`、`ReadURLsWithOptions(ctx, ReadOptions{...}, urls...)`
- `ReadEachURLs(ctx, ReadOptions{...}, handler, urls...)`、`ReadAppAssets(ctx, ReadAppOptions{...}, appIDs...)`、`ReadEachAppAssets(ctx, ReadAppOptions{...}, handler, appIDs...)`
- `DownloadURLs(ctx, dir, urls...)`、`DownloadURLsWithClient(ctx, httpClient, dir, urls...)`
- `DownloadAppAssets(ctx, DownloadAppOptions{...}, appIDs...)`
- `FetchStoreMediaURLs(ctx, storefront, StoreMediaOptions{...}, appIDs...)`
- `VerifyStoreMedia(ctx, storefront, VerifyStoreMediaOptions{...}, appIDs...)`
- `ReadStoreMedia(ctx, storefront, ReadStoreMediaOptions{...}, appIDs...)`、`ReadEachStoreMedia(ctx, storefront, ReadStoreMediaOptions{...}, handler, appIDs...)`
- `DownloadStoreMedia(ctx, storefront, DownloadStoreMediaOptions{...}, appIDs...)`
- `AllowHosts`、`AllowHostSuffixes`、`SteamStaticURLValidator`：用于直接 URL 校验
- `NewURLManifest`、`NewDownloadManifest`、`MarshalManifestJSON`、`WriteManifestJSON`

下载存放方式：

- `assets.StoreFlat`：所有生成资源直接放到目标目录，文件名带 AppID 前缀，例如 `550_header.jpg`
- `assets.StoreByAppID`：按 AppID 建子目录，例如 `550/header.jpg`

批量下载会尽量处理所有 URL。成功的文件会保留在磁盘上，失败项会同时出现在对应的 `DownloadResult.Error` 字段和最终聚合 error 中；直接 URL 下载遇到重复文件名时会自动追加后缀避免互相覆盖。

下载选项包含 `FilenameStyle`、`Overwrite`、`SkipExisting` 和 `Concurrency`。下载结果会标记为 `DownloadStatusDownloaded`、`DownloadStatusSkipped` 或 `DownloadStatusFailed`。

读取 helper 会把完整资源放到 `ReadResult.Data []byte`，适合调用方自行处理。默认每个资源最多读取 32 MiB；需要更大文件时显式设置 `MaxBytes`。大批量读取时优先用 `ReadEachURLs`、`ReadEachAppAssets` 或 `ReadEachStoreMedia`，可以逐个处理结果，不需要把整批数据都留在内存里。

直接 URL helper 会请求调用方传入的 HTTP(S) 地址。如果这些 URL 来自用户输入或其他不可信来源，请设置 `URLValidator`，例如 `assets.SteamStaticURLValidator` 或 `assets.AllowHosts(...)`，再进行验证、读取或下载。

对于商店视频资源，DASH/HLS 类型会保存 `.mpd` / `.m3u8` 播放清单 URL 本身；addon 不会展开下载视频分片。

运行示例：

```bash
go run ./examples/assets -app-ids 550,107100 -language schinese
go run ./examples/assets -verify-urls https://shared.steamstatic.com/store_item_assets/steam/apps/550/header.jpg
go run ./examples/assets -verify-apps -kind all
go run ./examples/assets -read-apps -kind header -proxy http://127.0.0.1:7897
go run ./examples/assets -download-apps -download-dir ./tmp/assets -download-mode by_app_id -kind all -skip-existing -concurrency 4 -manifest ./tmp/assets/manifest.json
go run ./examples/assets -app-ids 550 -store-media -kind all
go run ./examples/assets -app-ids 550 -read-store-media -kind movie_dash_h264 -proxy http://127.0.0.1:7897
go run ./examples/assets -app-ids 550 -download-store-media -download-dir ./tmp/assets-media -download-mode by_app_id -kind movie_dash_h264
go run ./examples/assets -app-ids 550 -store-media -kind all -proxy http://127.0.0.1:7897
```

## `addons/markup`

当你需要在存储、索引或渲染前转换并清洗 Steam BBCode / HTML 内容时，可以使用 `addons/markup`。

它负责：

- 转换常见 Steam BBCode 标签，例如 `[b]`、`[i]`、`[url]`、`[img]`、`[list]`、`[olist]`、`[video]`、`[youtube]`
- 将 `{STEAM_CLAN_IMAGE}` 替换为 Steam clan image CDN 前缀
- 使用安全默认策略清洗生成的 HTML 或已有 HTML
- 提供纯文本和摘要 helper，适合搜索索引和 metadata

它不负责：

- 请求 Steam 内容
- 替业务决定哪些清洗后的标签应该展示
- 保留不安全脚本、事件属性或 JavaScript URL

示例：

```go
html, err := markup.CleanSteamContent(`[b]Patch[/b] {STEAM_CLAN_IMAGE}/abc.png`)
text, err := markup.Summary(html, 120)
```

可运行示例：

```bash
go run ./examples/markup
```

## `addons/vdf`

当基于 `steam-go` 的工具还需要解析 Valve Data Format（VDF / KeyValues）文本文件时，可以使用 `addons/vdf`。

它负责：

- 桥接 `github.com/gofurry/vdf-go`
- 解析文本 VDF / KeyValues 文件
- 支持 `libraryfolders.vdf`、`appmanifest_*.acf` 等常见 Steam 文本文件
- 通过上游文档模型保留重复 key 和节点顺序
- 提供读取、查询、marshal 和小幅编辑 helper

它不负责：

- 实现 binary VDF
- 解析 `shortcuts.vdf`
- 自动扫描 Steam 安装目录
- 在调用方未显式传入路径时读取用户目录
- 提取账号、token、cookie 或 session

示例：

```go
doc, err := vdf.ParseFile("steamapps/appmanifest_730.acf")
appid := doc.Path("AppState", "appid").AsString()
```

可运行示例：

```bash
go run ./examples/vdf -file "C:\\Program Files (x86)\\Steam\\steamapps\\appmanifest_730.acf" -key AppState
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
