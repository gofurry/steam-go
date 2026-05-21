# Addons

This document keeps the addon-level usage notes that would otherwise make the repository homepage too noisy.

## `addons/openid`

Use `addons/openid` when you want browser-based Steam sign-in.

What it does:

- builds the Steam OpenID login URL
- verifies the callback against Steam with `check_authentication`
- returns `SteamID64`, `ClaimedID`, and the recovered `state`

What it does not do:

- it does not replace Web API credentials
- it does not fetch profile data by itself
- it does not manage sessions for you

The example now shows a more realistic pattern:

- generate a random `state`
- store it in a cookie before redirecting to Steam
- verify the callback with `addons/openid`
- compare the returned `state` with the cookie
- clear the cookie after a successful login

Run it in direct mode:

```bash
go run ./examples/openid
```

Run it with a proxy for the server-side verification request:

```bash
go run ./examples/openid --proxy http://127.0.0.1:7897
```

This is especially useful for China-region networks where the browser login page may work while the server-side `check_authentication` request still needs a proxy.

## `addons/websession`

Use `addons/websession` when you want one manual Steam web-login workflow built on top of `client.API.AuthenticationService`.

What it does:

- starts one credentials-based auth session
- accepts one optional Steam Guard code submission
- polls the auth session until Steam issues tokens
- exchanges one refresh token for Store / Community web cookies
- validates both Store and Community sessions

What it does not do:

- it does not persist passwords, refresh tokens, or cookies for you
- it does not read browser cookies or Steam client local state
- it does not silently print secrets in the example output
- it does not persist or rotate one refresh token for you after login succeeds

Construction modes:

- `websession.NewClientFromSteamClient(...)` is the recommended path and reuses the root SDK traffic-policy execution stack for Store / Community web traffic
- `websession.NewClient(...)` remains available as manual mode when you want to supply your own `http.Client`

The example accepts `-account` and `-proxy`, and uses hidden terminal prompts for sensitive values when environment variables are not set:

- `STEAM_ACCOUNT_NAME`
- `STEAM_PASSWORD`
- `STEAM_GUARD_CODE`

Run it:

```bash
go run ./examples/websession
```

## `addons/assets`

Use `addons/assets` when you want high-value public Store and Library image asset URLs from one or more Steam AppIDs.

What it does:

- locally builds static URLs such as `header.jpg`, `capsule_616x353.jpg`, `library_600x900_2x.jpg`, `library_hero.jpg`, and `logo_2x.png`
- provides one helper per asset family, for example `assets.HeaderURLs(...)` and `assets.LibraryLogoURLs(...)`
- provides preset kind groups such as `assets.StoreKinds()`, `assets.LibraryKinds()`, and `assets.AllKinds()`
- provides resource list helpers such as `assets.ListWithLanguage(...)`
- provides `assets.URLs(kind, appIDs...)` when the asset family is chosen at runtime
- provides `assets.All(...)` / `assets.AllWithLanguage(...)` when you want one struct per AppID
- verifies URL existence with `assets.VerifyURLs(...)` and `assets.VerifyAppAssets(...)`
- reads resources into memory with `assets.ReadURLs(...)` and `assets.ReadAppAssets(...)`
- downloads arbitrary URLs with `assets.DownloadURLs(...)`
- downloads constructed AppID assets with `assets.DownloadAppAssets(...)`
- requests Storefront screenshot, movie, and background URLs with `assets.FetchStoreMediaURLs(...)`
- verifies, reads, or downloads those Storefront media URLs with `assets.VerifyStoreMedia(...)`, `assets.ReadStoreMedia(...)`, and `assets.DownloadStoreMedia(...)`
- writes URL/download manifests with `assets.WriteManifestJSON(...)`

What it does not do:

- it does not create a client
- static URL construction does not call Steam; Store media discovery, verification, and download helpers perform explicit network requests
- it does not integrate SteamGridDB

Example:

```go
headers := assets.HeaderURLs(550, 107100)
heroes := assets.URLs(assets.KindLibraryHero, 550, 107100)
all := assets.AllWithLanguage("schinese", 550, 107100)
```

Exported helpers:

- `HeaderURLs`, `HeaderLocalizedURLs`
- `CapsuleSmallURLs`, `CapsuleMainURLs`
- `LibraryCapsuleURLs`, `LibraryCapsule2xURLs`, `LibraryHeroURLs`, `LibraryLogoURLs`, `LibraryLogo2xURLs`
- `StoreKinds`, `StoreKindsWithLocalized`, `LibraryKinds`, `DefaultKinds`, `AllKinds`
- `StoreMediaKinds`, `StoreBackgroundKinds`, `StoreScreenshotKinds`, `StoreMovieKinds`
- `List`, `ListWithLanguage`, `ListKinds`, `ListKindsWithLanguage`
- `URLs(kind, appIDs...)`, `LocalizedURLs(kind, language, appIDs...)`
- `All(appIDs...)`, `AllWithLanguage(language, appIDs...)`
- `CommunityIconURL`, `CommunityIconURLs`, `CommunityLogoURL`, `CommunityLogoURLs`, `ClientIconURL`, `ClientIconURLs` for known AppID/hash pairs
- `VerifyURLs(ctx, urls...)`, `VerifyURLsWithClient(ctx, httpClient, urls...)`, `VerifyAppAssets(ctx, opts, appIDs...)`
- `ReadURLs(ctx, urls...)`, `ReadURLsWithClient(ctx, httpClient, urls...)`, `ReadURLsWithOptions(ctx, ReadOptions{...}, urls...)`
- `ReadAppAssets(ctx, ReadAppOptions{...}, appIDs...)`
- `DownloadURLs(ctx, dir, urls...)`, `DownloadURLsWithClient(ctx, httpClient, dir, urls...)`
- `DownloadAppAssets(ctx, DownloadAppOptions{...}, appIDs...)`
- `FetchStoreMediaURLs(ctx, storefront, StoreMediaOptions{...}, appIDs...)`
- `VerifyStoreMedia(ctx, storefront, VerifyStoreMediaOptions{...}, appIDs...)`
- `ReadStoreMedia(ctx, storefront, ReadStoreMediaOptions{...}, appIDs...)`
- `DownloadStoreMedia(ctx, storefront, DownloadStoreMediaOptions{...}, appIDs...)`
- `NewURLManifest`, `NewDownloadManifest`, `MarshalManifestJSON`, `WriteManifestJSON`

Download modes:

- `assets.StoreFlat`: writes generated app asset files directly under the destination directory with AppID-prefixed names such as `550_header.jpg`
- `assets.StoreByAppID`: writes generated app asset files under child folders such as `550/header.jpg`

Batch downloads try every URL. Successful files remain on disk, and failed items are reported in both their `DownloadResult.Error` field and the final joined error.

Download options include `FilenameStyle`, `Overwrite`, `SkipExisting`, and `Concurrency`. Results include `DownloadStatusDownloaded`, `DownloadStatusSkipped`, or `DownloadStatusFailed`.

Read helpers return `ReadResult.Data` as `[]byte` for callers that want to process the resource themselves. They default to a 32 MiB per-resource limit; set `MaxBytes` explicitly for larger files.

For Storefront movie resources, DASH/HLS kinds save the `.mpd` / `.m3u8` playlist URL itself. The addon does not expand playlists into video segments.

Run the example:

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

## `addons/freeclaim`

Use `addons/freeclaim` when you want one addon-level bridge for Store promotion discovery and one explicit free-license claim.

What it does:

- searches current free Store promotions from the Store search HTML fragment response
- resolves one app's free package candidates from `client.Web.Storefront.GetAppDetails`
- checks app ownership through `dynamicstore/userdata`
- claims one free package only when you explicitly request it

What it does not do:

- it does not manage account passwords or browser cookies
- it does not read Steam client local tokens or any local account database
- it does not auto-claim everything
- it does not retry forever
- it does not silently escalate into bulk-claim automation

Construction modes:

- `freeclaim.NewClientFromSteamClient(...)` is the recommended path and reuses the root SDK traffic-policy execution stack for Store web traffic
- `freeclaim.NewClient(...)` remains available as manual mode when you want to supply your own `http.Client`

The example is read-only by default. Claim mode requires a refresh token through `STEAM_REFRESH_TOKEN` or one hidden terminal prompt.

Run the read-only search / package resolution example:

```bash
go run ./examples/freeclaim
```

Run an explicit claim after choosing an app and package:

```bash
go run ./examples/freeclaim -app-id 480 -package-id 12345 -claim
```

## `addons/a2s`

Use `addons/a2s` when you want to query a game server directly without pulling `a2s-go` into your own import tree.

The bridge currently re-exports the upstream `a2s-go` client and key result types, so you can call:

- `QueryInfo`
- `QueryPlayers`
- `QueryRules`

Example:

```bash
go run ./examples/a2s -server 1.2.3.4:27015 -query info
go run ./examples/a2s -server 1.2.3.4:27015 -query players
go run ./examples/a2s -server 1.2.3.4:27015 -query rules
```

## `addons/a2s/master`

Use `addons/a2s/master` when you want discovery against the A2S master server protocol.

The bridge follows the upstream `a2s-go/master` surface and is intended for:

- one-page discovery with `Query`
- streamed discovery with `Stream`

## `addons/a2s/scanner`

Use `addons/a2s/scanner` when you want to batch probe discovered addresses.

The bridge follows the upstream `a2s-go/scanner` package and supports:

- probing address lists
- consuming discovery streams
- batch `info`, `players`, and `rules` queries

When `a2s-go` publishes new stable releases, `steam-go` should keep the bridge version and examples in sync rather than re-implementing A2S logic locally.
