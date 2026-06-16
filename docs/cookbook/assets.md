# Cookbook: Assets Addon

Use `addons/assets` for public Store and Library asset URLs.

## Build Static URLs

```go
headers := assets.HeaderURLs(440, 570)
heroes := assets.URLs(assets.KindLibraryHero, 440, 570)

fmt.Println(headers)
fmt.Println(heroes)
```

## Verify App Assets

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

## Discover Official Store Item Assets

Use StoreBrowse-backed discovery when newer apps use hashed asset paths that
cannot be derived from the AppID alone.

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

## Notes

- Static URL builders do not perform network requests.
- Verification, read, download, Store media discovery, and Store item asset discovery helpers do perform explicit network requests.
- For direct URLs from untrusted input, configure a URL validator such as `assets.SteamStaticURLValidator`.
- Full example: `go run ./examples/assets -app-ids 4710650 -store-item-assets -kind all`.
