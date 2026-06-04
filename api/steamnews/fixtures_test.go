package steamnews

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFixtureGetNewsForAppDecodes(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "fixtures", "official", "ISteamNews", "GetNewsForApp", "v2", "public.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var resp GetNewsForAppResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	if resp.AppNews.AppID != 550 || resp.AppNews.Count != 1 || len(resp.AppNews.NewsItems) != 1 {
		t.Fatalf("unexpected news fixture: %#v", resp.AppNews)
	}
}
