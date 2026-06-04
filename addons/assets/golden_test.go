package assets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAssetURLGolden(t *testing.T) {
	t.Parallel()

	got := []URLItem{
		{AppID: 550, Kind: KindHeader, URL: HeaderURLs(550)[0]},
		{AppID: 550, Kind: KindHeaderLocalized, URL: HeaderLocalizedURLs("schinese", 550)[0]},
		{AppID: 550, Kind: KindLibraryHero, URL: LibraryHeroURLs(550)[0]},
		{AppID: 550, Kind: KindCommunityIconJPG, URL: CommunityIconURL(550, "abcdef0123456789")},
	}
	gotJSON, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatalf("marshal golden output: %v", err)
	}
	gotJSON = append(gotJSON, '\n')

	want, err := os.ReadFile(filepath.Join("..", "..", "testdata", "golden", "assets_urls.json"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if string(gotJSON) != string(want) {
		t.Fatalf("asset URL golden mismatch\ngot:\n%s\nwant:\n%s", gotJSON, want)
	}
}
