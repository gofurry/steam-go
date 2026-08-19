package assets

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gofurry/steam-go/addons/assets/internal/httpasset"
	"github.com/gofurry/steam-go/api/steamuser"
)

var defaultPlayerAvatarKinds = []Kind{
	KindPlayerAvatar,
	KindPlayerAvatarMedium,
	KindPlayerAvatarFull,
}

// FetchPlayerAvatarURLs returns avatar URLs exactly as provided by
// ISteamUser/GetPlayerSummaries. Results follow input Steam ID order and then
// requested kind order; AvatarHash is never used to construct fallback URLs.
func FetchPlayerAvatarURLs(ctx context.Context, service *steamuser.Service, opts PlayerAvatarOptions, steamIDs ...string) ([]URLItem, error) {
	if service == nil {
		return nil, fmt.Errorf("steamuser service must not be nil")
	}
	kinds, err := playerAvatarKinds(opts.Kinds)
	if err != nil {
		return nil, err
	}
	if len(steamIDs) == 0 {
		return []URLItem{}, nil
	}

	response, err := service.GetPlayerSummaries(ctx, steamIDs)
	if err != nil {
		return nil, err
	}
	players := make(map[string]steamuser.Player, len(response.Response.Players))
	for _, player := range response.Response.Players {
		players[player.SteamID] = player
	}

	items := make([]URLItem, 0, len(steamIDs)*len(kinds))
	for _, requestedSteamID := range steamIDs {
		steamID := strings.TrimSpace(requestedSteamID)
		player, ok := players[steamID]
		if !ok {
			continue
		}
		for _, kind := range kinds {
			rawURL := playerAvatarURL(player, kind)
			if rawURL == "" {
				continue
			}
			items = append(items, URLItem{
				SteamID: steamID,
				Kind:    kind,
				URL:     rawURL,
				Source:  SourceSteamUserPlayerSummaries,
			})
		}
	}
	return items, nil
}

// VerifyPlayerAvatars discovers and verifies player avatar URLs.
func VerifyPlayerAvatars(ctx context.Context, service *steamuser.Service, opts VerifyPlayerAvatarOptions, steamIDs ...string) ([]VerifyResult, error) {
	items, err := FetchPlayerAvatarURLs(ctx, service, PlayerAvatarOptions{Kinds: opts.Kinds}, steamIDs...)
	if err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	results := make([]VerifyResult, 0, len(items))
	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return results, err
		}
		result, err := verifyURLItem(ctx, opts.HTTPClient, item)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

// ReadPlayerAvatars discovers player avatars and reads them into memory.
func ReadPlayerAvatars(ctx context.Context, service *steamuser.Service, opts ReadPlayerAvatarOptions, steamIDs ...string) ([]ReadResult, error) {
	items, err := FetchPlayerAvatarURLs(ctx, service, PlayerAvatarOptions{Kinds: opts.Kinds}, steamIDs...)
	if err != nil {
		return nil, err
	}
	return readURLItems(ctx, opts.HTTPClient, opts.MaxBytes, opts.Concurrency, items)
}

// DownloadPlayerAvatars discovers and downloads player avatars beneath
// <Dir>/<SteamID>/<kind>.<ext>. A missing URL extension falls back to .jpg.
func DownloadPlayerAvatars(ctx context.Context, service *steamuser.Service, opts DownloadPlayerAvatarOptions, steamIDs ...string) ([]DownloadResult, error) {
	if strings.TrimSpace(opts.Dir) == "" {
		return nil, fmt.Errorf("download dir must not be empty")
	}
	items, err := FetchPlayerAvatarURLs(ctx, service, PlayerAvatarOptions{Kinds: opts.Kinds}, steamIDs...)
	if err != nil {
		return nil, err
	}
	requests := playerDownloadRequests(opts.Dir, items, playerAvatarFallbackExtension)
	return downloadRequests(ctx, opts.HTTPClient, effectiveOverwrite(opts.Overwrite, opts.SkipExisting), opts.Concurrency, requests)
}

func playerAvatarKinds(kinds []Kind) ([]Kind, error) {
	if len(kinds) == 0 {
		return append([]Kind(nil), defaultPlayerAvatarKinds...), nil
	}
	for _, kind := range kinds {
		switch kind {
		case KindPlayerAvatar, KindPlayerAvatarMedium, KindPlayerAvatarFull:
		default:
			return nil, fmt.Errorf("unsupported player avatar kind %q", kind)
		}
	}
	return append([]Kind(nil), kinds...), nil
}

func playerAvatarURL(player steamuser.Player, kind Kind) string {
	var rawURL string
	switch kind {
	case KindPlayerAvatar:
		rawURL = player.Avatar
	case KindPlayerAvatarMedium:
		rawURL = player.AvatarMedium
	case KindPlayerAvatarFull:
		rawURL = player.AvatarFull
	}
	return strings.TrimSpace(rawURL)
}

func playerDownloadRequests(dir string, items []URLItem, fallbackExtension func(Kind) string) []downloadRequest {
	requests := make([]downloadRequest, 0, len(items))
	usedPaths := make(map[string]int)
	for _, item := range items {
		extension := playerAssetExtension(item.URL, fallbackExtension(item.Kind))
		path := filepath.Join(dir, item.SteamID, string(item.Kind)+extension)
		requests = append(requests, downloadRequest{
			item: item,
			url:  item.URL,
			path: uniqueDownloadPath(path, usedPaths),
		})
	}
	return requests
}

func playerAssetExtension(rawURL, fallback string) string {
	name, err := httpasset.Filename(rawURL)
	if err == nil {
		if extension := filepath.Ext(name); extension != "" {
			return extension
		}
	}
	return fallback
}

func playerAvatarFallbackExtension(Kind) string {
	return ".jpg"
}
