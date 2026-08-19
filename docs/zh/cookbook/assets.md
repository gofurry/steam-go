# Cookbook：Assets Addon

使用 `addons/assets` 获取公开 Store、Library 与玩家/Profile 资源 URL。

## 构造静态 URL

```go
headers := assets.HeaderURLs(440, 570)
heroes := assets.URLs(assets.KindLibraryHero, 440, 570)

fmt.Println(headers)
fmt.Println(heroes)
```

## 验证 App 资源

```go
results, err := assets.VerifyAppAssets(ctx, assets.VerifyAppOptions{
	Kinds: []assets.Kind{assets.KindHeader, assets.KindLibraryHero},
}, 440, 570)
if err != nil {
	panic(err)
}

for _, result := range results {
	fmt.Printf("%d %s exists=%v\n", result.AppID, result.Kind, result.Exists)
}
```

## 发现官方 Store Item Assets

当较新的 App 使用无法只根据 AppID 推导的 hashed asset path 时，使用
StoreBrowse-backed discovery。

```go
client, err := steam.NewClient(steam.WithSafeDefaults())
if err != nil {
	panic(err)
}
defer client.Close()

items, err := assets.FetchStoreItemAssetURLs(ctx, client.API.StoreBrowseService, assets.StoreItemAssetOptions{
	CountryCode: "US",
	Language:    "english",
	Kinds: []assets.Kind{
		assets.KindHeader2x,
		assets.KindLibraryHero2x,
	},
}, 4710650)
if err != nil {
	panic(err)
}

for _, item := range items {
	fmt.Printf("%d %s %s %s\n", item.AppID, item.Kind, item.Digest, item.URL)
}
```

## 发现玩家资源

玩家头像来自 official `ISteamUser/GetPlayerSummaries/v2` 返回的 URL。已装备的
Profile 背景、迷你背景、头像框与动态头像媒体来自当前 observed
`IPlayerService/GetProfileItemsEquipped/v1` surface。

```go
avatars, err := assets.FetchPlayerAvatarURLs(
	ctx,
	client.API.SteamUser,
	assets.PlayerAvatarOptions{},
	"76561198000000000",
)
if err != nil {
	panic(err)
}

profile, err := assets.FetchEquippedProfileAssetURLs(
	ctx,
	client.API.PlayerService,
	assets.PlayerProfileAssetOptions{Language: "schinese"},
	"76561198000000000",
)
if err != nil {
	panic(err)
}

fmt.Println(avatars, profile)
```

这些 helper 直接使用 Steam 返回的 URL，不根据 avatar hash 猜 CDN path，也不根据
Profile item ID 猜资源路径。verify、read、download variant 都会保留 optional
`SteamID` metadata；玩家资源下载路径为 `<Dir>/<SteamID>/<kind>.<ext>`。

## 说明

- 静态 URL builder 不发起网络请求。
- verify、read、download、Store media discovery、Store item asset discovery 和玩家资源 discovery helper 会显式发起网络请求。
- 如果 direct URL 来自不可信输入，应配置 URL validator，例如 `assets.SteamStaticURLValidator`。
- 完整示例：`go run ./examples/assets -app-ids 4710650 -store-item-assets -kind all`。
- 玩家资源 live 示例：`go run ./examples/live/playerassets -verify`（需要配置 Steam 凭据）。
