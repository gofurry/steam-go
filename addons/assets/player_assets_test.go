package assets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/api/steamuser"
)

const (
	playerOneSteamID = "76561198000000001"
	playerTwoSteamID = "76561198000000002"
)

func TestFetchPlayerAvatarURLsPreservesInputAndKindOrder(t *testing.T) {
	service, baseURL := newPlayerAvatarTestService(t)
	items, err := FetchPlayerAvatarURLs(context.Background(), service, PlayerAvatarOptions{
		Kinds: []Kind{KindPlayerAvatarFull, KindPlayerAvatar},
	}, " "+playerOneSteamID+" ", playerTwoSteamID)
	if err != nil {
		t.Fatalf("FetchPlayerAvatarURLs returned error: %v", err)
	}
	want := []URLItem{
		{SteamID: playerOneSteamID, Kind: KindPlayerAvatarFull, URL: baseURL + "/avatars/one-full.jpg", Source: SourceSteamUserPlayerSummaries},
		{SteamID: playerOneSteamID, Kind: KindPlayerAvatar, URL: baseURL + "/avatars/one-small.png", Source: SourceSteamUserPlayerSummaries},
		{SteamID: playerTwoSteamID, Kind: KindPlayerAvatarFull, URL: baseURL + "/avatars/two-full", Source: SourceSteamUserPlayerSummaries},
		{SteamID: playerTwoSteamID, Kind: KindPlayerAvatar, URL: baseURL + "/avatars/two-small.png", Source: SourceSteamUserPlayerSummaries},
	}
	if !reflect.DeepEqual(items, want) {
		t.Fatalf("items =\n%#v\nwant\n%#v", items, want)
	}
}

func TestFetchPlayerAvatarURLsDefaultsAndValidation(t *testing.T) {
	service, _ := newPlayerAvatarTestService(t)
	items, err := FetchPlayerAvatarURLs(context.Background(), service, PlayerAvatarOptions{}, playerOneSteamID, playerTwoSteamID)
	if err != nil {
		t.Fatalf("FetchPlayerAvatarURLs returned error: %v", err)
	}
	if len(items) != 5 {
		t.Fatalf("items = %d, want 5 after skipping one empty URL", len(items))
	}
	if items[0].Kind != KindPlayerAvatar || items[1].Kind != KindPlayerAvatarMedium || items[2].Kind != KindPlayerAvatarFull {
		t.Fatalf("unexpected default kind order: %#v", items[:3])
	}
	if _, err := FetchPlayerAvatarURLs(context.Background(), service, PlayerAvatarOptions{Kinds: []Kind{KindHeader}}, playerOneSteamID); err == nil {
		t.Fatal("expected invalid avatar kind error")
	}
	if _, err := FetchPlayerAvatarURLs(context.Background(), nil, PlayerAvatarOptions{}, playerOneSteamID); err == nil {
		t.Fatal("expected nil service error")
	}
	empty, err := FetchPlayerAvatarURLs(context.Background(), service, PlayerAvatarOptions{})
	if err != nil || len(empty) != 0 || empty == nil {
		t.Fatalf("empty result = %#v, %v", empty, err)
	}
}

func TestPlayerAvatarVerifyReadAndDownload(t *testing.T) {
	service, _ := newPlayerAvatarTestService(t)

	verified, err := VerifyPlayerAvatars(context.Background(), service, VerifyPlayerAvatarOptions{
		Kinds: []Kind{KindPlayerAvatar},
	}, playerOneSteamID, playerTwoSteamID)
	if err != nil {
		t.Fatalf("VerifyPlayerAvatars returned error: %v", err)
	}
	if len(verified) != 2 || !verified[0].Exists || verified[0].SteamID != playerOneSteamID || verified[1].SteamID != playerTwoSteamID {
		t.Fatalf("unexpected verify results: %#v", verified)
	}

	read, err := ReadPlayerAvatars(context.Background(), service, ReadPlayerAvatarOptions{
		Kinds:       []Kind{KindPlayerAvatarMedium},
		MaxBytes:    1024,
		Concurrency: 2,
	}, playerOneSteamID)
	if err != nil {
		t.Fatalf("ReadPlayerAvatars returned error: %v", err)
	}
	if len(read) != 1 || read[0].SteamID != playerOneSteamID || string(read[0].Data) != "asset:/avatars/one-medium.jpg" {
		t.Fatalf("unexpected read results: %#v", read)
	}

	dir := t.TempDir()
	downloaded, err := DownloadPlayerAvatars(context.Background(), service, DownloadPlayerAvatarOptions{
		Dir:         dir,
		Kinds:       []Kind{KindPlayerAvatar, KindPlayerAvatarFull},
		Concurrency: 2,
	}, playerOneSteamID, playerTwoSteamID)
	if err != nil {
		t.Fatalf("DownloadPlayerAvatars returned error: %v", err)
	}
	if len(downloaded) != 4 {
		t.Fatalf("downloads = %d, want 4", len(downloaded))
	}
	assertFile(t, filepath.Join(dir, playerOneSteamID, "player_avatar.png"), "asset:/avatars/one-small.png")
	assertFile(t, filepath.Join(dir, playerTwoSteamID, "player_avatar_full.jpg"), "asset:/avatars/two-full")
	for _, result := range downloaded {
		if result.SteamID == "" || result.Source != SourceSteamUserPlayerSummaries {
			t.Fatalf("player metadata lost in download: %#v", result)
		}
	}
	if _, err := DownloadPlayerAvatars(context.Background(), service, DownloadPlayerAvatarOptions{}, playerOneSteamID); err == nil {
		t.Fatal("expected empty download dir error")
	}
}

func TestSteamIDMetadataPropagatesAndOmitsCleanly(t *testing.T) {
	server := newAssetTestServer(t)
	item := URLItem{
		SteamID: playerOneSteamID,
		Kind:    KindPlayerAvatar,
		URL:     server.URL + "/header.jpg",
		Source:  SourceSteamUserPlayerSummaries,
	}

	verified, err := verifyURLItem(context.Background(), nil, item)
	if err != nil || verified.SteamID != playerOneSteamID {
		t.Fatalf("verify metadata = %#v, %v", verified, err)
	}
	read, err := readURLItems(context.Background(), nil, 1024, 1, []URLItem{item})
	if err != nil || len(read) != 1 || read[0].SteamID != playerOneSteamID {
		t.Fatalf("read metadata = %#v, %v", read, err)
	}

	request := downloadRequest{item: item, url: item.URL, path: filepath.Join(t.TempDir(), "avatar.jpg")}
	downloaded, err := downloadRequests(context.Background(), nil, OverwriteAlways, 1, []downloadRequest{request})
	if err != nil || len(downloaded) != 1 || downloaded[0].SteamID != playerOneSteamID {
		t.Fatalf("download metadata = %#v, %v", downloaded, err)
	}

	canceledContext, cancel := context.WithCancel(context.Background())
	cancel()
	canceledReads, err := readURLItems(canceledContext, nil, 1024, 1, []URLItem{item})
	if err == nil || canceledReads[0].SteamID != playerOneSteamID {
		t.Fatalf("canceled read metadata = %#v, %v", canceledReads, err)
	}
	canceledDownloads, err := downloadRequests(canceledContext, nil, OverwriteAlways, 1, []downloadRequest{request})
	if err == nil || canceledDownloads[0].SteamID != playerOneSteamID {
		t.Fatalf("canceled download metadata = %#v, %v", canceledDownloads, err)
	}

	appJSON, err := json.Marshal(URLItem{AppID: 550, URL: "https://example.test/header.jpg"})
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(appJSON, []byte("steam_id")) {
		t.Fatalf("empty steam_id was not omitted: %s", appJSON)
	}
	playerJSON, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(playerJSON, []byte(`"steam_id":"`+playerOneSteamID+`"`)) {
		t.Fatalf("steam_id missing from JSON: %s", playerJSON)
	}
}

func TestPlayerAvatarConstants(t *testing.T) {
	t.Parallel()
	if KindPlayerAvatar != "player_avatar" || KindPlayerAvatarMedium != "player_avatar_medium" || KindPlayerAvatarFull != "player_avatar_full" {
		t.Fatal("unexpected player avatar kind constants")
	}
	if SourceSteamUserPlayerSummaries != "steamuser_player_summaries" {
		t.Fatalf("unexpected player avatar source %q", SourceSteamUserPlayerSummaries)
	}
}

func newPlayerAvatarTestService(t *testing.T) (*steamuser.Service, string) {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ISteamUser/GetPlayerSummaries/v2/":
			if got := r.URL.Query().Get("steamids"); got == "" {
				t.Error("steamids query is empty")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"response":{"players":[
                    {"steamid":%q,"avatar":%q,"avatarmedium":"","avatarfull":%q,"avatarhash":"do-not-use"},
                    {"steamid":%q,"avatar":%q,"avatarmedium":%q,"avatarfull":%q,"avatarhash":"do-not-use"}
                ]}}`,
				playerTwoSteamID, server.URL+"/avatars/two-small.png", server.URL+"/avatars/two-full",
				playerOneSteamID, server.URL+"/avatars/one-small.png", server.URL+"/avatars/one-medium.jpg", server.URL+"/avatars/one-full.jpg",
			)
		default:
			if !strings.HasPrefix(r.URL.Path, "/avatars/") {
				http.NotFound(w, r)
				return
			}
			body := "asset:" + r.URL.Path
			w.Header().Set("Content-Type", "image/test")
			w.Header().Set("Content-Length", fmt.Sprint(len(body)))
			if r.Method != http.MethodHead {
				_, _ = io.WriteString(w, body)
			}
		}
	}))
	t.Cleanup(server.Close)

	client, err := steam.NewClient(steam.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("steam.NewClient returned error: %v", err)
	}
	t.Cleanup(client.Close)
	return client.API.SteamUser, server.URL
}
