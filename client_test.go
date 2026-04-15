package steam_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	steam "github.com/GoFurry/steam-go"
	"github.com/GoFurry/steam-go/api/playerservice"
	"github.com/GoFurry/steam-go/api/steamnews"
	"github.com/GoFurry/steam-go/api/steamuserstats"
)

func TestNewClientRequiresAPIKey(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient()
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client")
	}
}

func TestSteamUserGetPlayerSummaries(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("key"); got != "test-key" {
			t.Fatalf("unexpected api key: %s", got)
		}
		if got := r.URL.Query().Get("access_token"); got != "" {
			t.Fatalf("unexpected access token: %s", got)
		}
		if got := r.URL.Query().Get("steamids"); got != "1,2" {
			t.Fatalf("unexpected steamids: %s", got)
		}
		_, _ = w.Write([]byte(`{"response":{"players":[{"steamid":"1","personaname":"one"},{"steamid":"2","personaname":"two"}]}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	resp, err := client.SteamUser.GetPlayerSummaries(context.Background(), []string{"1", "2"})
	if err != nil {
		t.Fatalf("GetPlayerSummaries returned error: %v", err)
	}
	if len(resp.Response.Players) != 2 {
		t.Fatalf("unexpected player count: %d", len(resp.Response.Players))
	}
	if resp.Response.Players[0].PersonaName != "one" {
		t.Fatalf("unexpected player name: %s", resp.Response.Players[0].PersonaName)
	}
}

func TestSteamUserGetPlayerSummariesWithoutAPIKey(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("key"); got != "" {
			t.Fatalf("unexpected api key: %s", got)
		}
		_, _ = w.Write([]byte(`{"response":{"players":[{"steamid":"1","personaname":"one"}]}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(steam.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	resp, err := client.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	if err != nil {
		t.Fatalf("GetPlayerSummaries returned error: %v", err)
	}
	if len(resp.Response.Players) != 1 {
		t.Fatalf("unexpected player count: %d", len(resp.Response.Players))
	}
}

func TestSteamUserGetPlayerSummariesWithAccessToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("key"); got != "test-key" {
			t.Fatalf("unexpected api key: %s", got)
		}
		if got := r.URL.Query().Get("access_token"); got != "access-token-a" {
			t.Fatalf("unexpected access token: %s", got)
		}
		_, _ = w.Write([]byte(`{"response":{"players":[{"steamid":"1","personaname":"one"}]}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithBaseURL(server.URL),
		steam.WithAPIKey("test-key"),
		steam.WithAccessToken("access-token-a"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	if err != nil {
		t.Fatalf("GetPlayerSummaries returned error: %v", err)
	}
}

func TestSteamUserGetPlayerSummariesWithRotatingAPIKeys(t *testing.T) {
	t.Parallel()

	var seen []string
	var idx atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := idx.Add(1)
		key := r.URL.Query().Get("key")
		seen = append(seen, key)
		_, _ = w.Write([]byte(`{"response":{"players":[{"steamid":"1","personaname":"one"}]}}`))
		if current > 2 {
			t.Fatalf("unexpected request count: %d", current)
		}
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithBaseURL(server.URL),
		steam.WithAPIKeys("key-a", "key-b"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	for i := 0; i < 2; i++ {
		_, err = client.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
		if err != nil {
			t.Fatalf("GetPlayerSummaries returned error: %v", err)
		}
	}

	if len(seen) != 2 {
		t.Fatalf("unexpected seen count: %d", len(seen))
	}
	if seen[0] != "key-a" || seen[1] != "key-b" {
		t.Fatalf("unexpected key rotation order: %#v", seen)
	}
}

func TestSteamUserGetPlayerSummariesWithRotatingAccessTokens(t *testing.T) {
	t.Parallel()

	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.URL.Query().Get("access_token"))
		_, _ = w.Write([]byte(`{"response":{"players":[{"steamid":"1","personaname":"one"}]}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithBaseURL(server.URL),
		steam.WithAccessTokens("token-a", "token-b"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	for i := 0; i < 2; i++ {
		_, err = client.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
		if err != nil {
			t.Fatalf("GetPlayerSummaries returned error: %v", err)
		}
	}

	if len(seen) != 2 {
		t.Fatalf("unexpected seen count: %d", len(seen))
	}
	if seen[0] != "token-a" || seen[1] != "token-b" {
		t.Fatalf("unexpected access token rotation order: %#v", seen)
	}
}

func TestSteamUserGetPlayerSummariesRaw(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"response":{"players":[]}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	body, err := client.SteamUser.GetPlayerSummariesRaw(context.Background(), []string{"1"})
	if err != nil {
		t.Fatalf("GetPlayerSummariesRaw returned error: %v", err)
	}
	if string(body) != `{"response":{"players":[]}}` {
		t.Fatalf("unexpected body: %s", string(body))
	}
}

func TestSteamUserValidation(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient(steam.WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.SteamUser.GetPlayerSummaries(context.Background(), nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	tooMany := make([]string, 101)
	for i := range tooMany {
		tooMany[i] = "1"
	}
	_, err = client.SteamUser.GetPlayerSummaries(context.Background(), tooMany)
	expectKind(t, err, steam.ErrorKindRequestBuild)
}

func TestPlayerServiceOptions(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if got := query.Get("include_appinfo"); got != "1" {
			t.Fatalf("unexpected include_appinfo: %s", got)
		}
		if got := query.Get("include_played_free_games"); got != "1" {
			t.Fatalf("unexpected include_played_free_games: %s", got)
		}
		filters := query["appids_filter"]
		if len(filters) != 2 || filters[0] != "570" || filters[1] != "730" {
			t.Fatalf("unexpected appids_filter: %#v", filters)
		}
		_, _ = w.Write([]byte(`{"response":{"game_count":1,"games":[{"appid":570,"name":"Dota 2"}]}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	resp, err := client.PlayerService.GetOwnedGames(
		context.Background(),
		"76561197960435530",
		&playerservice.GetOwnedGamesOptions{
			IncludePlayedFreeGames: true,
			AppIDsFilter:           []uint32{570, 730},
		},
	)
	if err != nil {
		t.Fatalf("GetOwnedGames returned error: %v", err)
	}
	if resp.Response.GameCount != 1 {
		t.Fatalf("unexpected game count: %d", resp.Response.GameCount)
	}
	if resp.Response.Games[0].AppID != 570 {
		t.Fatalf("unexpected app id: %d", resp.Response.Games[0].AppID)
	}
}

func TestSteamNewsOptions(t *testing.T) {
	t.Parallel()

	endDate := time.Unix(1_700_000_000, 0).UTC()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if got := query.Get("appid"); got != "570" {
			t.Fatalf("unexpected appid: %s", got)
		}
		if got := query.Get("maxlength"); got != "200" {
			t.Fatalf("unexpected maxlength: %s", got)
		}
		if got := query.Get("count"); got != "3" {
			t.Fatalf("unexpected count: %s", got)
		}
		if got := query.Get("enddate"); got != "1700000000" {
			t.Fatalf("unexpected enddate: %s", got)
		}
		if got := query.Get("feeds"); got != "steam_community_announcements,steam_blog" {
			t.Fatalf("unexpected feeds: %s", got)
		}
		_, _ = w.Write([]byte(`{"appnews":{"appid":570,"newsitems":[{"gid":"1","title":"update"}],"count":1}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	resp, err := client.SteamNews.GetNewsForApp(
		context.Background(),
		570,
		&steamnews.GetNewsForAppOptions{
			MaxLength: 200,
			EndDate:   endDate,
			Count:     3,
			Feeds:     []string{"steam_community_announcements", "steam_blog"},
		},
	)
	if err != nil {
		t.Fatalf("GetNewsForApp returned error: %v", err)
	}
	if resp.AppNews.Count != 1 {
		t.Fatalf("unexpected news count: %d", resp.AppNews.Count)
	}
}

func TestSteamUserStatsGetPlayerAchievements(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if got := query.Get("steamid"); got != "76561197960435530" {
			t.Fatalf("unexpected steamid: %s", got)
		}
		if got := query.Get("appid"); got != "550" {
			t.Fatalf("unexpected appid: %s", got)
		}
		if got := query.Get("l"); got != "en" {
			t.Fatalf("unexpected language: %s", got)
		}
		_, _ = w.Write([]byte(`{"playerstats":{"steamID":"76561197960435530","gameName":"Left 4 Dead 2","achievements":[{"apiname":"ACH_WIN_ONE_GAME","achieved":1,"unlocktime":123,"name":"Winner","description":"Win one game"}],"success":true}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	resp, err := client.SteamUserStats.GetPlayerAchievements(
		context.Background(),
		"76561197960435530",
		550,
		&steamuserstats.GetPlayerAchievementsOptions{Language: "en"},
	)
	if err != nil {
		t.Fatalf("GetPlayerAchievements returned error: %v", err)
	}
	if resp.PlayerStats.GameName != "Left 4 Dead 2" {
		t.Fatalf("unexpected game name: %s", resp.PlayerStats.GameName)
	}
	if len(resp.PlayerStats.Achievements) != 1 {
		t.Fatalf("unexpected achievement count: %d", len(resp.PlayerStats.Achievements))
	}
}

func TestSteamUserStatsGetPlayerAchievementsRaw(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("l"); got != "" {
			t.Fatalf("unexpected language: %s", got)
		}
		_, _ = w.Write([]byte(`{"playerstats":{"steamID":"1","gameName":"Game","achievements":[],"success":true}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	body, err := client.SteamUserStats.GetPlayerAchievementsRaw(context.Background(), "1", 550, nil)
	if err != nil {
		t.Fatalf("GetPlayerAchievementsRaw returned error: %v", err)
	}
	if string(body) != `{"playerstats":{"steamID":"1","gameName":"Game","achievements":[],"success":true}}` {
		t.Fatalf("unexpected body: %s", string(body))
	}
}

func TestSteamUserStatsValidation(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient(steam.WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.SteamUserStats.GetPlayerAchievements(context.Background(), "", 550, nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.SteamUserStats.GetPlayerAchievements(context.Background(), "1", 0, nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)
}

func TestSteamUserStatsAPIResponseError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"playerstats":{"steamID":"1","gameName":"Game","achievements":[],"success":false}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.SteamUserStats.GetPlayerAchievements(context.Background(), "1", 550, nil)
	expectKind(t, err, steam.ErrorKindAPIResponse)
}

func TestHTTPStatusError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	expectKind(t, err, steam.ErrorKindHTTPStatus)
}

func TestDecodeError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"response":`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	expectKind(t, err, steam.ErrorKindDecode)
}

func TestRetryOnServerError(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			http.Error(w, "retry", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"response":{"players":[]}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithAPIKey("test-key"),
		steam.WithBaseURL(server.URL),
		steam.WithRetry(1),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	if err != nil {
		t.Fatalf("GetPlayerSummaries returned error: %v", err)
	}
	if attempts.Load() != 2 {
		t.Fatalf("unexpected attempts: %d", attempts.Load())
	}
}

func TestRetryOnUnauthorizedWithKeyFailover(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.URL.Query().Get("key"))
		if attempts.Add(1) == 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"response":{"players":[]}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithBaseURL(server.URL),
		steam.WithAPIKeys("key-a", "key-b"),
		steam.WithRetry(1),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	if err != nil {
		t.Fatalf("GetPlayerSummaries returned error: %v", err)
	}
	if attempts.Load() != 2 {
		t.Fatalf("unexpected attempts: %d", attempts.Load())
	}
	if len(seen) != 2 || seen[0] != "key-a" || seen[1] != "key-b" {
		t.Fatalf("unexpected key rotation: %#v", seen)
	}
}

func TestRetryOnRateLimitWithKeyFailover(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.URL.Query().Get("key"))
		if attempts.Add(1) == 1 {
			http.Error(w, "rate limit", http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"response":{"players":[]}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithBaseURL(server.URL),
		steam.WithAPIKeys("key-a", "key-b"),
		steam.WithRetry(1),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	if err != nil {
		t.Fatalf("GetPlayerSummaries returned error: %v", err)
	}
	if attempts.Load() != 2 {
		t.Fatalf("unexpected attempts: %d", attempts.Load())
	}
	if len(seen) != 2 || seen[0] != "key-a" || seen[1] != "key-b" {
		t.Fatalf("unexpected key rotation: %#v", seen)
	}
}

func TestNoKeyProviderDoesNotRetryOnRateLimit(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		http.Error(w, "rate limit", http.StatusTooManyRequests)
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithBaseURL(server.URL),
		steam.WithRetry(1),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	expectKind(t, err, steam.ErrorKindHTTPStatus)
	if attempts.Load() != 1 {
		t.Fatalf("unexpected attempts: %d", attempts.Load())
	}
}

func TestContextTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte(`{"response":{"players":[]}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithAPIKey("test-key"),
		steam.WithBaseURL(server.URL),
		steam.WithTimeout(50*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	expectKind(t, err, steam.ErrorKindTransport)
}

func TestProxySelector(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"response":{"players":[]}}`))
	}))
	defer server.Close()

	selector := &stubProxySelector{}
	client, err := steam.NewClient(
		steam.WithAPIKey("test-key"),
		steam.WithBaseURL(server.URL),
		steam.WithProxySelector(selector),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	if err != nil {
		t.Fatalf("GetPlayerSummaries returned error: %v", err)
	}
	if selector.calls.Load() == 0 {
		t.Fatal("expected proxy selector to be called")
	}
}

type stubProxySelector struct {
	calls atomic.Int32
}

func (s *stubProxySelector) Next(*http.Request) (*url.URL, error) {
	s.calls.Add(1)
	return nil, nil
}

func newTestClient(t *testing.T, baseURL string) *steam.Client {
	t.Helper()

	client, err := steam.NewClient(
		steam.WithAPIKey("test-key"),
		steam.WithBaseURL(baseURL),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	return client
}

func expectKind(t *testing.T, err error, kind steam.ErrorKind) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *steam.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Kind != kind {
		t.Fatalf("unexpected error kind: got %s want %s", apiErr.Kind, kind)
	}
}
