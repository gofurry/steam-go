package assets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofurry/steam-go/web/storefront"
)

func TestFixtureStoreMediaExtractsURLs(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "fixtures", "addons", "assets", "store_media_app_550.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var envelope storefront.AppDetailsEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	items := storeMediaItems(550, envelope["550"].Data, []Kind{
		KindStoreBackground,
		KindScreenshotThumbnail,
		KindMovieWebM480,
	})
	if got, want := len(items), 4; got != want {
		t.Fatalf("media item count = %d, want %d: %#v", got, want, items)
	}
	if items[0].Kind != KindStoreBackground || items[1].ID != 1 || items[3].Name != "Fixture Trailer" {
		t.Fatalf("unexpected media items: %#v", items)
	}
}
