package market

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFixturePriceOverviewDecodes(t *testing.T) {
	t.Parallel()

	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "fixtures", "web", "market", "GetPriceOverview", "app_440_key.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var resp PriceOverviewResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	if !resp.Success || resp.LowestPrice != "$2.30" || resp.Volume != "1,234" {
		t.Fatalf("unexpected price overview fixture: %#v", resp)
	}
}
