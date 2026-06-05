package community

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFixtureInventoryPaginationDecodes(t *testing.T) {
	t.Parallel()

	var page1 InventoryResponse
	readFixtureJSON(t, filepath.Join("web", "community", "GetInventory", "page_1.json"), &page1)
	if len(page1.Assets) != 1 || len(page1.Descriptions) != 1 || !page1.MoreItems.Bool() || page1.LastAssetID != "101" {
		t.Fatalf("unexpected inventory page 1 fixture: %#v", page1)
	}
	assertRawMessagesJSON(t, "descriptions", page1.Descriptions[0].Descriptions)
	assertRawMessagesJSON(t, "actions", page1.Descriptions[0].Actions)

	var page2 InventoryResponse
	readFixtureJSON(t, filepath.Join("web", "community", "GetInventory", "page_2.json"), &page2)
	if len(page2.Assets) != 1 || page2.MoreItems.Bool() || page2.LastAssetID != "101" {
		t.Fatalf("unexpected inventory page 2 fixture: %#v", page2)
	}
}

func readFixtureJSON(t *testing.T, name string, v any) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "fixtures", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
}

func assertRawMessagesJSON(t *testing.T, name string, values []json.RawMessage) {
	t.Helper()
	if len(values) == 0 {
		t.Fatalf("%s raw JSON slice is empty", name)
	}
	for _, raw := range values {
		if !json.Valid(raw) {
			t.Fatalf("%s raw JSON is invalid: %s", name, string(raw))
		}
	}
}
