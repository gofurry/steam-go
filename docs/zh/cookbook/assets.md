# Cookbook：Assets Addon

使用 `addons/assets` 获取公开 Store 和 Library 资源 URL。

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

## 说明

- 静态 URL builder 不发起网络请求。
- verify、read、download、Store media discovery 和 Store item asset discovery helper 会显式发起网络请求。
- 如果 direct URL 来自不可信输入，应配置 URL validator，例如 `assets.SteamStaticURLValidator`。
- 完整示例：`go run ./examples/assets -app-ids 4710650 -store-item-assets -kind all`。
