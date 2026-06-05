package market

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFixturePriceOverviewDecodes(t *testing.T) {
	t.Parallel()

	var resp PriceOverviewResponse
	readFixtureJSON(t, "app_440_key.json", &resp)
	if !resp.Success || resp.LowestPrice != "$2.30" || resp.Volume != "1,234" {
		t.Fatalf("unexpected price overview fixture: %#v", resp)
	}
}

func TestFixturePriceOverviewSuccessFalseDecodes(t *testing.T) {
	t.Parallel()

	var resp PriceOverviewResponse
	readFixtureJSON(t, "app_440_key_success_false.json", &resp)
	if resp.Success || resp.LowestPrice != "" || resp.Volume != "" || resp.MedianPrice != "" {
		t.Fatalf("unexpected success=false price overview fixture: %#v", resp)
	}
}

func readFixtureJSON(t *testing.T, name string, v any) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "fixtures", "web", "market", "GetPriceOverview", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
}
