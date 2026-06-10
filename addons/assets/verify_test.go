package assets

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVerifyURLs(t *testing.T) {
	server := newAssetTestServer(t)

	results, err := VerifyURLs(context.Background(), server.URL+"/header.jpg", server.URL+"/missing.jpg")
	if err != nil {
		t.Fatalf("VerifyURLs returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("VerifyURLs returned %d results, want 2", len(results))
	}
	if !results[0].Exists || results[0].StatusCode != http.StatusOK || results[0].ContentType != "image/jpeg" {
		t.Fatalf("first verify result = %#v", results[0])
	}
	if results[1].Exists || results[1].StatusCode != http.StatusNotFound {
		t.Fatalf("second verify result = %#v", results[1])
	}
}

func TestVerifyURLsFallbackToGET(t *testing.T) {
	server := newAssetTestServer(t)

	results, err := VerifyURLs(context.Background(), server.URL+"/fallback.jpg")
	if err != nil {
		t.Fatalf("VerifyURLs returned error: %v", err)
	}
	if len(results) != 1 || !results[0].Exists || results[0].StatusCode != http.StatusOK {
		t.Fatalf("fallback result = %#v", results)
	}
}

func TestVerifyAppAssets(t *testing.T) {
	server := newAssetTestServer(t)
	client := &http.Client{Transport: rewriteHostTransport{base: server.URL}}

	results, err := VerifyAppAssets(context.Background(), VerifyAppOptions{
		Kinds:      []Kind{KindHeader, KindLibraryHero},
		HTTPClient: client,
	}, 550)
	if err != nil {
		t.Fatalf("VerifyAppAssets returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("VerifyAppAssets returned %d results, want 2", len(results))
	}
	if results[0].AppID != 550 || results[0].Kind != KindHeader || !results[0].Exists {
		t.Fatalf("first result = %#v", results[0])
	}
	if results[1].AppID != 550 || results[1].Kind != KindLibraryHero || results[1].Exists {
		t.Fatalf("second result = %#v", results[1])
	}
}

func TestVerifyURLsCanceledBeforeLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results, err := VerifyURLsWithOptions(ctx, VerifyOptions{}, "https://example.com/header.jpg")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("VerifyURLsWithOptions error = %v, want context canceled", err)
	}
	if len(results) != 0 {
		t.Fatalf("results = %d, want 0", len(results))
	}
}

func TestVerifyAppAssetsCanceledBeforeLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results, err := VerifyAppAssets(ctx, VerifyAppOptions{Kinds: []Kind{KindHeader}}, 550)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("VerifyAppAssets error = %v, want context canceled", err)
	}
	if len(results) != 0 {
		t.Fatalf("results = %d, want 0", len(results))
	}
}

func newAssetTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/header.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Content-Length", "11")
			if r.Method != http.MethodHead {
				_, _ = io.WriteString(w, "header-body")
			}
		case "/same/header.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = io.WriteString(w, "same-header")
		case "/other/header.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = io.WriteString(w, "other-header")
		case "/fallback.jpg":
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = io.WriteString(w, "fallback-body")
		case "/store_item_assets/steam/apps/550/header.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = io.WriteString(w, "header-body")
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

type rewriteHostTransport struct {
	base string
}

func (t rewriteHostTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	baseReq, err := http.NewRequestWithContext(req.Context(), req.Method, t.base+req.URL.Path, nil)
	if err != nil {
		return nil, err
	}
	cloned.URL = baseReq.URL
	return http.DefaultTransport.RoundTrip(cloned)
}

func stringResponse(req *http.Request, status int, contentType, body string) *http.Response {
	return &http.Response{
		StatusCode:    status,
		Status:        http.StatusText(status),
		Header:        http.Header{"Content-Type": []string{contentType}, "Content-Length": []string{"11"}},
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: int64(len(body)),
		Request:       req,
	}
}
