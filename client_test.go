package steam_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	steam "github.com/GoFurry/steam-go"
	"github.com/GoFurry/steam-go/api/accountcartservice"
	"github.com/GoFurry/steam-go/api/familygroupsservice"
	"github.com/GoFurry/steam-go/api/loyaltyrewardsservice"
	"github.com/GoFurry/steam-go/api/newsservice"
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
	if client.API == nil {
		t.Fatal("expected api entrypoint")
	}
	if client.API.AccountCartService == nil || client.API.BillingService == nil || client.API.CommunityService == nil {
		t.Fatal("expected new core services to be initialized")
	}
	if client.API.FamilyGroupsService == nil || client.API.LoyaltyRewardsService == nil {
		t.Fatal("expected access-token services to be initialized")
	}
	if client.API.MobileNotificationService == nil || client.API.NewsService == nil {
		t.Fatal("expected additional service groups to be initialized")
	}
}

func TestAccountCartServiceGetCart(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/IAccountCartService/GetCart/v1/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("access_token"); got != "access-token-a" {
			t.Fatalf("unexpected access token: %s", got)
		}
		if got := r.URL.Query().Get("user_country"); got != "gb" {
			t.Fatalf("unexpected user_country: %s", got)
		}
		_, _ = w.Write([]byte(`{"response":{"cart":{"line_items":[{"line_item_id":"1","type":1,"packageid":123,"is_valid":true,"time_added":1700000000,"price_when_added":{"amount_in_cents":"999","currency_code":1,"formatted_amount":"$9.99"},"flags":{"is_gift":false,"is_private":true}}],"subtotal":{"amount_in_cents":"999","currency_code":1,"formatted_amount":"$9.99"},"is_valid":true}}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithBaseURL(server.URL),
		steam.WithAccessToken("access-token-a"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	resp, err := client.API.AccountCartService.GetCart(context.Background(), &accountcartservice.GetCartOptions{UserCountry: "gb"})
	if err != nil {
		t.Fatalf("GetCart returned error: %v", err)
	}
	if len(resp.Response.Cart.LineItems) != 1 {
		t.Fatalf("unexpected line items: %d", len(resp.Response.Cart.LineItems))
	}
	if resp.Response.Cart.Subtotal.AmountInCents != "999" {
		t.Fatalf("unexpected subtotal: %s", resp.Response.Cart.Subtotal.AmountInCents)
	}
}

func TestAccountCartServiceDeleteCart(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/IAccountCartService/DeleteCart/v1/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"response":{}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithBaseURL(server.URL),
		steam.WithAccessToken("access-token-a"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.API.AccountCartService.DeleteCart(context.Background())
	if err != nil {
		t.Fatalf("DeleteCart returned error: %v", err)
	}
}

func TestBillingServiceGetRecurringSubscriptionsCount(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("access_token"); got != "access-token-a" {
			t.Fatalf("unexpected access token: %s", got)
		}
		_, _ = w.Write([]byte(`{"response":{"active_subscriptions_count":2,"inactive_subscriptions_count":1}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithBaseURL(server.URL),
		steam.WithAccessToken("access-token-a"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	resp, err := client.API.BillingService.GetRecurringSubscriptionsCount(context.Background())
	if err != nil {
		t.Fatalf("GetRecurringSubscriptionsCount returned error: %v", err)
	}
	if resp.Response.ActiveSubscriptionsCount != 2 || resp.Response.InactiveSubscriptionsCount != 1 {
		t.Fatalf("unexpected response: %#v", resp.Response)
	}
}

func TestCommunityServiceGetApps(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if got := query.Get("appids[0]"); got != "550" {
			t.Fatalf("unexpected appids[0]: %s", got)
		}
		if got := query.Get("appids[1]"); got != "570" {
			t.Fatalf("unexpected appids[1]: %s", got)
		}
		_, _ = w.Write([]byte(`{"response":{"apps":[{"appid":550,"name":"Left 4 Dead 2","icon":"iconhash","community_visible_stats":true,"propagation":"public","app_type":1,"content_descriptorids":[1],"content_descriptorids_including_dlc":[1,2]}]}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	resp, err := client.API.CommunityService.GetApps(context.Background(), []uint32{550, 570})
	if err != nil {
		t.Fatalf("GetApps returned error: %v", err)
	}
	if len(resp.Response.Apps) != 1 {
		t.Fatalf("unexpected app count: %d", len(resp.Response.Apps))
	}
	if resp.Response.Apps[0].Name != "Left 4 Dead 2" {
		t.Fatalf("unexpected app name: %s", resp.Response.Apps[0].Name)
	}
}

func TestCommunityServiceValidation(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient()
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.API.CommunityService.GetApps(context.Background(), nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.CommunityService.GetApps(context.Background(), []uint32{0})
	expectKind(t, err, steam.ErrorKindRequestBuild)
}

func TestFamilyGroupsService(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if got := query.Get("family_groupid"); got != "1136785" {
			t.Fatalf("unexpected family_groupid: %s", got)
		}

		switch r.URL.Path {
		case "/IFamilyGroupsService/GetChangeLog/v1/":
			_, _ = w.Write([]byte(`{"response":{"changes":[{"timestamp":"1700000000","actor_steamid":"1","type":1,"body":"{}","by_support":false}]}}`))
		case "/IFamilyGroupsService/GetFamilyGroup/v1/":
			_, _ = w.Write([]byte(`{"response":{"name":"Test Family","members":[{"steamid":"1","role":1,"time_joined":1700000000,"cooldown_seconds_remaining":0}],"free_spots":4,"country":"CN","slot_cooldown_remaining_seconds":0,"slot_cooldown_overrides":0}}`))
		case "/IFamilyGroupsService/GetFamilyGroupForUser/v1/":
			if got := query.Get("include_family_group_response"); got != "true" {
				t.Fatalf("unexpected include_family_group_response: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"family_groupid":"1136785","is_not_member_of_any_group":false,"latest_time_joined":1700000000,"latest_joined_family_groupid":"1136785","role":2,"cooldown_seconds_remaining":0,"family_group":{"name":"Test Family","members":[],"free_spots":4,"country":"CN","slot_cooldown_remaining_seconds":0,"slot_cooldown_overrides":0},"can_undelete_last_joined_family":false,"membership_history":[{"family_groupid":"1136785","rtime_joined":1700000000,"rtime_left":0,"role":2,"participated":true}]}}`))
		case "/IFamilyGroupsService/GetPlaytimeSummary/v1/":
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			_, _ = w.Write([]byte(`{"response":{"entries":[{"steamid":"1","appid":550,"first_played":1700000000,"latest_played":1700001000,"seconds_played":600}]}}`))
		case "/IFamilyGroupsService/GetSharedLibraryApps/v1/":
			_, _ = w.Write([]byte(`{"response":{"owner_steamid":"1","apps":[{"appid":550,"owner_steamids":["1"],"name":"Left 4 Dead 2","capsule_filename":"capsule.jpg","img_icon_hash":"iconhash","exclude_reason":0,"rt_time_acquired":1700000000,"rt_last_played":1700001000,"rt_playtime":600,"app_type":1}]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	changeLog, err := client.API.FamilyGroupsService.GetChangeLog(context.Background(), "1136785")
	if err != nil {
		t.Fatalf("GetChangeLog returned error: %v", err)
	}
	if len(changeLog.Response.Changes) != 1 {
		t.Fatalf("unexpected change count: %d", len(changeLog.Response.Changes))
	}

	familyGroup, err := client.API.FamilyGroupsService.GetFamilyGroup(context.Background(), "1136785")
	if err != nil {
		t.Fatalf("GetFamilyGroup returned error: %v", err)
	}
	if familyGroup.Response.Name != "Test Family" {
		t.Fatalf("unexpected family group name: %s", familyGroup.Response.Name)
	}

	familyForUser, err := client.API.FamilyGroupsService.GetFamilyGroupForUser(
		context.Background(),
		"1136785",
		&familygroupsservice.GetFamilyGroupForUserOptions{IncludeFamilyGroupResponse: true},
	)
	if err != nil {
		t.Fatalf("GetFamilyGroupForUser returned error: %v", err)
	}
	if familyForUser.Response.FamilyGroupID != "1136785" {
		t.Fatalf("unexpected family group id: %s", familyForUser.Response.FamilyGroupID)
	}

	playtime, err := client.API.FamilyGroupsService.GetPlaytimeSummary(context.Background(), "1136785")
	if err != nil {
		t.Fatalf("GetPlaytimeSummary returned error: %v", err)
	}
	if len(playtime.Response.Entries) != 1 {
		t.Fatalf("unexpected playtime count: %d", len(playtime.Response.Entries))
	}

	sharedApps, err := client.API.FamilyGroupsService.GetSharedLibraryApps(context.Background(), "1136785")
	if err != nil {
		t.Fatalf("GetSharedLibraryApps returned error: %v", err)
	}
	if len(sharedApps.Response.Apps) != 1 {
		t.Fatalf("unexpected shared app count: %d", len(sharedApps.Response.Apps))
	}
}

func TestFamilyGroupsServiceValidation(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient()
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.API.FamilyGroupsService.GetChangeLog(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)
}

func TestLoyaltyRewardsService(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("steamid"); got != "76561198370695025" {
			t.Fatalf("unexpected steamid: %s", got)
		}

		switch r.URL.Path {
		case "/ILoyaltyRewardsService/GetEquippedProfileItems/v1/":
			if got := r.URL.Query().Get("language"); got != "zh" {
				t.Fatalf("unexpected language: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"active_definitions":[{"appid":753,"defid":1,"type":1,"community_item_class":1,"community_item_type":1,"point_cost":"100","timestamp_created":1700000000,"timestamp_updated":1700000001,"timestamp_available":1700000002,"timestamp_available_end":1700000003,"quantity":"1","internal_description":"desc","active":true,"community_item_data":{"item_name":"name","item_title":"title","item_description":"desc","item_image_small":"small.png","item_image_large":"large.png","item_movie_webm":"movie.webm","item_movie_mp4":"movie.mp4","animated":true,"tiled":false},"usable_duration":0,"bundle_discount":0}],"inactive_definitions":[]}}`))
		case "/ILoyaltyRewardsService/GetReactionsSummaryForUser/v1/":
			_, _ = w.Write([]byte(`{"response":{"total":[{"reactionid":1,"given":2,"received":3,"points_given":"20","points_received":"30"}],"user_reviews":[],"ugc":[],"profile":[],"total_given":2,"total_received":3,"total_points_given":"20","total_points_received":"30"}}`))
		case "/ILoyaltyRewardsService/GetSummary/v1/":
			_, _ = w.Write([]byte(`{"response":{"summary":{"points":"100","points_earned":"200","points_spent":"100"},"timestamp_updated":1700000000,"auditid_highwater":"abc"}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	items, err := client.API.LoyaltyRewardsService.GetEquippedProfileItems(
		context.Background(),
		"76561198370695025",
		&loyaltyrewardsservice.GetEquippedProfileItemsOptions{Language: "zh"},
	)
	if err != nil {
		t.Fatalf("GetEquippedProfileItems returned error: %v", err)
	}
	if len(items.Response.ActiveDefinitions) != 1 {
		t.Fatalf("unexpected active definition count: %d", len(items.Response.ActiveDefinitions))
	}

	reactions, err := client.API.LoyaltyRewardsService.GetReactionsSummaryForUser(context.Background(), "76561198370695025")
	if err != nil {
		t.Fatalf("GetReactionsSummaryForUser returned error: %v", err)
	}
	if reactions.Response.TotalGiven != 2 {
		t.Fatalf("unexpected total given: %d", reactions.Response.TotalGiven)
	}

	summary, err := client.API.LoyaltyRewardsService.GetSummary(context.Background(), "76561198370695025")
	if err != nil {
		t.Fatalf("GetSummary returned error: %v", err)
	}
	if summary.Response.Summary.Points != "100" {
		t.Fatalf("unexpected points: %s", summary.Response.Summary.Points)
	}
}

func TestLoyaltyRewardsServiceValidation(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient()
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.API.LoyaltyRewardsService.GetSummary(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)
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

	resp, err := client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1", "2"})
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

	resp, err := client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
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

	_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
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
		_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
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
		_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
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
	body, err := client.API.SteamUser.GetPlayerSummariesRaw(context.Background(), []string{"1"})
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

	_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	tooMany := make([]string, 101)
	for i := range tooMany {
		tooMany[i] = "1"
	}
	_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), tooMany)
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
	resp, err := client.API.PlayerService.GetOwnedGames(
		context.Background(),
		"76561198370695025",
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

func TestMobileNotificationServiceUsesExplicitAccessToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("access_token"); got != "user-token" {
			t.Fatalf("unexpected access token: %s", got)
		}
		_, _ = w.Write([]byte(`{"response":{"notifications":[{"user_notification_type":1,"count":1}],"account_alert_count":0}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithBaseURL(server.URL),
		steam.WithAccessToken("global-token"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	resp, err := client.API.MobileNotificationService.GetUserNotificationCounts(context.Background(), "user-token")
	if err != nil {
		t.Fatalf("GetUserNotificationCounts returned error: %v", err)
	}
	if len(resp.Response.Notifications) != 1 || resp.Response.Notifications[0].Count != 1 {
		t.Fatalf("unexpected notifications: %#v", resp.Response.Notifications)
	}
}

func TestNewsServiceConvertHTMLToBBCode(t *testing.T) {
	t.Parallel()

	preserve := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if got := query.Get("content"); got != "<strong>ok</strong>" {
			t.Fatalf("unexpected content: %s", got)
		}
		if got := query.Get("preserve_newlines"); got != "false" {
			t.Fatalf("unexpected preserve_newlines: %s", got)
		}
		_, _ = w.Write([]byte(`{"response":{"converted_content":"[b]ok[/b]","found_html":true}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	resp, err := client.API.NewsService.ConvertHTMLToBBCode(
		context.Background(),
		"<strong>ok</strong>",
		&newsservice.ConvertHTMLToBBCodeOptions{PreserveNewlines: &preserve},
	)
	if err != nil {
		t.Fatalf("ConvertHTMLToBBCode returned error: %v", err)
	}
	if resp.Response.ConvertedContent != "[b]ok[/b]" || !resp.Response.FoundHTML {
		t.Fatalf("unexpected response: %#v", resp.Response)
	}
}

func TestPlayerServiceClientGetLastPlayedTimes(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if r.URL.Path != "/IPlayerService/ClientGetLastPlayedTimes/v1/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := query.Get("access_token"); got != "user-token" {
			t.Fatalf("unexpected access token: %s", got)
		}
		if got := query.Get("min_last_played"); got != "100" {
			t.Fatalf("unexpected min_last_played: %s", got)
		}
		_, _ = w.Write([]byte(`{"response":{"games":[{"appid":10,"last_playtime":1499238870,"playtime_forever":6,"first_playtime":1499238456,"playtime_disconnected":0}]}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithBaseURL(server.URL),
		steam.WithAccessToken("global-token"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	resp, err := client.API.PlayerService.ClientGetLastPlayedTimes(
		context.Background(),
		"user-token",
		&playerservice.ClientGetLastPlayedTimesOptions{MinLastPlayed: 100},
	)
	if err != nil {
		t.Fatalf("ClientGetLastPlayedTimes returned error: %v", err)
	}
	if len(resp.Response.Games) != 1 || resp.Response.Games[0].AppID != 10 {
		t.Fatalf("unexpected response: %#v", resp.Response.Games)
	}
}

func TestPlayerServiceGetAchievementsProgress(t *testing.T) {
	t.Parallel()

	includeUnvetted := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		query := r.URL.Query()
		if got := query.Get("access_token"); got != "user-token" {
			t.Fatalf("unexpected access token: %s", got)
		}
		if got := query.Get("steamid"); got != "76561198370695025" {
			t.Fatalf("unexpected steamid: %s", got)
		}
		if got := query.Get("language"); got != "zh" {
			t.Fatalf("unexpected language: %s", got)
		}
		if got := query.Get("appids[0]"); got != "550" {
			t.Fatalf("unexpected appids[0]: %s", got)
		}
		if got := query.Get("appids[1]"); got != "10" {
			t.Fatalf("unexpected appids[1]: %s", got)
		}
		if got := query.Get("include_unvetted_apps"); got != "true" {
			t.Fatalf("unexpected include_unvetted_apps: %s", got)
		}
		_, _ = w.Write([]byte(`{"response":{"achievement_progress":[{"appid":550,"unlocked":95,"total":101,"percentage":94.0594024658203,"all_unlocked":false,"cache_time":1778219448,"vetted":true}]}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithBaseURL(server.URL),
		steam.WithAccessToken("global-token"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	resp, err := client.API.PlayerService.GetAchievementsProgress(
		context.Background(),
		"user-token",
		&playerservice.GetAchievementsProgressOptions{
			SteamID:             "76561198370695025",
			Language:            "zh",
			AppIDs:              []uint32{550, 10},
			IncludeUnvettedApps: &includeUnvetted,
		},
	)
	if err != nil {
		t.Fatalf("GetAchievementsProgress returned error: %v", err)
	}
	if len(resp.Response.AchievementProgress) != 1 || resp.Response.AchievementProgress[0].AppID != 550 {
		t.Fatalf("unexpected response: %#v", resp.Response.AchievementProgress)
	}
}

func TestPlayerServiceProfileAndBadgeEndpoints(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/IPlayerService/GetAnimatedAvatar/v1/":
			if got := r.URL.Query().Get("language"); got != "en" {
				t.Fatalf("unexpected language: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"avatar":{"communityitemid":"32110862377","image_small":"items/2459330/avatar.gif","image_large":"items/2459330/avatar.jpg","name":"Animated Avatar","item_title":"Animated Avatar","item_description":"","appid":2459330,"item_type":22,"item_class":15}}}`))
		case "/IPlayerService/GetAvatarFrame/v1/":
			_, _ = w.Write([]byte(`{"response":{"avatar_frame":{"communityitemid":"32110862378","image_small":"items/2459330/frame.png","image_large":"items/2459330/frame_large.png","name":"Avatar Frame","item_title":"Avatar Frame","item_description":"","appid":2459330,"item_type":23,"item_class":14}}}`))
		case "/IPlayerService/GetMiniProfileBackground/v1/":
			_, _ = w.Write([]byte(`{"response":{"profile_background":{"communityitemid":"27764729866","image_large":"items/2459330/mini.jpg","name":"Mini Profile","item_title":"Mini Profile","item_description":"","appid":2459330,"item_type":24,"item_class":13,"movie_webm":"items/2459330/mini.webm","movie_mp4":"items/2459330/mini.mp4"}}}`))
		case "/IPlayerService/GetProfileBackground/v1/":
			_, _ = w.Write([]byte(`{"response":{"profile_background":{"communityitemid":"24662042712","image_large":"items/223850/bg.jpg","name":"Profile Background","item_title":"Profile Background","item_description":"","appid":223850,"item_type":16,"item_class":3}}}`))
		case "/IPlayerService/GetBadges/v1/":
			_, _ = w.Write([]byte(`{"response":{"badges":[{"badgeid":13,"level":120,"completion_time":1764684296,"xp":348,"scarcity":16438708}],"player_xp":1618,"player_level":13,"player_xp_needed_to_level_up":182,"player_xp_needed_current_level":1600}}`))
		case "/IPlayerService/GetCommunityBadgeProgress/v1/":
			_, _ = w.Write([]byte(`{"response":{"quests":[{"questid":115,"completed":true}]}}`))
		case "/IPlayerService/GetFavoriteBadge/v1/":
			_, _ = w.Write([]byte(`{"response":{"has_favorite_badge":true,"communityitemid":"13109033435","item_type":1,"border_color":0,"appid":926340,"level":4}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	avatar, err := client.API.PlayerService.GetAnimatedAvatar(context.Background(), "76561198856448829", &playerservice.PlayerProfileItemOptions{Language: "en"})
	if err != nil {
		t.Fatalf("GetAnimatedAvatar returned error: %v", err)
	}
	if avatar.Response.Avatar.ItemType != 22 {
		t.Fatalf("unexpected avatar item type: %d", avatar.Response.Avatar.ItemType)
	}

	frame, err := client.API.PlayerService.GetAvatarFrame(context.Background(), "76561198856448829", nil)
	if err != nil {
		t.Fatalf("GetAvatarFrame returned error: %v", err)
	}
	if frame.Response.AvatarFrame.ItemType != 23 {
		t.Fatalf("unexpected avatar frame item type: %d", frame.Response.AvatarFrame.ItemType)
	}

	miniProfile, err := client.API.PlayerService.GetMiniProfileBackground(context.Background(), "76561198856448829", nil)
	if err != nil {
		t.Fatalf("GetMiniProfileBackground returned error: %v", err)
	}
	if miniProfile.Response.ProfileBackground.MovieWebM != "items/2459330/mini.webm" {
		t.Fatalf("unexpected mini profile movie: %s", miniProfile.Response.ProfileBackground.MovieWebM)
	}

	profileBackground, err := client.API.PlayerService.GetProfileBackground(context.Background(), "76561198856448829", nil)
	if err != nil {
		t.Fatalf("GetProfileBackground returned error: %v", err)
	}
	if profileBackground.Response.ProfileBackground.ItemClass != 3 {
		t.Fatalf("unexpected profile background item class: %d", profileBackground.Response.ProfileBackground.ItemClass)
	}

	badges, err := client.API.PlayerService.GetBadges(context.Background(), "76561198856448829")
	if err != nil {
		t.Fatalf("GetBadges returned error: %v", err)
	}
	if badges.Response.PlayerLevel != 13 {
		t.Fatalf("unexpected player level: %d", badges.Response.PlayerLevel)
	}

	progress, err := client.API.PlayerService.GetCommunityBadgeProgress(context.Background(), "76561198370695025")
	if err != nil {
		t.Fatalf("GetCommunityBadgeProgress returned error: %v", err)
	}
	if len(progress.Response.Quests) != 1 || !progress.Response.Quests[0].Completed {
		t.Fatalf("unexpected quests: %#v", progress.Response.Quests)
	}

	favorite, err := client.API.PlayerService.GetFavoriteBadge(context.Background(), "76561198370695025")
	if err != nil {
		t.Fatalf("GetFavoriteBadge returned error: %v", err)
	}
	if !favorite.Response.HasFavoriteBadge || favorite.Response.Level != 4 {
		t.Fatalf("unexpected favorite badge: %#v", favorite.Response)
	}
}

func TestPlayerServiceCommunityPreferencesAndFriendsGameplayInfo(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/IPlayerService/GetCommunityPreferences/v1/":
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			if got := r.URL.Query().Get("access_token"); got != "user-token" {
				t.Fatalf("unexpected access token: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"preferences":{"parenthesize_nicknames":false,"text_filter_setting":3,"text_filter_ignore_friends":true,"text_filter_words_revision":0,"timestamp_updated":1713894485},"content_descriptor_preferences":{}}}`))
		case "/IPlayerService/GetFriendsGameplayInfo/v1/":
			if got := r.URL.Query().Get("access_token"); got != "user-token" {
				t.Fatalf("unexpected access token: %s", got)
			}
			if got := r.URL.Query().Get("appid"); got != "550" {
				t.Fatalf("unexpected appid: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"your_info":{"steamid":"76561198370695025","minutes_played":177,"minutes_played_forever":51162,"owned":true},"played_recently":[{"steamid":"76561199066283701","minutes_played":171,"minutes_played_forever":30175}],"played_ever":[{"steamid":"76561198437107999","minutes_played_forever":191}],"owns":[{"steamid":"76561198856448829"}],"in_wishlist":[{"steamid":"76561198819552989"}]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithBaseURL(server.URL),
		steam.WithAccessToken("global-token"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	prefs, err := client.API.PlayerService.GetCommunityPreferences(context.Background(), "user-token")
	if err != nil {
		t.Fatalf("GetCommunityPreferences returned error: %v", err)
	}
	if prefs.Response.Preferences.TextFilterSetting != 3 {
		t.Fatalf("unexpected preferences: %#v", prefs.Response.Preferences)
	}

	achievements, err := client.API.PlayerService.GetFriendsGameplayInfo(context.Background(), "user-token", 550)
	if err != nil {
		t.Fatalf("GetFriendsGameplayInfo returned error: %v", err)
	}
	if achievements.Response.YourInfo.SteamID != "76561198370695025" {
		t.Fatalf("unexpected your_info: %#v", achievements.Response.YourInfo)
	}
	if len(achievements.Response.PlayedEver) != 1 || achievements.Response.PlayedEver[0].MinutesPlayedForever != 191 {
		t.Fatalf("unexpected played_ever: %#v", achievements.Response.PlayedEver)
	}
}

func TestPlayerServiceNicknameLinkAndProfileInventoryEndpoints(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/IPlayerService/GetNicknameList/v1/":
			if got := r.URL.Query().Get("access_token"); got != "user-token" {
				t.Fatalf("unexpected access token: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"nicknames":[{"accountid":248405772,"nickname":"好哥哥"}]}}`))
		case "/IPlayerService/GetPlayerLinkDetails/v1/":
			query := r.URL.Query()
			if got := query.Get("steamids[0]"); got != "76561198370695025" {
				t.Fatalf("unexpected steamids[0]: %s", got)
			}
			if got := query.Get("steamids[1]"); got != "76561198856448829" {
				t.Fatalf("unexpected steamids[1]: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"accounts":[{"public_data":{"steamid":"76561198370695025","visibility_state":3,"profile_state":1,"sha_digest_avatar":"wiUiRE+2XCD7GdpK78HlgOGuq+Q=","persona_name":"百兽发布","profile_url":"ILWC","content_country_restricted":false},"private_data":{"time_created":1488556497,"last_logoff_time":1778165536,"last_seen_online":1777122496}}]}}`))
		case "/IPlayerService/GetProfileCustomization/v1/":
			query := r.URL.Query()
			if got := query.Get("include_inactive_customizations"); got != "true" {
				t.Fatalf("unexpected include_inactive_customizations: %s", got)
			}
			if got := query.Get("include_purchased_customizations"); got != "true" {
				t.Fatalf("unexpected include_purchased_customizations: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"customizations":[{"customization_type":17,"large":false,"slots":[{"slot":0,"appid":591420,"title":"11_20"}],"active":true,"customization_style":0,"purchaseid":"0","level":2}],"slots_available":11,"profile_theme":{"theme_id":"Midnight","title":"#ProfileTheme_Midnight"},"purchased_customizations":[{"purchaseid":"7442062","customization_type":24,"level":1}],"profile_preferences":{"hide_profile_awards":false}}}`))
		case "/IPlayerService/GetProfileItemsEquipped/v1/":
			_, _ = w.Write([]byte(`{"response":{"profile_background":{"communityitemid":"24662042712","image_large":"items/223850/bg.jpg","name":"Time Spy","item_title":"Time Spy","item_description":"","appid":223850,"item_type":16,"item_class":3},"mini_profile_background":{"communityitemid":"27764729866","image_large":"items/2459330/mini.jpg","name":"Mini Profile","item_title":"Mini Profile","item_description":"","appid":2459330,"item_type":24,"item_class":13,"movie_webm":"items/2459330/mini.webm","movie_mp4":"items/2459330/mini.mp4"},"avatar_frame":{"communityitemid":"32455405307","image_small":"items/1276800/frame_small.png","image_large":"items/1276800/frame_large.png","name":"Timewinder Rewind","item_title":"Timewinder Rewind","item_description":"Rewind the past, Control the future!","appid":1276800,"item_type":19,"item_class":14},"animated_avatar":{},"profile_modifier":{},"steam_deck_keyboard_skin":{}}}`))
		case "/IPlayerService/GetProfileItemsOwned/v1/":
			query := r.URL.Query()
			if got := query.Get("access_token"); got != "user-token" {
				t.Fatalf("unexpected access token: %s", got)
			}
			if got := query.Get("language"); got != "zh" {
				t.Fatalf("unexpected language: %s", got)
			}
			if got := query.Get("filters[0]"); got != "3" {
				t.Fatalf("unexpected filters[0]: %s", got)
			}
			if got := query.Get("filters[1]"); got != "13" {
				t.Fatalf("unexpected filters[1]: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"profile_backgrounds":[{"communityitemid":"24662042712","image_large":"items/223850/bg.jpg","name":"Time Spy","item_title":"Time Spy","item_description":"","appid":223850,"item_type":16,"item_class":3}],"mini_profile_backgrounds":[{"communityitemid":"27764729866","image_large":"items/2459330/mini.jpg","name":"Mini Profile","item_title":"Mini Profile","item_description":"","appid":2459330,"item_type":24,"item_class":13,"movie_webm":"items/2459330/mini.webm","movie_mp4":"items/2459330/mini.mp4"}],"avatar_frames":[{"communityitemid":"32455405307","image_small":"items/1276800/frame_small.png","image_large":"items/1276800/frame_large.png","name":"Timewinder Rewind","item_title":"Timewinder Rewind","item_description":"Rewind the past, Control the future!","appid":1276800,"item_type":19,"item_class":14}],"animated_avatars":[{"communityitemid":"27764729864","image_small":"items/2459330/avatar.gif","image_large":"items/2459330/avatar.jpg","name":"Animated Avatar","item_title":"Animated Avatar","item_description":"","appid":2459330,"item_type":22,"item_class":15}],"profile_modifiers":[{"communityitemid":"27764729863","image_small":"items/2459330/mod_small.jpg","image_large":"items/2459330/mod_large.jpg","name":"Summer in the City (Day)","item_title":"Summer in the City (Day)","item_description":"","appid":2459330,"item_type":21,"item_class":8,"profile_colors":[{"style_name":"backgroundgradient_left","color":"rgba(175, 111, 37, 1)"}]}],"steam_deck_keyboard_skins":[]}}`))
		case "/IPlayerService/GetProfileThemesAvailable/v1/":
			if got := r.URL.Query().Get("access_token"); got != "user-token" {
				t.Fatalf("unexpected access token: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"profile_themes":[{"theme_id":"","title":"#ProfileTheme_Default"},{"theme_id":"Midnight","title":"#ProfileTheme_Midnight"}]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithBaseURL(server.URL),
		steam.WithAccessToken("global-token"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	nicknames, err := client.API.PlayerService.GetNicknameList(context.Background(), "user-token")
	if err != nil {
		t.Fatalf("GetNicknameList returned error: %v", err)
	}
	if len(nicknames.Response.Nicknames) != 1 || nicknames.Response.Nicknames[0].Nickname != "好哥哥" {
		t.Fatalf("unexpected nicknames: %#v", nicknames.Response.Nicknames)
	}

	linkDetails, err := client.API.PlayerService.GetPlayerLinkDetails(context.Background(), []string{"76561198370695025", "76561198856448829"})
	if err != nil {
		t.Fatalf("GetPlayerLinkDetails returned error: %v", err)
	}
	if len(linkDetails.Response.Accounts) != 1 || linkDetails.Response.Accounts[0].PublicData.PersonaName != "百兽发布" {
		t.Fatalf("unexpected accounts: %#v", linkDetails.Response.Accounts)
	}

	customization, err := client.API.PlayerService.GetProfileCustomization(
		context.Background(),
		"76561198370695025",
		&playerservice.GetProfileCustomizationOptions{
			IncludeInactiveCustomizations:  true,
			IncludePurchasedCustomizations: true,
		},
	)
	if err != nil {
		t.Fatalf("GetProfileCustomization returned error: %v", err)
	}
	if customization.Response.ProfileTheme.ThemeID != "Midnight" || len(customization.Response.Customizations) != 1 {
		t.Fatalf("unexpected customization response: %#v", customization.Response)
	}

	equipped, err := client.API.PlayerService.GetProfileItemsEquipped(context.Background(), "76561198370695025", nil)
	if err != nil {
		t.Fatalf("GetProfileItemsEquipped returned error: %v", err)
	}
	if equipped.Response.MiniProfileBackground.MovieMP4 != "items/2459330/mini.mp4" {
		t.Fatalf("unexpected equipped mini profile movie: %s", equipped.Response.MiniProfileBackground.MovieMP4)
	}

	owned, err := client.API.PlayerService.GetProfileItemsOwned(
		context.Background(),
		"user-token",
		&playerservice.GetProfileItemsOwnedOptions{
			Language: "zh",
			Filters:  []int32{3, 13},
		},
	)
	if err != nil {
		t.Fatalf("GetProfileItemsOwned returned error: %v", err)
	}
	if len(owned.Response.ProfileModifiers) != 1 || len(owned.Response.ProfileModifiers[0].ProfileColors) != 1 {
		t.Fatalf("unexpected owned profile modifiers: %#v", owned.Response.ProfileModifiers)
	}

	themes, err := client.API.PlayerService.GetProfileThemesAvailable(context.Background(), "user-token")
	if err != nil {
		t.Fatalf("GetProfileThemesAvailable returned error: %v", err)
	}
	if len(themes.Response.ProfileThemes) != 2 || themes.Response.ProfileThemes[1].ThemeID != "Midnight" {
		t.Fatalf("unexpected themes: %#v", themes.Response.ProfileThemes)
	}
}

func TestExplicitAccessTokenValidation(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient(steam.WithAccessToken("global-token"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.API.MobileNotificationService.GetUserNotificationCounts(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.ClientGetLastPlayedTimes(context.Background(), "", nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetFriendsGameplayInfo(context.Background(), "user-token", 0)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetNicknameList(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetPlayerLinkDetails(context.Background(), nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetPlayerLinkDetails(context.Background(), []string{""})
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetProfileItemsOwned(context.Background(), "", nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetProfileThemesAvailable(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.NewsService.ConvertHTMLToBBCode(context.Background(), "", nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)
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
		_, _ = w.Write([]byte(`{"appnews":{"appid":570,"newsitems":[{"gid":"1","title":"update","tags":["patchnotes"]}],"count":1}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	resp, err := client.API.SteamNews.GetNewsForApp(
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
	if len(resp.AppNews.NewsItems) != 1 {
		t.Fatalf("unexpected news item count: %d", len(resp.AppNews.NewsItems))
	}
	if len(resp.AppNews.NewsItems[0].Tags) != 1 || resp.AppNews.NewsItems[0].Tags[0] != "patchnotes" {
		t.Fatalf("unexpected tags: %#v", resp.AppNews.NewsItems[0].Tags)
	}
}

func TestSteamUserStatsGetPlayerAchievements(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if got := query.Get("steamid"); got != "76561198370695025" {
			t.Fatalf("unexpected steamid: %s", got)
		}
		if got := query.Get("appid"); got != "550" {
			t.Fatalf("unexpected appid: %s", got)
		}
		if got := query.Get("l"); got != "en" {
			t.Fatalf("unexpected language: %s", got)
		}
		_, _ = w.Write([]byte(`{"playerstats":{"steamID":"76561198370695025","gameName":"Left 4 Dead 2","achievements":[{"apiname":"ACH_WIN_ONE_GAME","achieved":1,"unlocktime":123,"name":"Winner","description":"Win one game"}],"success":true}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	resp, err := client.API.SteamUserStats.GetPlayerAchievements(
		context.Background(),
		"76561198370695025",
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
	body, err := client.API.SteamUserStats.GetPlayerAchievementsRaw(context.Background(), "1", 550, nil)
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

	_, err = client.API.SteamUserStats.GetPlayerAchievements(context.Background(), "", 550, nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SteamUserStats.GetPlayerAchievements(context.Background(), "1", 0, nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)
}

func TestSteamUserStatsAPIResponseError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"playerstats":{"steamID":"1","gameName":"Game","achievements":[],"success":false}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.API.SteamUserStats.GetPlayerAchievements(context.Background(), "1", 550, nil)
	expectKind(t, err, steam.ErrorKindAPIResponse)
}

func TestHTTPStatusError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	expectKind(t, err, steam.ErrorKindHTTPStatus)
}

func TestDecodeError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"response":`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
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

	_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	if err != nil {
		t.Fatalf("GetPlayerSummaries returned error: %v", err)
	}
	if attempts.Load() != 2 {
		t.Fatalf("unexpected attempts: %d", attempts.Load())
	}
}

func TestRetryOnServerErrorReusesSameAPIKey(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.URL.Query().Get("key"))
		if attempts.Add(1) == 1 {
			http.Error(w, "retry", http.StatusInternalServerError)
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

	_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	if err != nil {
		t.Fatalf("GetPlayerSummaries returned error: %v", err)
	}
	if len(seen) != 2 || seen[0] != "key-a" || seen[1] != "key-a" {
		t.Fatalf("unexpected key reuse on retry: %#v", seen)
	}
}

func TestRetryOnServerErrorReusesSameAccessToken(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.URL.Query().Get("access_token"))
		if attempts.Add(1) == 1 {
			http.Error(w, "retry", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"response":{"players":[]}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithBaseURL(server.URL),
		steam.WithAccessTokens("token-a", "token-b"),
		steam.WithRetry(1),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	if err != nil {
		t.Fatalf("GetPlayerSummaries returned error: %v", err)
	}
	if len(seen) != 2 || seen[0] != "token-a" || seen[1] != "token-a" {
		t.Fatalf("unexpected access token reuse on retry: %#v", seen)
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

	_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
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

	_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
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

	_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
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

	_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
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

	_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	if err != nil {
		t.Fatalf("GetPlayerSummaries returned error: %v", err)
	}
	if selector.calls.Load() == 0 {
		t.Fatal("expected proxy selector to be called")
	}
}

func TestProxySelectorCalledPerRequest(t *testing.T) {
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

	for range 2 {
		_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
		if err != nil {
			t.Fatalf("GetPlayerSummaries returned error: %v", err)
		}
	}

	if got := selector.calls.Load(); got != 2 {
		t.Fatalf("expected proxy selector to be called twice, got %d", got)
	}
}

func TestProxySelectorErrorReturnsTransportError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"response":{"players":[]}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithAPIKey("test-key"),
		steam.WithBaseURL(server.URL),
		steam.WithProxySelector(errorProxySelector{}),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	expectKind(t, err, steam.ErrorKindTransport)
}

func TestWithHTTPClientPreservesCustomRoundTripperWithoutProxy(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient(
		steam.WithHTTPClient(&http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path != "/ISteamUser/GetPlayerSummaries/v2/" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{"response":{"players":[]}}`)),
					Header:     make(http.Header),
				}, nil
			}),
		}),
		steam.WithBaseURL("https://api.steampowered.com"),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.API.SteamUser.GetPlayerSummaries(context.Background(), []string{"1"})
	if err != nil {
		t.Fatalf("GetPlayerSummaries returned error: %v", err)
	}
}

func TestWithHTTPClientAndProxySelectorRejectsUnsupportedRoundTripper(t *testing.T) {
	t.Parallel()

	_, err := steam.NewClient(
		steam.WithHTTPClient(&http.Client{
			Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("should not be called")
			}),
		}),
		steam.WithProxySelector(errorProxySelector{}),
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWithMaxResponseBodyBytesValidation(t *testing.T) {
	t.Parallel()

	_, err := steam.NewClient(steam.WithMaxResponseBodyBytes(0))
	if err == nil {
		t.Fatal("expected error")
	}
}

type stubProxySelector struct {
	calls atomic.Int32
}

func (s *stubProxySelector) Next(*http.Request) (*url.URL, error) {
	s.calls.Add(1)
	return nil, nil
}

type errorProxySelector struct{}

func (errorProxySelector) Next(*http.Request) (*url.URL, error) {
	return nil, errors.New("proxy selector failed")
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
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
