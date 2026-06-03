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

## 说明

- 静态 URL builder 不发起网络请求。
- verify、read、download 和 Store media discovery helper 会显式发起网络请求。
- 如果 direct URL 来自不可信输入，应配置 URL validator，例如 `assets.SteamStaticURLValidator`。
- 完整示例：`go run ./examples/assets -app-ids 550,107100`。
