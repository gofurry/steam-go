package steamuserstats

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFixtureGlobalAchievementPercentagesDecodes(t *testing.T) {
	t.Parallel()

	var resp GetGlobalAchievementPercentagesForAppResponse
	readFixtureJSON(t, filepath.Join("official", "ISteamUserStats", "GetGlobalAchievementPercentagesForApp", "v2", "public.json"), &resp)
	if got := resp.AchievementPercentages.Achievements[0].Name; got != "ACH_FIXTURE_START" {
		t.Fatalf("achievement name = %q", got)
	}
}

func TestFixtureNumberOfCurrentPlayersDecodes(t *testing.T) {
	t.Parallel()

	var resp GetNumberOfCurrentPlayersResponse
	readFixtureJSON(t, filepath.Join("official", "ISteamUserStats", "GetNumberOfCurrentPlayers", "v1", "public.json"), &resp)
	if resp.Response.Result != 1 || resp.Response.PlayerCount != 12345 {
		t.Fatalf("unexpected current players fixture: %#v", resp.Response)
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
