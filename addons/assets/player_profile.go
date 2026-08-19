package assets

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/gofurry/steam-go/api/playerservice"
)

var defaultPlayerProfileAssetKinds = []Kind{
	KindProfileBackgroundSmall,
	KindProfileBackgroundLarge,
	KindMiniProfileBackgroundSmall,
	KindMiniProfileBackgroundLarge,
	KindAvatarFrameSmall,
	KindAvatarFrameLarge,
	KindAnimatedAvatarSmall,
	KindAnimatedAvatarLarge,
	KindAnimatedAvatarWebM,
	KindAnimatedAvatarMP4,
	KindAnimatedAvatarWebMSmall,
	KindAnimatedAvatarMP4Small,
}

// FetchEquippedProfileAssetURLs returns URLs present in
// IPlayerService/GetProfileItemsEquipped responses. This observed PlayerService
// surface is not currently listed in Valve's public Web API reference. Relative
// paths without a returned host are skipped rather than assigned a guessed CDN.
func FetchEquippedProfileAssetURLs(ctx context.Context, service *playerservice.Service, opts PlayerProfileAssetOptions, steamIDs ...string) ([]URLItem, error) {
	if service == nil {
		return nil, fmt.Errorf("playerservice service must not be nil")
	}
	kinds, err := playerProfileAssetKinds(opts.Kinds)
	if err != nil {
		return nil, err
	}
	if len(steamIDs) == 0 {
		return []URLItem{}, nil
	}

	items := make([]URLItem, 0, len(steamIDs)*len(kinds))
	for _, requestedSteamID := range steamIDs {
		steamID := strings.TrimSpace(requestedSteamID)
		response, err := service.GetProfileItemsEquipped(ctx, requestedSteamID, &playerservice.PlayerProfileItemOptions{
			Language: opts.Language,
		})
		if err != nil {
			return items, err
		}
		for _, kind := range kinds {
			profileItem, rawURL := equippedProfileAsset(response.Response, kind)
			rawURL = normalizeReturnedProfileAssetURL(rawURL)
			if rawURL == "" {
				continue
			}
			name := strings.TrimSpace(profileItem.ItemTitle)
			if name == "" {
				name = strings.TrimSpace(profileItem.Name)
			}
			items = append(items, URLItem{
				SteamID: steamID,
				Kind:    kind,
				Name:    name,
				URL:     rawURL,
				Source:  SourcePlayerServiceProfileItemsEquipped,
			})
		}
	}
	return items, nil
}

// VerifyEquippedProfileAssets discovers and verifies equipped profile assets.
func VerifyEquippedProfileAssets(ctx context.Context, service *playerservice.Service, opts VerifyPlayerProfileAssetOptions, steamIDs ...string) ([]VerifyResult, error) {
	items, err := FetchEquippedProfileAssetURLs(ctx, service, PlayerProfileAssetOptions{
		Language: opts.Language,
		Kinds:    opts.Kinds,
	}, steamIDs...)
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

// ReadEquippedProfileAssets discovers equipped profile assets and reads them
// into memory.
func ReadEquippedProfileAssets(ctx context.Context, service *playerservice.Service, opts ReadPlayerProfileAssetOptions, steamIDs ...string) ([]ReadResult, error) {
	items, err := FetchEquippedProfileAssetURLs(ctx, service, PlayerProfileAssetOptions{
		Language: opts.Language,
		Kinds:    opts.Kinds,
	}, steamIDs...)
	if err != nil {
		return nil, err
	}
	return readURLItems(ctx, opts.HTTPClient, opts.MaxBytes, opts.Concurrency, items)
}

// DownloadEquippedProfileAssets discovers and downloads equipped profile
// assets beneath <Dir>/<SteamID>/<kind>.<ext>.
func DownloadEquippedProfileAssets(ctx context.Context, service *playerservice.Service, opts DownloadPlayerProfileAssetOptions, steamIDs ...string) ([]DownloadResult, error) {
	if strings.TrimSpace(opts.Dir) == "" {
		return nil, fmt.Errorf("download dir must not be empty")
	}
	items, err := FetchEquippedProfileAssetURLs(ctx, service, PlayerProfileAssetOptions{
		Language: opts.Language,
		Kinds:    opts.Kinds,
	}, steamIDs...)
	if err != nil {
		return nil, err
	}
	requests := playerDownloadRequests(opts.Dir, items, playerProfileFallbackExtension)
	return downloadRequests(ctx, opts.HTTPClient, effectiveOverwrite(opts.Overwrite, opts.SkipExisting), opts.Concurrency, requests)
}

func playerProfileAssetKinds(kinds []Kind) ([]Kind, error) {
	if len(kinds) == 0 {
		return append([]Kind(nil), defaultPlayerProfileAssetKinds...), nil
	}
	for _, kind := range kinds {
		if !isPlayerProfileAssetKind(kind) {
			return nil, fmt.Errorf("unsupported player profile asset kind %q", kind)
		}
	}
	return append([]Kind(nil), kinds...), nil
}

func isPlayerProfileAssetKind(kind Kind) bool {
	switch kind {
	case KindProfileBackgroundSmall, KindProfileBackgroundLarge,
		KindMiniProfileBackgroundSmall, KindMiniProfileBackgroundLarge,
		KindAvatarFrameSmall, KindAvatarFrameLarge,
		KindAnimatedAvatarSmall, KindAnimatedAvatarLarge,
		KindAnimatedAvatarWebM, KindAnimatedAvatarMP4,
		KindAnimatedAvatarWebMSmall, KindAnimatedAvatarMP4Small:
		return true
	default:
		return false
	}
}

func equippedProfileAsset(payload playerservice.ProfileItemsEquippedPayload, kind Kind) (playerservice.ProfileItem, string) {
	switch kind {
	case KindProfileBackgroundSmall:
		return payload.ProfileBackground, payload.ProfileBackground.ImageSmall
	case KindProfileBackgroundLarge:
		return payload.ProfileBackground, payload.ProfileBackground.ImageLarge
	case KindMiniProfileBackgroundSmall:
		return payload.MiniProfileBackground, payload.MiniProfileBackground.ImageSmall
	case KindMiniProfileBackgroundLarge:
		return payload.MiniProfileBackground, payload.MiniProfileBackground.ImageLarge
	case KindAvatarFrameSmall:
		return payload.AvatarFrame, payload.AvatarFrame.ImageSmall
	case KindAvatarFrameLarge:
		return payload.AvatarFrame, payload.AvatarFrame.ImageLarge
	case KindAnimatedAvatarSmall:
		return payload.AnimatedAvatar, payload.AnimatedAvatar.ImageSmall
	case KindAnimatedAvatarLarge:
		return payload.AnimatedAvatar, payload.AnimatedAvatar.ImageLarge
	case KindAnimatedAvatarWebM:
		return payload.AnimatedAvatar, payload.AnimatedAvatar.MovieWebM
	case KindAnimatedAvatarMP4:
		return payload.AnimatedAvatar, payload.AnimatedAvatar.MovieMP4
	case KindAnimatedAvatarWebMSmall:
		return payload.AnimatedAvatar, payload.AnimatedAvatar.MovieWebMSmall
	case KindAnimatedAvatarMP4Small:
		return payload.AnimatedAvatar, payload.AnimatedAvatar.MovieMP4Small
	default:
		return playerservice.ProfileItem{}, ""
	}
}

func normalizeReturnedProfileAssetURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if strings.HasPrefix(rawURL, "//") {
		parsed, err := url.Parse("https:" + rawURL)
		if err != nil || parsed.Host == "" {
			return ""
		}
		return parsed.String()
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" || (!strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https")) {
		return ""
	}
	return rawURL
}

func playerProfileFallbackExtension(kind Kind) string {
	switch kind {
	case KindAnimatedAvatarWebM, KindAnimatedAvatarWebMSmall:
		return ".webm"
	case KindAnimatedAvatarMP4, KindAnimatedAvatarMP4Small:
		return ".mp4"
	default:
		return ".jpg"
	}
}
