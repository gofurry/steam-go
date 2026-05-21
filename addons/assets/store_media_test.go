package assets

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/gofurry/steam-go/internal/request"
	itraffic "github.com/gofurry/steam-go/internal/traffic"
	"github.com/gofurry/steam-go/web/storefront"
)

func TestFetchStoreMediaURLs(t *testing.T) {
	mediaServer := newStoreMediaAssetServer(t)
	service := newStoreMediaTestService(t, storeMediaAppDetails(mediaServer.URL))

	items, err := FetchStoreMediaURLs(context.Background(), service, StoreMediaOptions{
		Kinds: []Kind{KindScreenshotFull, KindMovieThumbnail, KindMovieDASHH264, KindMovieHLSH264, KindStoreBackground},
	}, 550)
	if err != nil {
		t.Fatalf("FetchStoreMediaURLs returned error: %v", err)
	}
	if len(items) != 9 {
		t.Fatalf("items = %d, want 9: %#v", len(items), items)
	}
	if items[0].Kind != KindScreenshotFull || items[0].ID != 0 || !strings.HasSuffix(items[0].URL, "/ss0_full.jpg") {
		t.Fatalf("first screenshot = %#v", items[0])
	}
	if items[2].Kind != KindMovieThumbnail || items[2].ID != 5952 || items[2].Name != "Trailer One" {
		t.Fatalf("movie thumbnail = %#v", items[2])
	}
	if items[8].Kind != KindStoreBackground || !strings.HasSuffix(items[8].URL, "/background.jpg") {
		t.Fatalf("background = %#v", items[8])
	}
}

func TestVerifyAndDownloadStoreMedia(t *testing.T) {
	mediaServer := newStoreMediaAssetServer(t)
	service := newStoreMediaTestService(t, storeMediaAppDetails(mediaServer.URL))

	verifyResults, err := VerifyStoreMedia(context.Background(), service, VerifyStoreMediaOptions{
		Kinds: []Kind{KindScreenshotThumbnail, KindMovieDASHH264},
	}, 550)
	if err != nil {
		t.Fatalf("VerifyStoreMedia returned error: %v", err)
	}
	if len(verifyResults) != 4 {
		t.Fatalf("verify results = %d, want 4", len(verifyResults))
	}
	for _, result := range verifyResults {
		if !result.Exists {
			t.Fatalf("verify result should exist: %#v", result)
		}
	}

	dir := t.TempDir()
	downloadResults, err := DownloadStoreMedia(context.Background(), service, DownloadStoreMediaOptions{
		Dir:           dir,
		Kinds:         []Kind{KindMovieDASHH264},
		Mode:          StoreByAppID,
		FilenameStyle: FilenameKind,
		Concurrency:   2,
	}, 550)
	if err != nil {
		t.Fatalf("DownloadStoreMedia returned error: %v", err)
	}
	if len(downloadResults) != 2 {
		t.Fatalf("download results = %d, want 2", len(downloadResults))
	}
	assertFile(t, filepath.Join(dir, "550", "movie_dash_h264_5952.mpd"), "media:h264_5952.mpd")
	assertFile(t, filepath.Join(dir, "550", "movie_dash_h264_5489.mpd"), "media:h264_5489.mpd")
}

func TestDownloadStoreMediaOriginalNamesAreUnique(t *testing.T) {
	mediaServer := newStoreMediaAssetServer(t)
	service := newStoreMediaTestService(t, storeMediaAppDetails(mediaServer.URL))
	dir := t.TempDir()

	results, err := DownloadStoreMedia(context.Background(), service, DownloadStoreMediaOptions{
		Dir:   dir,
		Kinds: []Kind{KindMovieHLSH264},
		Mode:  StoreByAppID,
	}, 550)
	if err != nil {
		t.Fatalf("DownloadStoreMedia returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	if results[0].Path == results[1].Path {
		t.Fatalf("duplicate paths: %#v", results)
	}
	if _, err := os.Stat(results[0].Path); err != nil {
		t.Fatalf("first path missing: %v", err)
	}
	if _, err := os.Stat(results[1].Path); err != nil {
		t.Fatalf("second path missing: %v", err)
	}
}

func newStoreMediaAssetServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := filepath.Base(r.URL.Path)
		w.Header().Set("Content-Type", "asset/test")
		if r.Method != http.MethodHead {
			_, _ = io.WriteString(w, "media:"+name)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func storeMediaAppDetails(baseURL string) string {
	return `{"550":{"success":true,"data":{"steam_appid":550,"name":"Left 4 Dead 2","background":"` + baseURL + `/background.jpg","background_raw":"` + baseURL + `/background_raw.jpg","screenshots":[{"id":0,"path_thumbnail":"` + baseURL + `/ss0_thumb.jpg","path_full":"` + baseURL + `/ss0_full.jpg"},{"id":1,"path_thumbnail":"` + baseURL + `/ss1_thumb.jpg","path_full":"` + baseURL + `/ss1_full.jpg"}],"movies":[{"id":5952,"name":"Trailer One","thumbnail":"` + baseURL + `/movie_5952.jpg","dash_h264":"` + baseURL + `/h264_5952.mpd","hls_h264":"` + baseURL + `/master.m3u8"},{"id":5489,"name":"Trailer Two","thumbnail":"` + baseURL + `/movie_5489.jpg","dash_h264":"` + baseURL + `/h264_5489.mpd","hls_h264":"` + baseURL + `/master.m3u8"}]}}}`
}

func newStoreMediaTestService(t *testing.T, body string) *storefront.Service {
	t.Helper()
	transport := &storeMediaTransport{body: body}
	executor, err := request.NewExecutor(
		"https://store.steampowered.com",
		nil,
		nil,
		64*1024,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    transport,
		},
		map[itraffic.Class]request.ExecutionPolicy{
			itraffic.ClassPublicStorePage: {
				Retry:        0,
				RetryBackoff: request.DefaultRetryBackoffConfig(),
				Transport:    transport,
			},
		},
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}
	return storefront.NewService(executor)
}

type storeMediaTransport struct {
	mu       sync.Mutex
	requests []url.Values
	body     string
}

func (t *storeMediaTransport) Do(_ context.Context, req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	t.requests = append(t.requests, req.URL.Query())
	t.mu.Unlock()

	return &http.Response{
		StatusCode:    http.StatusOK,
		Header:        http.Header{"Content-Type": []string{"application/json"}, "Content-Length": []string{strconv.Itoa(len(t.body))}},
		Body:          io.NopCloser(strings.NewReader(t.body)),
		ContentLength: int64(len(t.body)),
		Request:       req,
	}, nil
}
