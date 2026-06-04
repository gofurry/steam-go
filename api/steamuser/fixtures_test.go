package steamuser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFixtureGetPlayerSummariesDecodes(t *testing.T) {
	t.Parallel()

	var resp GetPlayerSummariesResponse
	readFixtureJSON(t, filepath.Join("official", "ISteamUser", "GetPlayerSummaries", "v2", "public.json"), &resp)
	if got := len(resp.Response.Players); got != 1 {
		t.Fatalf("players = %d, want 1", got)
	}
	if got := resp.Response.Players[0].SteamID; got != "76561198000000001" {
		t.Fatalf("SteamID = %q", got)
	}
}

func TestFixtureGetFriendListDecodes(t *testing.T) {
	t.Parallel()

	var resp GetFriendListResponse
	readFixtureJSON(t, filepath.Join("official", "ISteamUser", "GetFriendList", "v1", "public.json"), &resp)
	if got := resp.FriendsList.Friends[0].Relationship; got != "friend" {
		t.Fatalf("relationship = %q", got)
	}
}

func TestFixtureGetPlayerBansDecodes(t *testing.T) {
	t.Parallel()

	var resp GetPlayerBansResponse
	readFixtureJSON(t, filepath.Join("official", "ISteamUser", "GetPlayerBans", "v1", "public.json"), &resp)
	if len(resp.Players) != 1 || resp.Players[0].EconomyBan != "none" {
		t.Fatalf("unexpected player bans fixture: %#v", resp.Players)
	}
}

func TestFixtureGetUserGroupListDecodes(t *testing.T) {
	t.Parallel()

	var resp GetUserGroupListResponse
	readFixtureJSON(t, filepath.Join("official", "ISteamUser", "GetUserGroupList", "v1", "public.json"), &resp)
	if !resp.Response.Success || len(resp.Response.Groups) != 1 {
		t.Fatalf("unexpected group list fixture: %#v", resp.Response)
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
