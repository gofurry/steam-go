package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/addons/assets"
	"github.com/gofurry/steam-go/examples/live/internal/realtest"
)

func main() {
	steamID := flag.String("steam-id", realtest.DefaultSteamID, "SteamID64 whose player assets should be inspected")
	vanity := flag.String("vanity", "", "optional Steam Community vanity token to resolve")
	language := flag.String("language", "english", "Profile item language")
	verify := flag.Bool("verify", false, "verify every discovered player asset URL")
	readSmall := flag.Bool("read-small", false, "read a bounded set of small player images into memory")
	downloadDir := flag.String("download-dir", "", "optional root directory for small player image downloads")
	concurrency := flag.Int("concurrency", 2, "read/download concurrency")
	timeout := flag.Duration("timeout", 30*time.Second, "HTTP timeout")
	flag.Parse()

	cfg, err := realtest.LoadConfig()
	if err != nil {
		realtest.Fatalf("load config failed: %v", err)
	}
	if cfg.Key == "" && cfg.AccessToken == "" {
		fmt.Println("skip: set STEAM_API_KEY or STEAM_ACCESS_TOKEN, or add examples/live/key.txt or access-token.txt")
		return
	}

	client, err := realtest.NewClient(cfg)
	if err != nil {
		realtest.Fatalf("create client failed: %v", err)
	}
	defer client.Close()

	assetHTTPClient, err := steam.NewHTTPClientWithProxySelector(cfg.ProxySelector, *timeout)
	if err != nil {
		realtest.Fatalf("create asset HTTP client failed: %v", err)
	}
	defer assetHTTPClient.CloseIdleConnections()

	ctx, cancel := context.WithTimeout(realtest.BackgroundContext(), 2*time.Minute)
	defer cancel()
	realtest.PrintProxy(cfg)

	if *vanity != "" {
		resolved, err := client.API.SteamUser.ResolveVanityURL(ctx, *vanity, nil)
		if err != nil {
			realtest.Fatalf("ResolveVanityURL failed: %v", err)
		}
		fmt.Printf("vanity=%q success=%d steamid=%s message=%q\n", *vanity, resolved.Response.Success, resolved.Response.SteamID, resolved.Response.Message)
	}

	avatarItems, err := assets.FetchPlayerAvatarURLs(
		ctx,
		client.API.SteamUser,
		assets.PlayerAvatarOptions{},
		*steamID,
	)
	if err != nil {
		realtest.Fatalf("FetchPlayerAvatarURLs failed: %v", err)
	}
	printURLItems("avatar", avatarItems)

	profileItems, err := assets.FetchEquippedProfileAssetURLs(
		ctx,
		client.API.PlayerService,
		assets.PlayerProfileAssetOptions{Language: *language},
		*steamID,
	)
	if err != nil {
		realtest.Fatalf("FetchEquippedProfileAssetURLs failed: %v", err)
	}
	printURLItems("profile", profileItems)

	if *verify {
		verifyPlayerAssets(ctx, client, assetHTTPClient, *steamID, *language)
	}
	if *readSmall {
		readSmallPlayerAssets(ctx, client, assetHTTPClient, *steamID, *language, *concurrency)
	}
	if *downloadDir != "" {
		downloadSmallPlayerAssets(ctx, client, assetHTTPClient, *steamID, *language, *downloadDir, *concurrency)
	}
}

func verifyPlayerAssets(ctx context.Context, client *steam.Client, assetHTTPClient *http.Client, steamID, language string) {
	avatars, err := assets.VerifyPlayerAvatars(ctx, client.API.SteamUser, assets.VerifyPlayerAvatarOptions{
		HTTPClient: assetHTTPClient,
	}, steamID)
	if err != nil {
		realtest.Fatalf("VerifyPlayerAvatars failed: %v", err)
	}
	printVerifyResults("avatar verify", avatars)

	profile, err := assets.VerifyEquippedProfileAssets(ctx, client.API.PlayerService, assets.VerifyPlayerProfileAssetOptions{
		Language:   language,
		HTTPClient: assetHTTPClient,
	}, steamID)
	if err != nil {
		realtest.Fatalf("VerifyEquippedProfileAssets failed: %v", err)
	}
	printVerifyResults("profile verify", profile)
}

func readSmallPlayerAssets(ctx context.Context, client *steam.Client, assetHTTPClient *http.Client, steamID, language string, concurrency int) {
	avatars, avatarErr := assets.ReadPlayerAvatars(ctx, client.API.SteamUser, assets.ReadPlayerAvatarOptions{
		Kinds:       []assets.Kind{assets.KindPlayerAvatarMedium},
		HTTPClient:  assetHTTPClient,
		MaxBytes:    8 << 20,
		Concurrency: concurrency,
	}, steamID)
	printReadResults("avatar read", avatars)
	if avatarErr != nil {
		fmt.Printf("avatar read error=%v\n", avatarErr)
	}

	profile, profileErr := assets.ReadEquippedProfileAssets(ctx, client.API.PlayerService, assets.ReadPlayerProfileAssetOptions{
		Language:    language,
		Kinds:       smallProfileImageKinds(),
		HTTPClient:  assetHTTPClient,
		MaxBytes:    8 << 20,
		Concurrency: concurrency,
	}, steamID)
	printReadResults("profile read", profile)
	if profileErr != nil {
		fmt.Printf("profile read error=%v\n", profileErr)
	}
}

func downloadSmallPlayerAssets(ctx context.Context, client *steam.Client, assetHTTPClient *http.Client, steamID, language, dir string, concurrency int) {
	avatars, avatarErr := assets.DownloadPlayerAvatars(ctx, client.API.SteamUser, assets.DownloadPlayerAvatarOptions{
		Dir:         filepath.Join(dir, "avatars"),
		Kinds:       []assets.Kind{assets.KindPlayerAvatarMedium},
		HTTPClient:  assetHTTPClient,
		Overwrite:   assets.OverwriteNever,
		Concurrency: concurrency,
	}, steamID)
	printDownloadResults("avatar download", avatars)
	if avatarErr != nil {
		fmt.Printf("avatar download error=%v\n", avatarErr)
	}

	profile, profileErr := assets.DownloadEquippedProfileAssets(ctx, client.API.PlayerService, assets.DownloadPlayerProfileAssetOptions{
		Dir:         filepath.Join(dir, "profile"),
		Language:    language,
		Kinds:       smallProfileImageKinds(),
		HTTPClient:  assetHTTPClient,
		Overwrite:   assets.OverwriteNever,
		Concurrency: concurrency,
	}, steamID)
	printDownloadResults("profile download", profile)
	if profileErr != nil {
		fmt.Printf("profile download error=%v\n", profileErr)
	}
}

func smallProfileImageKinds() []assets.Kind {
	return []assets.Kind{
		assets.KindProfileBackgroundSmall,
		assets.KindMiniProfileBackgroundSmall,
		assets.KindAvatarFrameSmall,
		assets.KindAnimatedAvatarSmall,
	}
}

func printURLItems(label string, items []assets.URLItem) {
	fmt.Printf("== %s URLs (%d) ==\n", label, len(items))
	for _, item := range items {
		fmt.Printf("steamid=%s kind=%s name=%q url=%s\n", item.SteamID, item.Kind, item.Name, item.URL)
	}
}

func printVerifyResults(label string, results []assets.VerifyResult) {
	fmt.Printf("== %s (%d) ==\n", label, len(results))
	for _, result := range results {
		fmt.Printf("steamid=%s kind=%s exists=%t status=%d type=%s\n", result.SteamID, result.Kind, result.Exists, result.StatusCode, result.ContentType)
	}
}

func printReadResults(label string, results []assets.ReadResult) {
	fmt.Printf("== %s (%d) ==\n", label, len(results))
	for _, result := range results {
		fmt.Printf("steamid=%s kind=%s bytes=%d status=%d error=%q\n", result.SteamID, result.Kind, result.BytesRead, result.StatusCode, result.Error)
	}
}

func printDownloadResults(label string, results []assets.DownloadResult) {
	fmt.Printf("== %s (%d) ==\n", label, len(results))
	for _, result := range results {
		fmt.Printf("steamid=%s kind=%s status=%s path=%s error=%q\n", result.SteamID, result.Kind, result.Status, result.Path, result.Error)
	}
}
