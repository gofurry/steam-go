package storefront

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFixtureAppDetailsDecodesAndRawSubtreesAreJSON(t *testing.T) {
	t.Parallel()

	var envelope AppDetailsEnvelope
	readFixtureJSON(t, filepath.Join("web", "storefront", "GetAppDetails", "app_550_en.json"), &envelope)
	app := envelope["550"].Data
	if app.Name != "Left 4 Dead 2 Fixture" || app.SteamAppID != 550 {
		t.Fatalf("unexpected app details fixture: %#v", app)
	}
	assertRawJSON(t, "package_groups", app.PackageGroups)
	assertRawJSON(t, "ratings", app.Ratings)
	requiredAge, ok, err := app.SteamGermanyRequiredAge()
	if err != nil {
		t.Fatalf("SteamGermanyRequiredAge returned error: %v", err)
	}
	if !ok || requiredAge != "18" {
		t.Fatalf("unexpected steam germany required age: %q ok=%v", requiredAge, ok)
	}
}

func TestFixtureAppDetailsRegionMissingFieldsDecodes(t *testing.T) {
	t.Parallel()

	var envelope AppDetailsEnvelope
	readFixtureJSON(t, filepath.Join("web", "storefront", "GetAppDetails", "app_550_region_missing.json"), &envelope)
	app := envelope["550"].Data
	if app.Name != "Left 4 Dead 2 Region Fixture" || app.PriceOverview == nil {
		t.Fatalf("unexpected sparse app details fixture: %#v", app)
	}
	if app.PriceOverview.Currency != "JPY" || app.PCRequirements == nil || app.PCRequirements.Minimum != "" {
		t.Fatalf("unexpected regional fields: %#v", app)
	}
	assertRawJSON(t, "package_groups", app.PackageGroups)
	assertRawJSON(t, "ratings", app.Ratings)
}

func TestFixturePackageDetailsDecodesAndRawSubtreeIsJSON(t *testing.T) {
	t.Parallel()

	var envelope PackageDetailsEnvelope
	readFixtureJSON(t, filepath.Join("web", "storefront", "GetPackageDetails", "package_469_en.json"), &envelope)
	pkg := envelope["469"].Data
	if pkg.PackageID != 469 || len(pkg.Apps) != 1 {
		t.Fatalf("unexpected package details fixture: %#v", pkg)
	}
	assertRawJSON(t, "details", pkg.Details)
}

func TestFixtureAppReviewsDecodes(t *testing.T) {
	t.Parallel()

	var resp AppReviewsResponse
	readFixtureJSON(t, filepath.Join("web", "storefront", "GetAppReviews", "app_550_en.json"), &resp)
	if resp.QuerySummary.TotalReviews != 12 || resp.Reviews[0].WeightedVoteScore.Float64() != 0.5 {
		t.Fatalf("unexpected reviews fixture: %#v", resp)
	}
}

func TestFixtureAppReviewsCursorPagesDecode(t *testing.T) {
	t.Parallel()

	var page1 AppReviewsResponse
	readFixtureJSON(t, filepath.Join("web", "storefront", "GetAppReviews", "app_550_cursor_page_1.json"), &page1)
	if page1.Cursor != "AoIIPw%3D%3D" || len(page1.Reviews) != 2 {
		t.Fatalf("unexpected reviews cursor page 1: %#v", page1)
	}
	if page1.Reviews[0].WeightedVoteScore.Float64() != 0.75 || page1.Reviews[1].WeightedVoteScore.Float64() != 0.5 {
		t.Fatalf("unexpected weighted scores: %#v", page1.Reviews)
	}

	var page2 AppReviewsResponse
	readFixtureJSON(t, filepath.Join("web", "storefront", "GetAppReviews", "app_550_cursor_page_2.json"), &page2)
	if page2.Cursor != page1.Cursor || len(page2.Reviews) != 0 {
		t.Fatalf("unexpected reviews cursor page 2: %#v", page2)
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

func assertRawJSON(t *testing.T, name string, raw json.RawMessage) {
	t.Helper()
	if len(raw) == 0 {
		t.Fatalf("%s raw JSON is empty", name)
	}
	if !json.Valid(raw) {
		t.Fatalf("%s raw JSON is invalid: %s", name, string(raw))
	}
}
