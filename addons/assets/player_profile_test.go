package assets

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/api/playerservice"
)

func TestFetchEquippedProfileAssetURLs(t *testing.T) {
	testAPI := newPlayerProfileTestAPI(t)
	items, err := FetchEquippedProfileAssetURLs(context.Background(), testAPI.service, PlayerProfileAssetOptions{
		Language: "japanese",
		Kinds: []Kind{
			KindProfileBackgroundLarge,
			KindAvatarFrameSmall,
			KindAnimatedAvatarMP4,
		},
	}, playerTwoSteamID, " "+playerOneSteamID+" ")
	if err != nil {
		t.Fatalf("FetchEquippedProfileAssetURLs returned error: %v", err)
	}
	if len(items) != 6 {
		t.Fatalf("items = %d, want 6: %#v", len(items), items)
	}
	if items[0].SteamID != playerTwoSteamID || items[0].Kind != KindProfileBackgroundLarge || items[0].URL != "https://cdn.example.test/profile/background-large.jpg" || items[0].Name != "Background Title" {
		t.Fatalf("unexpected first item: %#v", items[0])
	}
	if items[1].Kind != KindAvatarFrameSmall || items[1].Name != "Frame Name" {
		t.Fatalf("name fallback failed: %#v", items[1])
	}
	if items[3].SteamID != playerOneSteamID || items[3].Kind != KindProfileBackgroundLarge {
		t.Fatalf("input order was not preserved: %#v", items)
	}
	for _, item := range items {
		if item.Source != SourcePlayerServiceProfileItemsEquipped || item.AppID != 0 {
			t.Fatalf("unexpected profile metadata: %#v", item)
		}
	}
	for _, query := range testAPI.queries(t) {
		if query.Get("language") != "japanese" {
			t.Fatalf("language query = %q", query.Get("language"))
		}
	}
}

func TestFetchEquippedProfileAssetURLsDefaultsAndValidation(t *testing.T) {
	testAPI := newPlayerProfileTestAPI(t)
	items, err := FetchEquippedProfileAssetURLs(context.Background(), testAPI.service, PlayerProfileAssetOptions{}, playerOneSteamID)
	if err != nil {
		t.Fatalf("FetchEquippedProfileAssetURLs returned error: %v", err)
	}
	if len(items) != 11 {
		t.Fatalf("items = %d, want 11 after skipping one relative URL: %#v", len(items), items)
	}
	for _, item := range items {
		if item.Kind == KindMiniProfileBackgroundLarge {
			t.Fatalf("relative URL should have been skipped: %#v", item)
		}
	}
	if _, err := FetchEquippedProfileAssetURLs(context.Background(), testAPI.service, PlayerProfileAssetOptions{Kinds: []Kind{KindHeader}}, playerOneSteamID); err == nil {
		t.Fatal("expected invalid profile kind error")
	}
	if _, err := FetchEquippedProfileAssetURLs(context.Background(), nil, PlayerProfileAssetOptions{}, playerOneSteamID); err == nil {
		t.Fatal("expected nil service error")
	}
	empty, err := FetchEquippedProfileAssetURLs(context.Background(), testAPI.service, PlayerProfileAssetOptions{})
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("empty result = %#v, %v", empty, err)
	}
}

func TestEquippedProfileAssetsVerifyReadAndDownload(t *testing.T) {
	testAPI := newPlayerProfileTestAPI(t)
	verified, err := VerifyEquippedProfileAssets(context.Background(), testAPI.service, VerifyPlayerProfileAssetOptions{
		Kinds: []Kind{KindProfileBackgroundSmall, KindAvatarFrameSmall},
	}, playerOneSteamID)
	if err != nil {
		t.Fatalf("VerifyEquippedProfileAssets returned error: %v", err)
	}
	if len(verified) != 2 || !verified[0].Exists || verified[0].SteamID != playerOneSteamID || verified[1].Name != "Frame Name" {
		t.Fatalf("unexpected verify results: %#v", verified)
	}

	read, err := ReadEquippedProfileAssets(context.Background(), testAPI.service, ReadPlayerProfileAssetOptions{
		Kinds:       []Kind{KindAnimatedAvatarSmall},
		MaxBytes:    1024,
		Concurrency: 2,
	}, playerOneSteamID)
	if err != nil {
		t.Fatalf("ReadEquippedProfileAssets returned error: %v", err)
	}
	if len(read) != 1 || read[0].SteamID != playerOneSteamID || string(read[0].Data) != "profile:/profile/animated-small" {
		t.Fatalf("unexpected read results: %#v", read)
	}

	dir := t.TempDir()
	downloaded, err := DownloadEquippedProfileAssets(context.Background(), testAPI.service, DownloadPlayerProfileAssetOptions{
		Dir: dir,
		Kinds: []Kind{
			KindAnimatedAvatarSmall,
			KindAnimatedAvatarWebM,
			KindAnimatedAvatarMP4,
		},
		Concurrency: 2,
	}, playerOneSteamID)
	if err != nil {
		t.Fatalf("DownloadEquippedProfileAssets returned error: %v", err)
	}
	if len(downloaded) != 3 {
		t.Fatalf("downloads = %d, want 3", len(downloaded))
	}
	assertFile(t, filepath.Join(dir, playerOneSteamID, "animated_avatar_small.jpg"), "profile:/profile/animated-small")
	assertFile(t, filepath.Join(dir, playerOneSteamID, "animated_avatar_webm.webm"), "profile:/profile/animated.webm")
	assertFile(t, filepath.Join(dir, playerOneSteamID, "animated_avatar_mp4.mp4"), "profile:/profile/animated.mp4")
	for _, result := range downloaded {
		if result.SteamID != playerOneSteamID || result.Source != SourcePlayerServiceProfileItemsEquipped {
			t.Fatalf("profile metadata lost: %#v", result)
		}
	}
	if _, err := DownloadEquippedProfileAssets(context.Background(), testAPI.service, DownloadPlayerProfileAssetOptions{}, playerOneSteamID); err == nil {
		t.Fatal("expected empty download dir error")
	}
}

func TestEquippedProfileAssetDuplicatePathsAreUnique(t *testing.T) {
	testAPI := newPlayerProfileTestAPI(t)
	dir := t.TempDir()
	results, err := DownloadEquippedProfileAssets(context.Background(), testAPI.service, DownloadPlayerProfileAssetOptions{
		Dir:   dir,
		Kinds: []Kind{KindAvatarFrameSmall, KindAvatarFrameSmall},
	}, playerOneSteamID)
	if err != nil {
		t.Fatalf("DownloadEquippedProfileAssets returned error: %v", err)
	}
	if len(results) != 2 || results[0].Path == results[1].Path || !strings.HasSuffix(results[1].Path, "_2.png") {
		t.Fatalf("duplicate paths were not uniquified: %#v", results)
	}
}

func TestNormalizeReturnedProfileAssetURL(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		" https://cdn.example.test/item.jpg ": "https://cdn.example.test/item.jpg",
		"http://cdn.example.test/item.jpg":    "http://cdn.example.test/item.jpg",
		"//cdn.example.test/item.jpg":         "https://cdn.example.test/item.jpg",
		"relative/item.jpg":                   "",
		"file:///tmp/item.jpg":                "",
		"//":                                  "",
		"":                                    "",
	}
	for input, want := range tests {
		if got := normalizeReturnedProfileAssetURL(input); got != want {
			t.Fatalf("normalizeReturnedProfileAssetURL(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestPlayerProfileAssetConstants(t *testing.T) {
	t.Parallel()
	want := map[Kind]string{
		KindProfileBackgroundSmall: "profile_background_small", KindProfileBackgroundLarge: "profile_background_large",
		KindMiniProfileBackgroundSmall: "mini_profile_background_small", KindMiniProfileBackgroundLarge: "mini_profile_background_large",
		KindAvatarFrameSmall: "avatar_frame_small", KindAvatarFrameLarge: "avatar_frame_large",
		KindAnimatedAvatarSmall: "animated_avatar_small", KindAnimatedAvatarLarge: "animated_avatar_large",
		KindAnimatedAvatarWebM: "animated_avatar_webm", KindAnimatedAvatarMP4: "animated_avatar_mp4",
		KindAnimatedAvatarWebMSmall: "animated_avatar_webm_small", KindAnimatedAvatarMP4Small: "animated_avatar_mp4_small",
	}
	for kind, text := range want {
		if string(kind) != text {
			t.Fatalf("kind = %q, want %q", kind, text)
		}
	}
	if SourcePlayerServiceProfileItemsEquipped != "playerservice_profile_items_equipped" {
		t.Fatalf("unexpected source %q", SourcePlayerServiceProfileItemsEquipped)
	}
}

type playerProfileTestAPI struct {
	service  *playerservice.Service
	server   *httptest.Server
	mu       sync.Mutex
	requests []url.Values
}

func newPlayerProfileTestAPI(t *testing.T) *playerProfileTestAPI {
	t.Helper()
	testAPI := &playerProfileTestAPI{}
	testAPI.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/IPlayerService/GetProfileItemsEquipped/v1/":
			testAPI.mu.Lock()
			query := make(url.Values, len(r.URL.Query()))
			for key, values := range r.URL.Query() {
				query[key] = append([]string(nil), values...)
			}
			testAPI.requests = append(testAPI.requests, query)
			testAPI.mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"response":{
                    "profile_background":{"image_small":%q,"image_large":"//cdn.example.test/profile/background-large.jpg","name":"Background Name","item_title":"Background Title","appid":570},
                    "mini_profile_background":{"image_small":%q,"image_large":"relative/mini-large.jpg","item_title":"Mini Title"},
                    "avatar_frame":{"image_small":%q,"image_large":%q,"name":"Frame Name","item_title":""},
                    "animated_avatar":{"image_small":%q,"image_large":%q,"movie_webm":%q,"movie_mp4":%q,"movie_webm_small":%q,"movie_mp4_small":%q,"item_title":"Animated Title"}
                }}`,
				testAPI.server.URL+"/profile/background-small.jpg",
				testAPI.server.URL+"/profile/mini-small.jpg",
				testAPI.server.URL+"/profile/frame-small.png",
				testAPI.server.URL+"/profile/frame-large.png",
				testAPI.server.URL+"/profile/animated-small",
				testAPI.server.URL+"/profile/animated-large.gif",
				testAPI.server.URL+"/profile/animated.webm",
				testAPI.server.URL+"/profile/animated.mp4",
				testAPI.server.URL+"/profile/animated-small.webm",
				testAPI.server.URL+"/profile/animated-small.mp4",
			)
		default:
			if !strings.HasPrefix(r.URL.Path, "/profile/") {
				http.NotFound(w, r)
				return
			}
			body := "profile:" + r.URL.Path
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Length", fmt.Sprint(len(body)))
			if r.Method != http.MethodHead {
				_, _ = io.WriteString(w, body)
			}
		}
	}))
	t.Cleanup(testAPI.server.Close)
	client, err := steam.NewClient(steam.WithBaseURL(testAPI.server.URL))
	if err != nil {
		t.Fatalf("steam.NewClient returned error: %v", err)
	}
	t.Cleanup(client.Close)
	testAPI.service = client.API.PlayerService
	return testAPI
}

func (testAPI *playerProfileTestAPI) queries(t *testing.T) []url.Values {
	t.Helper()
	testAPI.mu.Lock()
	defer testAPI.mu.Unlock()
	queries := make([]url.Values, len(testAPI.requests))
	copy(queries, testAPI.requests)
	return queries
}
