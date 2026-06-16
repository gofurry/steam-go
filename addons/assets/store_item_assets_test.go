package assets

import (
	"context"
	"encoding/json"
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

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/api/storebrowseservice"
	"github.com/gofurry/steam-go/internal/request"
)

func TestFetchStoreItemAssetURLs(t *testing.T) {
	t.Parallel()

	body := readStoreItemAssetFixture(t)
	transport := &storeItemAssetsTransport{body: body}
	service := newStoreItemAssetsTestService(t, transport)

	items, err := FetchStoreItemAssetURLs(context.Background(), service, StoreItemAssetOptions{
		CountryCode: " US ",
		Language:    " english ",
		Kinds: []Kind{
			KindLibraryLogo,
			KindLibraryLogo2x,
			KindHeader2x,
			KindCommunityIconJPG,
			KindPageBackground,
			KindLibraryHero2x,
		},
	}, 4710650)
	if err != nil {
		t.Fatalf("FetchStoreItemAssetURLs returned error: %v", err)
	}
	if len(items) != 5 {
		t.Fatalf("item count = %d, want 5: %#v", len(items), items)
	}
	if items[0].Kind != KindLibraryLogo || items[0].Filename != "logo.png" || items[0].Source != SourceStoreBrowse {
		t.Fatalf("logo item = %#v", items[0])
	}
	if items[1].Kind != KindLibraryLogo2x || items[1].Filename != "logo_2x.png" {
		t.Fatalf("logo_2x item = %#v", items[1])
	}
	if items[2].Kind != KindHeader2x || !strings.Contains(items[2].URL, "?t=1781233831") {
		t.Fatalf("header_2x item = %#v", items[2])
	}
	if items[3].Kind != KindCommunityIconJPG || items[3].Digest != "09afddbb4f0012de04d82dd4719626e46ecaed08" || items[3].Filename != "community_icon.jpg" {
		t.Fatalf("community icon item = %#v", items[3])
	}
	if items[4].Kind != KindLibraryHero2x || items[4].Digest != "448851b668e4397d9863e571cf481b0e46e1315f" {
		t.Fatalf("library hero item = %#v", items[4])
	}

	req := transport.onlyRequest(t)
	if req.method != http.MethodGet || req.path != "/IStoreBrowseService/GetItems/v1/" {
		t.Fatalf("unexpected request: %#v", req)
	}
	var input storebrowseservice.GetItemsRequest
	if err := json.Unmarshal([]byte(req.query.Get("input_json")), &input); err != nil {
		t.Fatalf("decode input_json: %v", err)
	}
	if len(input.IDs) != 1 || input.IDs[0].AppID != 4710650 {
		t.Fatalf("unexpected ids: %#v", input.IDs)
	}
	if input.Context == nil || input.Context.CountryCode != "US" || input.Context.Language != "english" {
		t.Fatalf("unexpected context: %#v", input.Context)
	}
	if input.DataRequest == nil || !input.DataRequest.IncludeAssets {
		t.Fatalf("unexpected data_request: %#v", input.DataRequest)
	}
}

func TestFetchStoreItemAssetURLsStripQuery(t *testing.T) {
	t.Parallel()

	service := newStoreItemAssetsTestService(t, &storeItemAssetsTransport{body: readStoreItemAssetFixture(t)})
	items, err := FetchStoreItemAssetURLs(context.Background(), service, StoreItemAssetOptions{
		Kinds:      []Kind{KindHeader2x},
		StripQuery: true,
	}, 4710650)
	if err != nil {
		t.Fatalf("FetchStoreItemAssetURLs returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("item count = %d, want 1", len(items))
	}
	if strings.Contains(items[0].URL, "?") {
		t.Fatalf("query was not stripped: %s", items[0].URL)
	}
}

func TestResolveStoreItemAssetURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		baseURL  string
		format   string
		filename string
		want     string
	}{
		{
			name:     "filename placeholder",
			baseURL:  "https://cdn.test/store_item_assets/",
			format:   "steam/apps/1/${FILENAME}?t=1",
			filename: "abc/header.jpg",
			want:     "https://cdn.test/store_item_assets/steam/apps/1/abc/header.jpg?t=1",
		},
		{
			name:     "sprintf placeholder",
			baseURL:  "https://cdn.test/store_item_assets",
			format:   "steam/apps/1/%s",
			filename: "/abc/header.jpg",
			want:     "https://cdn.test/store_item_assets/steam/apps/1/abc/header.jpg",
		},
		{
			name:     "absolute filename",
			baseURL:  "https://cdn.test/store_item_assets/",
			filename: "https://other.test/asset.jpg",
			want:     "https://other.test/asset.jpg",
		},
		{
			name:     "relative path without format",
			baseURL:  "https://cdn.test/store_item_assets/",
			filename: "steam/apps/1/abc/header.jpg",
			want:     "https://cdn.test/store_item_assets/steam/apps/1/abc/header.jpg",
		},
	}
	for _, tt := range tests {
		got := ResolveStoreItemAssetURL(tt.baseURL, tt.format, tt.filename)
		if got != tt.want {
			t.Fatalf("%s: got %q want %q", tt.name, got, tt.want)
		}
	}
}

func TestParseStoreItemAssetURL(t *testing.T) {
	t.Parallel()

	digest, filename := ParseStoreItemAssetURL("https://shared.steamstatic.com/store_item_assets/steam/apps/4710650/448851b668e4397d9863e571cf481b0e46e1315f/library_hero_2x.jpg?t=1781233831")
	if digest != "448851b668e4397d9863e571cf481b0e46e1315f" || filename != "library_hero_2x.jpg" {
		t.Fatalf("digest=%q filename=%q", digest, filename)
	}
}

func TestVerifyReadAndDownloadStoreItemAssets(t *testing.T) {
	assetServer := newStoreItemAssetServer(t)
	body := strings.ReplaceAll(readStoreItemAssetFixture(t), "steam/apps/4710650/${FILENAME}?t=1781233831", assetServer.URL+"/store_item_assets/steam/apps/4710650/${FILENAME}")
	service := newStoreItemAssetsTestService(t, &storeItemAssetsTransport{body: body})

	verifyResults, err := VerifyStoreItemAssets(context.Background(), service, VerifyStoreItemAssetOptions{
		Kinds: []Kind{KindHeader2x},
	}, 4710650)
	if err != nil {
		t.Fatalf("VerifyStoreItemAssets returned error: %v", err)
	}
	if len(verifyResults) != 1 || !verifyResults[0].Exists || verifyResults[0].Digest == "" || verifyResults[0].Source != SourceStoreBrowse {
		t.Fatalf("verify results = %#v", verifyResults)
	}

	readResults, err := ReadStoreItemAssets(context.Background(), service, ReadStoreItemAssetOptions{
		Kinds: []Kind{KindHeader2x},
	}, 4710650)
	if err != nil {
		t.Fatalf("ReadStoreItemAssets returned error: %v", err)
	}
	if len(readResults) != 1 || string(readResults[0].Data) != "asset:header_2x.jpg" || readResults[0].Filename != "header_2x.jpg" {
		t.Fatalf("read results = %#v", readResults)
	}

	dir := t.TempDir()
	downloadResults, err := DownloadStoreItemAssets(context.Background(), service, DownloadStoreItemAssetOptions{
		Dir:           dir,
		Kinds:         []Kind{KindHeader2x, KindLibraryHero2x},
		Mode:          StoreByAppID,
		FilenameStyle: FilenameKind,
	}, 4710650)
	if err != nil {
		t.Fatalf("DownloadStoreItemAssets returned error: %v", err)
	}
	if len(downloadResults) != 2 {
		t.Fatalf("download result count = %d, want 2", len(downloadResults))
	}
	assertFile(t, filepath.Join(dir, "4710650", "header_2x.jpg"), "asset:header_2x.jpg")
	assertFile(t, filepath.Join(dir, "4710650", "library_hero_2x.jpg"), "asset:library_hero_2x.jpg")
}

func TestFixtureStoreItemAssetsExtractsURLs(t *testing.T) {
	t.Parallel()

	var resp storebrowseservice.GetItemsResponse
	if err := json.Unmarshal([]byte(readStoreItemAssetFixture(t)), &resp); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	if len(resp.Response.StoreItems) != 1 {
		t.Fatalf("store item count = %d, want 1", len(resp.Response.StoreItems))
	}
	items := storeItemAssetItems(4710650, resp.Response.StoreItems[0], []Kind{
		KindHeader2x,
		KindLibraryHero2x,
		KindCommunityIconJPG,
	}, DefaultStoreItemAssetBaseURL, false)
	if len(items) != 3 {
		t.Fatalf("item count = %d, want 3: %#v", len(items), items)
	}
	if items[0].Filename != "header_2x.jpg" || items[1].Digest == "" || items[2].Kind != KindCommunityIconJPG {
		t.Fatalf("unexpected items: %#v", items)
	}
}

func TestLiveFetchStoreItemAssetURLs(t *testing.T) {
	if os.Getenv("STEAM_GO_LIVE") != "1" {
		t.Skip("set STEAM_GO_LIVE=1 to run live Steam asset discovery")
	}

	client, err := steam.NewClient(steam.WithSafeDefaults())
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	items, err := FetchStoreItemAssetURLs(context.Background(), client.API.StoreBrowseService, StoreItemAssetOptions{
		CountryCode: "US",
		Language:    "english",
		Kinds: []Kind{
			KindHeader,
			KindHeader2x,
			KindLibraryCapsule,
			KindLibraryCapsule2x,
			KindLibraryHero,
			KindLibraryHero2x,
		},
	}, 4710650)
	if err != nil {
		t.Fatalf("FetchStoreItemAssetURLs returned error: %v", err)
	}

	seen := make(map[Kind]URLItem, len(items))
	for _, item := range items {
		seen[item.Kind] = item
	}
	for _, kind := range []Kind{KindHeader, KindHeader2x, KindLibraryCapsule, KindLibraryCapsule2x, KindLibraryHero, KindLibraryHero2x} {
		item, ok := seen[kind]
		if !ok {
			t.Fatalf("missing live asset kind %s in %#v", kind, items)
		}
		if item.URL == "" || item.Digest == "" || item.Filename == "" {
			t.Fatalf("incomplete live asset item for %s: %#v", kind, item)
		}
	}
}

func readStoreItemAssetFixture(t *testing.T) string {
	t.Helper()

	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "fixtures", "addons", "assets", "store_item_assets_app_4710650.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return string(body)
}

func newStoreItemAssetServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := filepath.Base(r.URL.Path)
		w.Header().Set("Content-Type", "image/test")
		w.Header().Set("Content-Length", strconv.Itoa(len("asset:"+name)))
		if r.Method != http.MethodHead {
			_, _ = io.WriteString(w, "asset:"+name)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func newStoreItemAssetsTestService(t *testing.T, transport *storeItemAssetsTransport) *storebrowseservice.Service {
	t.Helper()

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		64*1024,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    transport,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}
	return storebrowseservice.NewService(executor)
}

type storeItemAssetsTransport struct {
	mu       sync.Mutex
	requests []storeItemCapturedRequest
	body     string
}

type storeItemCapturedRequest struct {
	method string
	path   string
	query  url.Values
}

func (t *storeItemAssetsTransport) Do(_ context.Context, req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	clonedQuery := make(url.Values, len(req.URL.Query()))
	for key, values := range req.URL.Query() {
		copied := make([]string, len(values))
		copy(copied, values)
		clonedQuery[key] = copied
	}
	t.requests = append(t.requests, storeItemCapturedRequest{
		method: req.Method,
		path:   req.URL.Path,
		query:  clonedQuery,
	})
	t.mu.Unlock()

	return &http.Response{
		StatusCode:    http.StatusOK,
		Header:        http.Header{"Content-Type": []string{"application/json"}, "Content-Length": []string{strconv.Itoa(len(t.body))}},
		Body:          io.NopCloser(strings.NewReader(t.body)),
		ContentLength: int64(len(t.body)),
		Request:       req,
	}, nil
}

func (t *storeItemAssetsTransport) onlyRequest(tb testing.TB) storeItemCapturedRequest {
	tb.Helper()
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.requests) != 1 {
		tb.Fatalf("expected one request, got %d", len(t.requests))
	}
	return t.requests[0]
}
