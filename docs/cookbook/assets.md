# Cookbook: Assets Addon

Use `addons/assets` for public Store, Library, and player/Profile asset URLs.

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

## Discover Player Assets

Player avatars come from the URLs returned by official
`ISteamUser/GetPlayerSummaries/v2`. Equipped Profile backgrounds, mini-profile
backgrounds, avatar frames, and animated-avatar media come from the currently
observed `IPlayerService/GetProfileItemsEquipped/v1` surface.

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
	assets.PlayerProfileAssetOptions{Language: "english"},
	"76561198000000000",
)
if err != nil {
	panic(err)
}

fmt.Println(avatars, profile)
```

The helpers use Steam-returned URLs and never guess avatar CDN paths from a
hash or Profile paths from item IDs. Verify, read, and download variants keep
the optional `SteamID` metadata; player downloads use
`<Dir>/<SteamID>/<kind>.<ext>`.

## Notes

- Static URL builders do not perform network requests.
- Verification, read, download, Store media discovery, Store item asset discovery, and player asset discovery helpers do perform explicit network requests.
- For direct URLs from untrusted input, configure a URL validator such as `assets.SteamStaticURLValidator`.
- Full example: `go run ./examples/assets -app-ids 4710650 -store-item-assets -kind all`.
- Live player example: `go run ./examples/live/playerassets -verify` (requires configured Steam credentials).
