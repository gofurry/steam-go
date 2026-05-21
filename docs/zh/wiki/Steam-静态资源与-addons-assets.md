# Steam 静态资源

Steam 的每个 AppID 周围都有很多有价值的公开图片与媒体资源。

这些资源并不是都由某一个官方 Web API 方法统一返回，也不总是使用同一种 host 或路径结构。有些资源可以只根据 AppID 构造出来；有些资源则需要额外的 hash，或者需要从 Storefront metadata 中读取 Steam 返回的真实 URL。

`addons/assets` 的目标就是让 Go 开发者更容易处理这些资源。

它提供了一组轻量工具，用于构造常见的公开 Steam Store 与 Library 静态资源 URL、解析 Storefront 媒体 URL、验证 URL 是否存在、将资源读入内存、下载文件，以及写出资源 manifest。

## 这篇说明覆盖什么

这篇 Wiki 说明：

- 开发者通常需要哪些公开 Steam 静态资源
- 这些资源通常可以从哪些地方请求
- 哪些资源只需要 AppID 就能构造
- 哪些资源需要 hash 或 Storefront metadata
- `steam-go/addons/assets` 如何帮助处理这些资源
- 这个 addon 有哪些刻意不做的事情

这篇说明不是法律或授权声明。使用 Steam-hosted assets 时，请遵守 Steam 条款、Steamworks 规则以及你自己的产品要求。

## Steam 资源家族

Steamworks 大体把图形资源分为几类。

| 资源家族 | 常见用途 | 示例 |
|---|---|---|
| Store assets | 商店页、搜索结果、推荐、促销页面 | header capsule、small capsule、main capsule、截图、页面背景 |
| Library assets | Steam 客户端库展示 | library capsule、library hero、library logo、library header |
| Community and client icons | Steam Community 与 Steam 客户端的小尺寸展示 | App icon JPG、shortcut/client icon ICO 或 PNG |
| Storefront media | Store appdetails 返回的媒体 | 截图、视频缩略图、WebM/MP4/HLS/DASH URL、背景图 |

Steamworks 文档将 Store assets、Library assets、Community / Client icons 分开说明：

- Store graphical assets: https://partner.steamgames.com/doc/store/assets/standard
- Library assets: https://partner.steamgames.com/doc/store/assets/libraryassets
- Community and client icons: https://partner.steamgames.com/doc/store/assets/community
- Graphical assets overview: https://partner.steamgames.com/doc/store/assets

## 常见公开资源 Host

实际开发中，经常能看到这些公开资源 host：

| Host | 常见用途 |
|---|---|
| `shared.steamstatic.com` | Store 与 Library 静态资源，例如 `header.jpg`、`library_600x900_2x.jpg`、`library_hero.jpg`、`logo_2x.png` |
| `cdn.cloudflare.steamstatic.com` | Steam Community 与客户端图片资源，通常是 AppID/hash 路径 |
| `shared.fastly.steamstatic.com` | 可能在 Steam 返回的 URL 或浏览器观察到的资源里出现的 CDN host |

`addons/assets` 使用 `https://shared.steamstatic.com` 作为本地构造 Store 与 Library URL 时的 canonical static asset base。

它也会在下面这个路径下构造 community/client icon URL：

```text
https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/apps/{appid}/{hash}.{ext}
```

不要假设每一个 Steam CDN host 会永远保持同样的路径行为。如果 Steam metadata 已经返回了真实 URL，应优先使用返回值；如果是自己猜出来的 URL，在资源存在性重要时应先验证。

## 只需要 AppID 的 Store 与 Library 资源

很多高价值 Store 与 Library 资源可以只根据 AppID 构造。

以 AppID `107100` 为例，常见资源包括：

```text
https://shared.steamstatic.com/store_item_assets/steam/apps/107100/header.jpg
https://shared.steamstatic.com/store_item_assets/steam/apps/107100/library_600x900_2x.jpg
https://shared.steamstatic.com/store_item_assets/steam/apps/107100/library_hero.jpg
https://shared.steamstatic.com/store_item_assets/steam/apps/107100/logo_2x.png
```

`addons/assets` 将这些资源封装成 typed asset kinds。

| `addons/assets` kind | 文件名 |
|---|---|
| `assets.KindHeader` | `header.jpg` |
| `assets.KindHeaderLocalized` | `header_{language}.jpg` |
| `assets.KindCapsuleSmall` | `capsule_231x87.jpg` |
| `assets.KindCapsuleMain` | `capsule_616x353.jpg` |
| `assets.KindLibraryCapsule` | `library_600x900.jpg` |
| `assets.KindLibraryCapsule2x` | `library_600x900_2x.jpg` |
| `assets.KindLibraryHero` | `library_hero.jpg` |
| `assets.KindLibraryLogo` | `logo.png` |
| `assets.KindLibraryLogo2x` | `logo_2x.png` |

这些 helper 只是本地 URL 构造器。调用它们本身不会请求 Steam。

## 需要 Hash 的 Community 与 Client Icons

Community 与 client icon URL 需要已知 hash。

示例：

```text
https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/apps/107100/8377b4460f19465c261673f76f2656bdb3288273.jpg
https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/apps/107100/ad7f9414231a7a5bb96d74e21893a84972dcbee8.ico
```

这个 hash 不是由 `addons/assets` 根据 AppID 推导出来的。

它需要来自其他来源，例如 Steam owned games metadata、appinfo/client metadata，或者调用方已经信任的其他来源。`addons/assets` 只负责在调用方已经拥有 AppID/hash pair 的情况下构造最终 URL。

```go
icon := assets.CommunityIconURL(
    107100,
    "8377b4460f19465c261673f76f2656bdb3288273",
)

clientIcon := assets.ClientIconURL(
    107100,
    "ad7f9414231a7a5bb96d74e21893a84972dcbee8",
)
```

相关 kind：

| `addons/assets` kind | 含义 |
|---|---|
| `assets.KindCommunityIconJPG` | 已知 hash 时的 community/app icon JPG |
| `assets.KindCommunityLogoJPG` | 已知 hash 时的 community/logo JPG |
| `assets.KindClientIconICO` | 已知 hash 时的 client shortcut icon ICO |

## Storefront 媒体 URL

有些有价值的媒体资源更适合从 Storefront appdetails 中解析，而不是靠静态路径猜测。

`addons/assets` 可以通过现有 `client.Web.Storefront` service 请求 Storefront appdetails，并提取：

- Store 页面背景图
- Store 页面 raw 背景图
- 截图缩略图
- 完整截图
- 视频缩略图
- 视频 WebM URL
- 视频 MP4 URL
- 视频 DASH/HLS playlist URL

这些资源对应 Store media kinds，例如：

```go
assets.KindStoreBackground
assets.KindStoreBackgroundRaw
assets.KindScreenshotThumbnail
assets.KindScreenshotFull
assets.KindMovieThumbnail
assets.KindMovieWebM480
assets.KindMovieWebMMax
assets.KindMovieMP4480
assets.KindMovieMP4Max
assets.KindMovieDASHAV1
assets.KindMovieDASHH264
assets.KindMovieHLSH264
```

DASH/HLS helper 返回的是 playlist 或 manifest URL 本身。该 addon 不会把 playlist 展开成视频分片。

## 基础用法

导入 addon：

```go
import "github.com/gofurry/steam-go/addons/assets"
```

构造简单 URL：

```go
headers := assets.HeaderURLs(550, 107100)
heroes := assets.URLs(assets.KindLibraryHero, 550, 107100)
logos := assets.LibraryLogo2xURLs(550, 107100)
```

按 AppID 返回结构体：

```go
all := assets.AllWithLanguage("schinese", 550, 107100)

for _, item := range all {
    fmt.Println(item.AppID, item.Header, item.LibraryHero, item.LibraryLogo2x)
}
```

构造 typed URL item 列表：

```go
items := assets.ListKindsWithLanguage(
    "schinese",
    []assets.Kind{
        assets.KindHeader,
        assets.KindHeaderLocalized,
        assets.KindLibraryCapsule2x,
        assets.KindLibraryHero,
        assets.KindLibraryLogo2x,
    },
    550,
    107100,
)
```

## 验证 URL 是否存在

不是每个 AppID 都有每一种资源。

当资源存在性很重要时，应先验证：

```go
results, err := assets.VerifyAppAssets(ctx, assets.VerifyAppOptions{
    Kinds: []assets.Kind{
        assets.KindHeader,
        assets.KindLibraryHero,
        assets.KindLibraryLogo2x,
    },
    Language: "schinese",
}, 550, 107100)
if err != nil {
    return err
}

for _, result := range results {
    fmt.Println(result.AppID, result.Kind, result.Exists, result.StatusCode)
}
```

直接验证 URL：

```go
results, err := assets.VerifyURLsWithOptions(ctx, assets.VerifyOptions{
    URLValidator: assets.SteamStaticURLValidator,
}, "https://shared.steamstatic.com/store_item_assets/steam/apps/550/header.jpg")
```

`VerifyURLs` 会先使用 `HEAD`，当服务器返回 `405` 或 `501` 时回退到 `GET`。HTTP 非 2xx 响应会被报告为 `Exists=false`，而不是直接作为硬错误处理。

## 将资源读入内存

小批量读取：

```go
results, err := assets.ReadAppAssets(ctx, assets.ReadAppOptions{
    Kinds: []assets.Kind{assets.KindHeader, assets.KindLibraryHero},
    MaxBytes: 8 << 20,
}, 550)
```

大批量读取时，建议逐个结果处理，避免整批数据同时保留在内存中：

```go
err := assets.ReadEachAppAssets(ctx, assets.ReadAppOptions{
    Kinds:       []assets.Kind{assets.KindHeader},
    Concurrency: 4,
}, func(result assets.ReadResult) error {
    if result.Error != "" {
        fmt.Println("failed:", result.URL, result.Error)
        return nil
    }
    fmt.Println("read", result.URL, result.BytesRead)
    return nil
}, 550, 107100)
```

`ReadURLs` 默认每个资源最多读取 32 MiB。如需读取更大的资源，应显式设置 `MaxBytes`。

## 下载资源

下载根据 AppID 构造的资源：

```go
results, err := assets.DownloadAppAssets(ctx, assets.DownloadAppOptions{
    Dir:         "./tmp/assets",
    Mode:        assets.StoreByAppID,
    Language:    "schinese",
    Kinds:       []assets.Kind{assets.KindHeader, assets.KindLibraryHero, assets.KindLibraryLogo2x},
    SkipExisting: true,
    Concurrency: 4,
}, 550, 107100)
```

下载模式：

| Mode | 行为 |
|---|---|
| `assets.StoreFlat` | 将生成的文件直接写入目标目录，并使用 AppID 前缀，例如 `550_header.jpg` |
| `assets.StoreByAppID` | 将文件写入 AppID 子目录，例如 `550/header.jpg` |

下载结果包括：

- `DownloadStatusDownloaded`
- `DownloadStatusSkipped`
- `DownloadStatusFailed`

批量下载会尽量尝试每一个 URL。即使后续资源失败，已经成功下载的文件也会保留在磁盘上。

## 获取 Storefront 媒体

Storefront media 使用现有的 `client.Web.Storefront` service。

```go
client, err := steam.NewClient(steam.WithSafeDefaults())
if err != nil {
    return err
}
defer client.Close()

items, err := assets.FetchStoreMediaURLs(ctx, client.Web.Storefront, assets.StoreMediaOptions{
    Language: "schinese",
    Kinds: []assets.Kind{
        assets.KindScreenshotFull,
        assets.KindMovieThumbnail,
        assets.KindStoreBackground,
    },
}, 550)
if err != nil {
    return err
}

for _, item := range items {
    fmt.Println(item.AppID, item.Kind, item.ID, item.Name, item.URL)
}
```

也可以验证、读取或下载 Storefront media：

```go
verified, err := assets.VerifyStoreMedia(ctx, client.Web.Storefront, assets.VerifyStoreMediaOptions{
    Language: "schinese",
}, 550)
```

```go
results, err := assets.DownloadStoreMedia(ctx, client.Web.Storefront, assets.DownloadStoreMediaOptions{
    Dir:         "./tmp/store-media",
    Mode:        assets.StoreByAppID,
    Kinds:       []assets.Kind{assets.KindScreenshotFull, assets.KindMovieThumbnail},
    Concurrency: 4,
}, 550)
```

## 写出 Manifest

Manifest 适合 crawler、构建流水线、机器人和前端索引使用。

```go
items := assets.ListWithLanguage("schinese", 550, 107100)
manifest := assets.NewURLManifest(items)

if err := assets.WriteManifestJSON("./tmp/assets/manifest.json", manifest); err != nil {
    return err
}
```

下载结果也可以写成 manifest：

```go
results, err := assets.DownloadAppAssets(ctx, opts, 550, 107100)
if err != nil {
    fmt.Println("some downloads failed:", err)
}

manifest := assets.NewDownloadManifest(results)
_ = assets.WriteManifestJSON("./tmp/assets/download-manifest.json", manifest)
```

## Direct URL 安全

部分 helper 接受调用方传入的直接 URL：

- `VerifyURLs`
- `ReadURLs`
- `DownloadURLs`

如果这些 URL 来自用户或其他不可信来源，应在发送 HTTP 请求前设置 validator。

```go
results, err := assets.ReadURLsWithOptions(ctx, assets.ReadOptions{
    URLValidator: assets.SteamStaticURLValidator,
}, urls...)
```

可用 validator：

```go
assets.AllowHosts("shared.steamstatic.com")
assets.AllowHostSuffixes("steamstatic.com")
assets.SteamStaticURLValidator
```

`SteamStaticURLValidator` 接受 `steamstatic.com` 下的 host。如果你的应用需要更严格的 host 控制，请使用更严格的 validator。

## `addons/assets` 不做什么

`addons/assets` 刻意保持轻量。

它不做：

- 创建 `steam.Client`
- 接入 SteamGridDB
- 解析 Steam client appinfo
- 自己发现 client icon hash
- 保证每一个生成的 URL 都一定存在
- 将 DASH/HLS playlist 展开成视频分片
- 将公开静态资源 URL 当作官方稳定 Web API endpoint

它是一个面向公开 Steam asset resources 的实用工具包，用于构造、验证、读取和下载资源。

## 实用建议

- 需要可预测的图片候选时，可以构造 AppID-only 的 Store 与 Library 静态 URL。
- 在生产展示或下载前，建议先验证生成的 URL。
- 截图、视频和背景图更适合通过 Storefront media helper 从 appdetails 中解析。
- hash-based community/client icon 应和 AppID-only URL 构造分开处理。
- 大批量处理时使用低并发，并优先使用 streaming read helper。
- 保留 manifest，让下游任务知道到底解析了哪些 URL、哪些下载成功。
- 如果 Steam metadata 已经返回了真实 URL，优先使用返回值。
- 将 Steam 静态资源路径视为公开资源路径，而不是官方枚举 API。

## CLI 示例

仓库中包含一个 example 命令：

```bash
go run ./examples/assets -app-ids 550,107100 -language schinese
```

验证生成的 app assets：

```bash
go run ./examples/assets -verify-apps -kind all
```

下载生成的 assets：

```bash
go run ./examples/assets \
  -download-apps \
  -download-dir ./tmp/assets \
  -download-mode by_app_id \
  -kind all \
  -skip-existing \
  -concurrency 4 \
  -manifest ./tmp/assets/manifest.json
```

获取 Storefront media：

```bash
go run ./examples/assets -app-ids 550 -store-media -kind all
```

## 相关 Wiki 页面

- [公开商店页面访问说明](https://github.com/gofurry/steam-go/wiki/%E5%85%AC%E5%BC%80%E5%95%86%E5%BA%97%E9%A1%B5%E9%9D%A2%E8%AE%BF%E9%97%AE%E8%AF%B4%E6%98%8E)
- [流量策略](https://github.com/gofurry/steam-go/wiki/%E6%B5%81%E9%87%8F%E7%AD%96%E7%95%A5)
- [限流策略](https://github.com/gofurry/steam-go/wiki/%E9%99%90%E6%B5%81%E7%AD%96%E7%95%A5)
- [代理与网络策略](https://github.com/gofurry/steam-go/wiki/%E4%BB%A3%E7%90%86%E4%B8%8E%E7%BD%91%E7%BB%9C%E7%AD%96%E7%95%A5)
