# Steam Static Assets

Steam has many useful public image and media resources around each AppID.

They are not all exposed through one official Web API method, and they are not all stored under the same host or path shape. Some assets can be built from an AppID, while others require a hash or a URL returned by Storefront metadata.

`addons/assets` exists to make this easier for Go developers.

It provides small helpers for building common public Steam Store and Library static asset URLs, resolving Storefront media URLs, verifying whether URLs exist, reading resources into memory, downloading files, and writing asset manifests.

## What This Page Covers

This page explains:

- what kinds of public Steam static resources developers usually need
- where those resources are commonly requested from
- what can be built from only an AppID
- what requires a hash or Storefront metadata
- how `steam-go/addons/assets` helps with these resources
- what the addon intentionally does not do

This page is not a legal or licensing statement. Always follow Steam's terms, Steamworks rules, and your own product requirements when using Steam-hosted assets.

## Steam Asset Families

Steamworks groups graphical assets into several broad families.

| Family | Common use | Examples |
|---|---|---|
| Store assets | Store pages, search results, recommendations, sales surfaces | header capsule, small capsule, main capsule, screenshots, page background |
| Library assets | Steam client library presentation | library capsule, library hero, library logo, library header |
| Community and client icons | Steam Community and Steam client compact presentation | App icon JPG, shortcut/client icon ICO or PNG |
| Storefront media | Media returned by Store appdetails | screenshots, movie thumbnails, WebM/MP4/HLS/DASH URLs, backgrounds |

Steamworks documentation describes Store assets, Library assets, and Community / Client icons separately:

- Store graphical assets: https://partner.steamgames.com/doc/store/assets/standard
- Library assets: https://partner.steamgames.com/doc/store/assets/libraryassets
- Community and client icons: https://partner.steamgames.com/doc/store/assets/community
- Graphical assets overview: https://partner.steamgames.com/doc/store/assets

## Common Public Hosts

In practice, developers usually see these public resource hosts:

| Host | Typical use |
|---|---|
| `shared.steamstatic.com` | Store and Library static assets such as `header.jpg`, `library_600x900_2x.jpg`, `library_hero.jpg`, and `logo_2x.png` |
| `cdn.cloudflare.steamstatic.com` | Steam Community and client image resources, often with AppID/hash-based paths |
| `shared.fastly.steamstatic.com` | CDN host that may appear in returned URLs or browser-observed URLs |

`addons/assets` uses `https://shared.steamstatic.com` as the canonical static asset base for locally constructed Store and Library URLs.

It also builds community/client icon URLs under:

```text
https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/apps/{appid}/{hash}.{ext}
```

Do not assume every Steam CDN host follows the same path behavior forever. Prefer URLs returned by Steam metadata when available, and verify guessed URLs when existence matters.

## AppID-Only Store and Library Assets

Many high-value Store and Library assets can be built from only an AppID.

For AppID `107100`, common examples include:

```text
https://shared.steamstatic.com/store_item_assets/steam/apps/107100/header.jpg
https://shared.steamstatic.com/store_item_assets/steam/apps/107100/library_600x900_2x.jpg
https://shared.steamstatic.com/store_item_assets/steam/apps/107100/library_hero.jpg
https://shared.steamstatic.com/store_item_assets/steam/apps/107100/logo_2x.png
```

`addons/assets` exposes these as typed asset kinds.

| `addons/assets` kind | File name |
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

These helpers are local URL builders. They do not call Steam by themselves.

## Official Store Item Assets

Newer Steam games may use hashed Store item asset paths that cannot be derived
from the AppID alone:

```text
https://shared.steamstatic.com/store_item_assets/steam/apps/4710650/448851b668e4397d9863e571cf481b0e46e1315f/library_hero_2x.jpg
```

Use `FetchStoreItemAssetURLs` when you need those official hashed URLs. It uses
`client.API.StoreBrowseService.GetItems` with `data_request.include_assets=true`
and resolves Steam's returned `asset_url_format`.

```go
client, err := steam.NewClient(steam.WithSafeDefaults())
if err != nil {
    return err
}
defer client.Close()

items, err := assets.FetchStoreItemAssetURLs(ctx, client.API.StoreBrowseService, assets.StoreItemAssetOptions{
    CountryCode: "US",
    Language:    "english",
    Kinds: []assets.Kind{
        assets.KindHeader2x,
        assets.KindLibraryCapsule2x,
        assets.KindLibraryHero2x,
    },
}, 4710650)
if err != nil {
    return err
}

for _, item := range items {
    fmt.Println(item.AppID, item.Kind, item.URL, item.Digest, item.Filename, item.Source)
}
```

This path is intentionally separate from `FetchStoreMediaURLs`. Store item assets
cover Store/Library images such as headers, capsules, library hero, logo, page
background, and community icon. Storefront media covers screenshots, movies, and
Storefront backgrounds returned by appdetails.

By default, Steam's returned query string such as `?t=` is preserved. Set
`StoreItemAssetOptions.StripQuery` only when your downstream storage deliberately
does not want cache-version query parameters.

## Hash-Based Community and Client Icons

Community and client icon URLs require a known hash.

Examples:

```text
https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/apps/107100/8377b4460f19465c261673f76f2656bdb3288273.jpg
https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/apps/107100/ad7f9414231a7a5bb96d74e21893a84972dcbee8.ico
```

The hash is not derived from the AppID by `addons/assets`.

It must come from another source, such as Steam-owned game metadata, appinfo/client metadata, or another trusted source available to the caller. The addon only builds the final URL once the caller already has the AppID/hash pair.

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

Related kinds:

| `addons/assets` kind | Meaning |
|---|---|
| `assets.KindCommunityIconJPG` | community/app icon JPG when a hash is known |
| `assets.KindCommunityLogoJPG` | community/logo JPG when a hash is known |
| `assets.KindClientIconICO` | client shortcut icon ICO when a hash is known |

## Storefront Media URLs

Some useful media is better resolved from Storefront appdetails instead of guessed from a static path.

`addons/assets` can request Storefront appdetails through the existing `client.Web.Storefront` service and extract:

- Store page background
- raw Store page background
- screenshot thumbnails
- full screenshots
- movie thumbnails
- movie WebM URLs
- movie MP4 URLs
- movie DASH/HLS playlist URLs

These are represented by Store media kinds such as:

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

DASH/HLS helpers return the playlist or manifest URL itself. The addon does not expand playlists into video segments.

## Basic Usage

Import the addon:

```go
import "github.com/gofurry/steam-go/addons/assets"
```

Build simple URLs:

```go
headers := assets.HeaderURLs(550, 107100)
heroes := assets.URLs(assets.KindLibraryHero, 550, 107100)
logos := assets.LibraryLogo2xURLs(550, 107100)
```

Build one struct per AppID:

```go
all := assets.AllWithLanguage("schinese", 550, 107100)

for _, item := range all {
    fmt.Println(item.AppID, item.Header, item.LibraryHero, item.LibraryLogo2x)
}
```

Build a flat list of typed URL items:

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

## Verify URLs

Not every AppID has every asset.

Use verification when existence matters:

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

For direct URLs:

```go
results, err := assets.VerifyURLsWithOptions(ctx, assets.VerifyOptions{
    URLValidator: assets.SteamStaticURLValidator,
}, "https://shared.steamstatic.com/store_item_assets/steam/apps/550/header.jpg")
```

`VerifyURLs` uses `HEAD` first and falls back to `GET` when the server returns `405` or `501`. HTTP non-2xx responses are reported as `Exists=false` rather than treated as hard errors.

## Read Resources Into Memory

For small batches:

```go
results, err := assets.ReadAppAssets(ctx, assets.ReadAppOptions{
    Kinds: []assets.Kind{assets.KindHeader, assets.KindLibraryHero},
    MaxBytes: 8 << 20,
}, 550)
```

For large batches, prefer streaming one result at a time:

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

`ReadURLs` defaults to a 32 MiB per-resource limit. Set `MaxBytes` deliberately for larger resources.

## Download Assets

Download generated AppID assets:

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

Download modes:

| Mode | Behavior |
|---|---|
| `assets.StoreFlat` | writes generated files directly under the destination directory, with AppID-prefixed names such as `550_header.jpg` |
| `assets.StoreByAppID` | writes files under child folders such as `550/header.jpg` |

Download results include:

- `DownloadStatusDownloaded`
- `DownloadStatusSkipped`
- `DownloadStatusFailed`

Batch downloads try every URL. Successful files remain on disk even when later items fail.

## Fetch Storefront Media

Storefront media uses the existing `client.Web.Storefront` service.

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

You can also verify, read, or download Storefront media:

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

## Write a Manifest

Manifests are useful for crawlers, build pipelines, bots, and frontend indexing.

```go
items := assets.ListWithLanguage("schinese", 550, 107100)
manifest := assets.NewURLManifest(items)

if err := assets.WriteManifestJSON("./tmp/assets/manifest.json", manifest); err != nil {
    return err
}
```

For downloads:

```go
results, err := assets.DownloadAppAssets(ctx, opts, 550, 107100)
if err != nil {
    fmt.Println("some downloads failed:", err)
}

manifest := assets.NewDownloadManifest(results)
_ = assets.WriteManifestJSON("./tmp/assets/download-manifest.json", manifest)
```

## Direct URL Safety

Some helpers accept direct caller-supplied URLs:

- `VerifyURLs`
- `ReadURLs`
- `DownloadURLs`

If those URLs come from users or another untrusted source, set a validator before sending HTTP requests.

```go
results, err := assets.ReadURLsWithOptions(ctx, assets.ReadOptions{
    URLValidator: assets.SteamStaticURLValidator,
}, urls...)
```

Available validators:

```go
assets.AllowHosts("shared.steamstatic.com")
assets.AllowHostSuffixes("steamstatic.com")
assets.SteamStaticURLValidator
```

`SteamStaticURLValidator` accepts hosts under `steamstatic.com`. Use stricter validators when your application needs stricter host controls.

## What `addons/assets` Does Not Do

`addons/assets` intentionally stays small.

It does not:

- create a `steam.Client`
- integrate SteamGridDB
- parse Steam client appinfo
- discover client icon hashes by itself
- guarantee that every generated URL exists
- expand DASH/HLS playlists into video segments
- treat public static URLs as official stable Web API endpoints

It is a practical toolkit for constructing, validating, reading, and downloading public Steam asset resources.

## Practical Advice

- Build AppID-only static Store and Library URLs when you need predictable image candidates.
- Verify generated URLs before showing or downloading them in production workflows.
- Use Storefront media helpers for screenshots, movies, and backgrounds because those are better discovered from appdetails.
- Keep hash-based community/client icon handling separate from AppID-only URL construction.
- For large batches, use low concurrency and streaming read helpers.
- Keep manifests so downstream jobs know exactly which URLs were resolved and which downloads succeeded.
- Prefer URLs returned by Steam metadata when available.
- Treat Steam static resource paths as public resources, not as an official enumeration API.

## CLI Example

The repository includes an example command:

```bash
go run ./examples/assets -app-ids 550,107100 -language schinese
```

Verify generated app assets:

```bash
go run ./examples/assets -verify-apps -kind all
```

Download generated assets:

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

Fetch Storefront media:

```bash
go run ./examples/assets -app-ids 550 -store-media -kind all
```

## Related Wiki Pages

- [Public Store Page Access Notes](https://github.com/gofurry/steam-go/wiki/Public-Store-Page-Access-Notes)
- [Traffic Policy](https://github.com/gofurry/steam-go/wiki/Traffic-Policy)
- [Rate Limiting Strategy](https://github.com/gofurry/steam-go/wiki/Rate-Limiting-Strategy)
- [Proxy and Network Strategy](https://github.com/gofurry/steam-go/wiki/Proxy-and-Network-Strategy)
