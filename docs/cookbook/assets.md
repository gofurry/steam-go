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

## Notes

- Static URL builders do not perform network requests.
- Verification, read, download, and Store media discovery helpers do perform explicit network requests.
- For direct URLs from untrusted input, configure a URL validator such as `assets.SteamStaticURLValidator`.
- Full example: `go run ./examples/assets -app-ids 550,107100`.
