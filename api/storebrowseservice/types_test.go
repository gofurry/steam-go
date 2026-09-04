package storebrowseservice

import (
	"encoding/json"
	"testing"
)

func TestStoreItemAssetsUnmarshalJSON(t *testing.T) {
	t.Parallel()

	var assets StoreItemAssets
	err := json.Unmarshal([]byte(`{
		"library_capsule": "library_600x900.jpg",
		"library_capsule_2x": "library_600x900_2x.jpg",
		"last_modified": 1744981707
	}`), &assets)
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if got := assets["library_capsule"]; got != "library_600x900.jpg" {
		t.Fatalf("library_capsule = %q", got)
	}
	if got := assets["library_capsule_2x"]; got != "library_600x900_2x.jpg" {
		t.Fatalf("library_capsule_2x = %q", got)
	}
	if _, ok := assets["last_modified"]; ok {
		t.Fatalf("last_modified should not be retained: %#v", assets)
	}
}

func TestStoreItemAssetsUnmarshalJSONIgnoresNonStringFields(t *testing.T) {
	t.Parallel()

	var assets StoreItemAssets
	err := json.Unmarshal([]byte(`{
		"header": "header.jpg",
		"number": 1744981707,
		"bool": true,
		"null": null,
		"object": {"foo": "bar"},
		"array": ["future", "metadata"]
	}`), &assets)
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if got := assets["header"]; got != "header.jpg" {
		t.Fatalf("header = %q", got)
	}
	if len(assets) != 1 {
		t.Fatalf("assets = %#v, want only the string field", assets)
	}
}

func TestStoreItemAssetsUnmarshalJSONRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	var assets StoreItemAssets
	err := assets.UnmarshalJSON([]byte(`{
		"library_capsule": "library_600x900.jpg",
		"future_metadata": {"foo": }
	}`))
	if err == nil {
		t.Fatal("expected malformed JSON error")
	}
}
