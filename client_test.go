package steam_test

import (
	"context"
	"encoding/json"
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
	"github.com/GoFurry/steam-go/api/questservice"
	"github.com/GoFurry/steam-go/api/salefeatureservice"
	"github.com/GoFurry/steam-go/api/steamchartsservice"
	"github.com/GoFurry/steam-go/api/steamdirectory"
	"github.com/GoFurry/steam-go/api/steamnews"
	"github.com/GoFurry/steam-go/api/steamnotificationservice"
	"github.com/GoFurry/steam-go/api/steamuser"
	"github.com/GoFurry/steam-go/api/steamuserstats"
	"github.com/GoFurry/steam-go/api/storeservice"
	"github.com/GoFurry/steam-go/api/userstorevisitservice"
	"github.com/GoFurry/steam-go/api/wishlistservice"
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
	if client.API.QuestService == nil || client.API.SaleFeatureService == nil {
		t.Fatal("expected quest and sale feature services to be initialized")
	}
	if client.API.StoreBrowseService == nil || client.API.StoreCatalogService == nil {
		t.Fatal("expected store browse and catalog services to be initialized")
	}
	if client.API.StorePreferencesService == nil || client.API.StoreService == nil {
		t.Fatal("expected store preference and store services to be initialized")
	}
	if client.API.StoreTopSellersService == nil {
		t.Fatal("expected store top sellers service to be initialized")
	}
	if client.API.SteamApps == nil {
		t.Fatal("expected steam apps service to be initialized")
	}
	if client.API.SteamChartsService == nil {
		t.Fatal("expected steam charts service to be initialized")
	}
	if client.API.SteamDirectory == nil {
		t.Fatal("expected steam directory service to be initialized")
	}
	if client.API.SteamNotificationService == nil {
		t.Fatal("expected steam notification service to be initialized")
	}
	if client.API.SteamUser == nil || client.API.SteamUserStats == nil {
		t.Fatal("expected steam user services to be initialized")
	}
	if client.API.SteamUserOAuth == nil || client.API.SteamWebAPIUtil == nil {
		t.Fatal("expected steam oauth and web api util services to be initialized")
	}
	if client.API.UserAccountService == nil || client.API.UserReviewsService == nil || client.API.UserStoreVisitService == nil {
		t.Fatal("expected user account, review, and store visit services to be initialized")
	}
	if client.API.WishlistService == nil {
		t.Fatal("expected wishlist service to be initialized")
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

func TestQuestService(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/IQuestService/GetCommunityInventory/v1/":
			if got := r.URL.Query().Get("filter_appids[0]"); got != "550" {
				t.Fatalf("unexpected filter_appids[0]: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"items":[{"communityitemid":"8644390882","item_type":12,"appid":550,"owner":410429297,"attributes":[{"attributeid":4,"value":"1532415600"}],"used":false,"owner_origin":4,"amount":"1"}]}}`))
		case "/IQuestService/GetNumTradingCardsEarned/v1/":
			query := r.URL.Query()
			if got := query.Get("access_token"); got != "user-token" {
				t.Fatalf("unexpected access token: %s", got)
			}
			if got := query.Get("timestamp_start"); got != "1700000000" {
				t.Fatalf("unexpected timestamp_start: %s", got)
			}
			if got := query.Get("timestamp_end"); got != "1701000000" {
				t.Fatalf("unexpected timestamp_end: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"num_trading_cards":2}}`))
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

	inventory, err := client.API.QuestService.GetCommunityInventory(
		context.Background(),
		&questservice.GetCommunityInventoryOptions{FilterAppIDs: []uint32{550}},
	)
	if err != nil {
		t.Fatalf("GetCommunityInventory returned error: %v", err)
	}
	if len(inventory.Response.Items) != 1 || inventory.Response.Items[0].CommunityItemID != "8644390882" {
		t.Fatalf("unexpected inventory: %#v", inventory.Response.Items)
	}

	cards, err := client.API.QuestService.GetNumTradingCardsEarned(
		context.Background(),
		"user-token",
		&questservice.GetNumTradingCardsEarnedOptions{
			TimestampStart: 1700000000,
			TimestampEnd:   1701000000,
		},
	)
	if err != nil {
		t.Fatalf("GetNumTradingCardsEarned returned error: %v", err)
	}
	if cards.Response.NumTradingCards != 2 {
		t.Fatalf("unexpected num_trading_cards: %#v", cards.Response)
	}
}

func TestQuestServiceValidation(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient(steam.WithAccessToken("global-token"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.API.QuestService.GetCommunityInventory(context.Background(), &questservice.GetCommunityInventoryOptions{
		FilterAppIDs: []uint32{0},
	})
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.QuestService.GetNumTradingCardsEarned(context.Background(), "", nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)
}

func TestSaleFeatureService(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ISaleFeatureService/GetFriendsSharedYearInReview/v1/":
			query := r.URL.Query()
			if got := query.Get("steamid"); got != "76561198370695025" {
				t.Fatalf("unexpected steamid: %s", got)
			}
			if got := query.Get("year"); got != "2025" {
				t.Fatalf("unexpected year: %s", got)
			}
			if got := query.Get("return_private"); got != "false" {
				t.Fatalf("unexpected return_private: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"friend_shares":[{"steamid":"76561198337088545","privacy_state":2,"rt_privacy_updated":0,"privacy_override":false},{"steamid":"76561198856448829","privacy_state":3,"rt_privacy_updated":0,"privacy_override":false}]}}`))
		case "/ISaleFeatureService/GetUserYearAchievements/v1/":
			query := r.URL.Query()
			if got := query.Get("access_token"); got != "user-token" {
				t.Fatalf("unexpected access token: %s", got)
			}
			if got := query.Get("steamid"); got != "76561198370695025" {
				t.Fatalf("unexpected steamid: %s", got)
			}
			if got := query.Get("year"); got != "2022" {
				t.Fatalf("unexpected year: %s", got)
			}
			if got := query.Get("appids[0]"); got != "550" {
				t.Fatalf("unexpected appids[0]: %s", got)
			}
			if got := query.Get("total_only"); got != "false" {
				t.Fatalf("unexpected total_only: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"game_achievements":[{"appid":550,"achievements":[{"statid":0,"fieldid":0,"achievement_name_internal":"ACH_HONK_A_CLOWNS_NOSE"}],"all_time_unlocked_achievements":95,"unlocked_more_in_future":true}],"total_achievements":41,"total_rare_achievements":33,"total_games_with_achievements":1}}`))
		case "/ISaleFeatureService/GetUserYearInReview/v1/":
			query := r.URL.Query()
			if got := query.Get("steamid"); got != "76561198370695025" {
				t.Fatalf("unexpected steamid: %s", got)
			}
			if got := query.Get("year"); got != "2025" {
				t.Fatalf("unexpected year: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"stats":{"account_id":410429297,"year":2025,"playtime_stats":{"total_stats":{"total_playtime_seconds":1483455,"total_sessions":617,"windows_sessions":617,"total_playtime_percentagex100":10000,"windows_playtime_percentagex100":10000},"games":[{"appid":550,"stats":{"total_playtime_seconds":506679,"total_sessions":144,"windows_sessions":144,"total_playtime_percentagex100":3415,"windows_playtime_percentagex100":3415},"playtime_streak":{"longest_consecutive_days":10,"rtime_start":1764892800},"playtime_ranks":{"overall_rank":1,"windows_rank":1},"rtime_first_played":1737819327,"relative_game_stats":{"total_playtime_seconds":506679,"total_sessions":144,"windows_sessions":144,"total_playtime_percentagex100":10000,"windows_playtime_percentagex100":10000}}],"playtime_streak":{"longest_consecutive_days":15,"rtime_start":1746921600},"months":[],"game_summary":[{"appid":550,"new_this_year":false,"rtime_first_played_lifetime":1503739346,"demo":false,"playtest":false,"played_vr":false,"played_deck":false,"played_controller":false,"played_linux":false,"played_mac":false,"played_windows":true,"total_playtime_percentagex100":3415,"total_sessions":144,"rtime_release_date":1258434000}]},"demos_played":83,"playtests_played":2,"summary_stats":{"total_achievements":493,"total_games_with_achievements":48,"total_rare_achievements":32},"substantial":true,"by_numbers":{"screenshots_shared":225},"game_rankings":{"overall_ranking":{"category":"overall","rankings":[{"appid":550,"rank":1,"relative_playtime_percentagex100":520}]}}},"performance_stats":{"from_dbo":true,"overall_time_ms":"0"},"distribution":{"new_releases":14,"recent_releases":44,"classic_releases":40,"recent_cutoff_year":7}}}`))
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

	returnPrivate := false
	friendShares, err := client.API.SaleFeatureService.GetFriendsSharedYearInReview(
		context.Background(),
		"76561198370695025",
		2025,
		&salefeatureservice.GetFriendsSharedYearInReviewOptions{ReturnPrivate: &returnPrivate},
	)
	if err != nil {
		t.Fatalf("GetFriendsSharedYearInReview returned error: %v", err)
	}
	if len(friendShares.Response.FriendShares) != 2 || friendShares.Response.FriendShares[1].PrivacyState != 3 {
		t.Fatalf("unexpected friend shares: %#v", friendShares.Response.FriendShares)
	}

	totalOnly := false
	yearAchievements, err := client.API.SaleFeatureService.GetUserYearAchievements(
		context.Background(),
		"user-token",
		&salefeatureservice.GetUserYearAchievementsOptions{
			SteamID:   "76561198370695025",
			Year:      2022,
			AppIDs:    []uint32{550},
			TotalOnly: &totalOnly,
		},
	)
	if err != nil {
		t.Fatalf("GetUserYearAchievements returned error: %v", err)
	}
	if yearAchievements.Response.TotalAchievements != 41 || len(yearAchievements.Response.GameAchievements) != 1 {
		t.Fatalf("unexpected year achievements: %#v", yearAchievements.Response)
	}

	yearInReview, err := client.API.SaleFeatureService.GetUserYearInReview(
		context.Background(),
		"76561198370695025",
		2025,
	)
	if err != nil {
		t.Fatalf("GetUserYearInReview returned error: %v", err)
	}
	if yearInReview.Response.Stats.AccountID != 410429297 {
		t.Fatalf("unexpected account id: %d", yearInReview.Response.Stats.AccountID)
	}
	if len(yearInReview.Response.Stats.PlaytimeStats.Games) != 1 {
		t.Fatalf("unexpected playtime games: %#v", yearInReview.Response.Stats.PlaytimeStats.Games)
	}
	if yearInReview.Response.Distribution.RecentReleases != 44 {
		t.Fatalf("unexpected distribution: %#v", yearInReview.Response.Distribution)
	}
}

func TestSaleFeatureServiceValidation(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient(steam.WithAccessToken("global-token"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.API.SaleFeatureService.GetFriendsSharedYearInReview(context.Background(), "", 2025, nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SaleFeatureService.GetFriendsSharedYearInReview(context.Background(), "76561198370695025", 0, nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SaleFeatureService.GetUserYearAchievements(context.Background(), "", nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SaleFeatureService.GetUserYearAchievements(context.Background(), "user-token", nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SaleFeatureService.GetUserYearAchievements(context.Background(), "user-token", &salefeatureservice.GetUserYearAchievementsOptions{
		SteamID: "",
		Year:    2022,
	})
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SaleFeatureService.GetUserYearAchievements(context.Background(), "user-token", &salefeatureservice.GetUserYearAchievementsOptions{
		SteamID: "76561198370695025",
		Year:    0,
	})
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SaleFeatureService.GetUserYearAchievements(context.Background(), "user-token", &salefeatureservice.GetUserYearAchievementsOptions{
		SteamID: "76561198370695025",
		Year:    2022,
		AppIDs:  []uint32{0},
	})
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SaleFeatureService.GetUserYearInReview(context.Background(), "", 2025)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SaleFeatureService.GetUserYearInReview(context.Background(), "76561198370695025", 0)
	expectKind(t, err, steam.ErrorKindRequestBuild)
}

func TestSteamApps(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ISteamApps/GetSDRConfig/v1/":
			if got := r.URL.Query().Get("appid"); got != "550" {
				t.Fatalf("unexpected appid: %s", got)
			}
			_, _ = w.Write([]byte(`{"revision":1774928217,"pops":{"ams":{"desc":"Amsterdam (Netherlands)","geo":[4.9,52.37],"partners":1,"tier":1,"relays":[{"ipv4":"155.133.248.36","port_range":[27015,27060]}]},"eat":{"aliases":["mwh"],"desc":"Wenatchee (Washington)","geo":[-120.32,47.47],"partners":3,"tier":1}},"certs":["cert-a"],"p2p_share_ip":{"cn":20,"default":40},"relay_public_key":"relay-key","revoked_keys":["11146342570456886677"],"typical_pings":[["ams","fra",5],["eat","sea",3]],"success":true}`))
		case "/ISteamApps/GetServersAtAddress/v1/":
			if got := r.URL.Query().Get("addr"); got != "45.125.45.95" {
				t.Fatalf("unexpected addr: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"success":true,"servers":[{"addr":"45.125.45.95:28000","gmsindex":-1,"steamid":"90285522207964181","appid":550,"gamedir":"left4dead2","region":255,"secure":true,"lan":false,"gameport":28000,"specport":0},{"addr":"45.125.45.95:40000","gmsindex":-1,"steamid":"90285514591065109","appid":550,"gamedir":"left4dead2","region":255,"secure":false,"lan":false,"gameport":40000,"specport":0}]}}`))
		case "/ISteamApps/UpToDateCheck/v1/":
			query := r.URL.Query()
			if got := query.Get("appid"); got != "550" {
				t.Fatalf("unexpected appid: %s", got)
			}
			if got := query.Get("version"); got != "1" {
				t.Fatalf("unexpected version: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"success":true,"up_to_date":false,"version_is_listable":false,"required_version":2243,"message":"Your server is out of date, please update."}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	sdrConfig, err := client.API.SteamApps.GetSDRConfig(context.Background(), 550)
	if err != nil {
		t.Fatalf("GetSDRConfig returned error: %v", err)
	}
	if !sdrConfig.Success || sdrConfig.Revision != 1774928217 {
		t.Fatalf("unexpected sdr config: %#v", sdrConfig)
	}
	if sdrConfig.Pops["ams"].Relays[0].IPv4 != "155.133.248.36" {
		t.Fatalf("unexpected ams relay: %#v", sdrConfig.Pops["ams"].Relays)
	}
	if len(sdrConfig.TypicalPings) != 2 || sdrConfig.TypicalPings[0].FromPOP != "ams" || sdrConfig.TypicalPings[0].PingMS != 5 {
		t.Fatalf("unexpected typical pings: %#v", sdrConfig.TypicalPings)
	}

	serversAtAddress, err := client.API.SteamApps.GetServersAtAddress(context.Background(), "45.125.45.95")
	if err != nil {
		t.Fatalf("GetServersAtAddress returned error: %v", err)
	}
	if !serversAtAddress.Response.Success || len(serversAtAddress.Response.Servers) != 2 {
		t.Fatalf("unexpected servers at address: %#v", serversAtAddress.Response)
	}
	if serversAtAddress.Response.Servers[1].GamePort != 40000 {
		t.Fatalf("unexpected server game port: %#v", serversAtAddress.Response.Servers[1])
	}

	upToDate, err := client.API.SteamApps.UpToDateCheck(context.Background(), 550, 1)
	if err != nil {
		t.Fatalf("UpToDateCheck returned error: %v", err)
	}
	if upToDate.Response.UpToDate || upToDate.Response.RequiredVersion != 2243 {
		t.Fatalf("unexpected up-to-date response: %#v", upToDate.Response)
	}
}

func TestSteamAppsValidation(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient(steam.WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.API.SteamApps.GetSDRConfig(context.Background(), 0)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SteamApps.GetServersAtAddress(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SteamApps.UpToDateCheck(context.Background(), 0, 1)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SteamApps.UpToDateCheck(context.Background(), 550, 0)
	expectKind(t, err, steam.ErrorKindRequestBuild)
}

func TestSteamChartsService(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ISteamChartsService/GetBestOfYearPages/v1/":
			_, _ = w.Write([]byte(`{"response":{"pages":[{"name":"Best of Steam - 2024","url_path":"bestof2024","banner_url":["a.jpg",""],"banner_url_mobile":["m-a.jpg",""],"start_date":1730444400},{"name":"Best of Steam - 2023","url_path":"BestOf2023","banner_url":["b.png"],"banner_url_mobile":["m-b.jpg"],"start_date":1698822000}]}}`))
		case "/ISteamChartsService/GetGamesByConcurrentPlayers/v1/":
			_, _ = w.Write([]byte(`{"response":{"last_update":1778304619,"ranks":[{"rank":1,"appid":730,"concurrent_in_game":659296,"peak_in_game":1321704},{"rank":40,"appid":550,"concurrent_in_game":28622,"peak_in_game":36066}]}}`))
		case "/ISteamChartsService/GetMonthTopAppReleases/v1/":
			query := r.URL.Query()
			if got := query.Get("rtime_month"); got != "1746769043" {
				t.Fatalf("unexpected rtime_month: %s", got)
			}
			if got := query.Get("include_dlc"); got != "true" {
				t.Fatalf("unexpected include_dlc: %s", got)
			}
			if got := query.Get("top_results_limit"); got != "10" {
				t.Fatalf("unexpected top_results_limit: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"top_dlc_releases":[{"appid":2873470,"rtime_release":1747008000,"app_release_rank":1},{"appid":2974970,"rtime_release":1747353600,"app_release_rank":1}],"top_combined_app_and_dlc_releases":[{"appid":490110,"rtime_release":1747094400,"app_release_rank":1},{"appid":1153410,"rtime_release":1747785600,"app_release_rank":1}]}}`))
		case "/ISteamChartsService/GetMostPlayedGames/v1/":
			_, _ = w.Write([]byte(`{"response":{"rollup_date":1778198400,"ranks":[{"rank":1,"appid":730,"last_week_rank":1,"peak_in_game":1321704},{"rank":34,"appid":550,"last_week_rank":31,"peak_in_game":29429}]}}`))
		case "/ISteamChartsService/GetTopReleasesPages/v1/":
			_, _ = w.Write([]byte(`{"response":{"pages":[{"name":"Top Releases of February 2025","start_of_month":1738396800,"url_path":"top_february_2025","item_ids":[{"appid":2246340},{"appid":1771300}]},{"name":"Top Releases of January 2025","start_of_month":1738396800,"url_path":"top_january_2025","item_ids":[{"appid":2384580}]}]}}`))
		case "/ISteamChartsService/GetYearTopAppReleases/v1/":
			query := r.URL.Query()
			if got := query.Get("rtime_year"); got != "1746769043" {
				t.Fatalf("unexpected rtime_year: %s", got)
			}
			if got := query.Get("include_dlc"); got != "true" {
				t.Fatalf("unexpected include_dlc: %s", got)
			}
			if got := query.Get("top_results_limit"); got != "20" {
				t.Fatalf("unexpected top_results_limit: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"top_dlc_releases":[{"appid":3067190,"rtime_release":1762214400,"app_release_rank":1},{"appid":3080080,"rtime_release":1746489600,"app_release_rank":1}],"top_combined_app_and_dlc_releases":[{"appid":1903340,"rtime_release":1745452800,"app_release_rank":1},{"appid":1984270,"rtime_release":1759363200,"app_release_rank":1}],"top_app_list":[{"appid":730,"app_release_rank":4,"type":1},{"appid":550,"app_release_rank":4,"type":2}]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	bestOfYearPages, err := client.API.SteamChartsService.GetBestOfYearPages(context.Background())
	if err != nil {
		t.Fatalf("GetBestOfYearPages returned error: %v", err)
	}
	if len(bestOfYearPages.Response.Pages) != 2 || bestOfYearPages.Response.Pages[0].URLPath != "bestof2024" {
		t.Fatalf("unexpected best-of-year pages: %#v", bestOfYearPages.Response.Pages)
	}

	concurrentPlayers, err := client.API.SteamChartsService.GetGamesByConcurrentPlayers(context.Background())
	if err != nil {
		t.Fatalf("GetGamesByConcurrentPlayers returned error: %v", err)
	}
	if concurrentPlayers.Response.LastUpdate != 1778304619 || len(concurrentPlayers.Response.Ranks) != 2 {
		t.Fatalf("unexpected concurrent players: %#v", concurrentPlayers.Response)
	}

	rtimeMonth := uint32(1746769043)
	monthLimit := uint32(10)
	includeDLC := true
	monthTopReleases, err := client.API.SteamChartsService.GetMonthTopAppReleases(context.Background(), &steamchartsservice.GetMonthTopAppReleasesOptions{
		RTimeMonth:      &rtimeMonth,
		IncludeDLC:      &includeDLC,
		TopResultsLimit: &monthLimit,
	})
	if err != nil {
		t.Fatalf("GetMonthTopAppReleases returned error: %v", err)
	}
	if len(monthTopReleases.Response.TopDLCReleases) != 2 || monthTopReleases.Response.TopCombinedAppAndDLCReleases[0].AppID != 490110 {
		t.Fatalf("unexpected month top releases: %#v", monthTopReleases.Response)
	}

	mostPlayedGames, err := client.API.SteamChartsService.GetMostPlayedGames(context.Background())
	if err != nil {
		t.Fatalf("GetMostPlayedGames returned error: %v", err)
	}
	if mostPlayedGames.Response.RollupDate != 1778198400 || mostPlayedGames.Response.Ranks[1].LastWeekRank != 31 {
		t.Fatalf("unexpected most played games: %#v", mostPlayedGames.Response)
	}

	topReleasesPages, err := client.API.SteamChartsService.GetTopReleasesPages(context.Background())
	if err != nil {
		t.Fatalf("GetTopReleasesPages returned error: %v", err)
	}
	if len(topReleasesPages.Response.Pages) != 2 || topReleasesPages.Response.Pages[0].ItemIDs[0].AppID != 2246340 {
		t.Fatalf("unexpected top releases pages: %#v", topReleasesPages.Response.Pages)
	}

	rtimeYear := uint32(1746769043)
	yearLimit := uint32(20)
	yearTopReleases, err := client.API.SteamChartsService.GetYearTopAppReleases(context.Background(), &steamchartsservice.GetYearTopAppReleasesOptions{
		RTimeYear:       &rtimeYear,
		IncludeDLC:      &includeDLC,
		TopResultsLimit: &yearLimit,
	})
	if err != nil {
		t.Fatalf("GetYearTopAppReleases returned error: %v", err)
	}
	if len(yearTopReleases.Response.TopDLCReleases) != 2 || len(yearTopReleases.Response.TopAppList) != 2 || yearTopReleases.Response.TopAppList[1].AppID != 550 {
		t.Fatalf("unexpected year top releases: %#v", yearTopReleases.Response)
	}
}

func TestSteamDirectory(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ISteamDirectory/GetCMListForConnect/v1/":
			query := r.URL.Query()
			if got := query.Get("cellid"); got != "123" {
				t.Fatalf("unexpected cellid: %s", got)
			}
			if got := query.Get("cmtype"); got != "websockets" {
				t.Fatalf("unexpected cmtype: %s", got)
			}
			if got := query.Get("realm"); got != "steamglobal" {
				t.Fatalf("unexpected realm: %s", got)
			}
			if got := query.Get("maxcount"); got != "5" {
				t.Fatalf("unexpected maxcount: %s", got)
			}
			if got := query.Get("qoslevel"); got != "2" {
				t.Fatalf("unexpected qoslevel: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"serverlist":[{"endpoint":"cmp1-lax1.steamserver.net:443","legacy_endpoint":"cmp2-lax1.steamserver.net:443","type":"websockets","dc":"lax1","realm":"steamglobal","load":18,"wtd_load":12.5522112846374512},{"endpoint":"205.196.6.148:27017","legacy_endpoint":"205.196.6.148:27017","type":"netfilter","dc":"sea1","realm":"steamglobal","load":15,"wtd_load":52.0302999019622803}],"success":true,"message":""}}`))
		case "/ISteamDirectory/GetSteamPipeDomains/v1/":
			query := r.URL.Query()
			if got := query.Get("key"); got != "test-key" {
				t.Fatalf("unexpected api key: %s", got)
			}
			if len(query) != 1 {
				t.Fatalf("unexpected query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"response":{"domainlist":["*.steamcontent.com","cs.steampowered.com","fastly.cdn.steampipe.steamcontent.com"],"result":1,"message":""}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	cellID := uint32(123)
	maxCount := uint32(5)
	qosLevel := uint32(2)
	cmList, err := client.API.SteamDirectory.GetCMListForConnect(context.Background(), &steamdirectory.GetCMListForConnectOptions{
		CellID:   &cellID,
		CMType:   "websockets",
		Realm:    "steamglobal",
		MaxCount: &maxCount,
		QOSLevel: &qosLevel,
	})
	if err != nil {
		t.Fatalf("GetCMListForConnect returned error: %v", err)
	}
	if !cmList.Response.Success || len(cmList.Response.ServerList) != 2 {
		t.Fatalf("unexpected cm list: %#v", cmList.Response)
	}
	if cmList.Response.ServerList[0].DC != "lax1" || cmList.Response.ServerList[1].WTDLoad <= 0 {
		t.Fatalf("unexpected cm server list: %#v", cmList.Response.ServerList)
	}

	steamPipeDomains, err := client.API.SteamDirectory.GetSteamPipeDomains(context.Background())
	if err != nil {
		t.Fatalf("GetSteamPipeDomains returned error: %v", err)
	}
	if steamPipeDomains.Response.Result != 1 || len(steamPipeDomains.Response.DomainList) != 3 {
		t.Fatalf("unexpected steam pipe domains: %#v", steamPipeDomains.Response)
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

func TestSteamUserAdditionalEndpoints(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if got := query.Get("key"); got != "test-key" {
			t.Fatalf("unexpected api key: %s", got)
		}
		switch r.URL.Path {
		case "/ISteamUser/GetFriendList/v1/":
			if got := query.Get("steamid"); got != "76561198370695025" {
				t.Fatalf("unexpected steamid: %s", got)
			}
			if got := query.Get("relationship"); got != "friend" {
				t.Fatalf("unexpected relationship: %s", got)
			}
			_, _ = w.Write([]byte(`{"friendslist":{"friends":[{"steamid":"76561198291978477","relationship":"friend","friend_since":1712215394},{"steamid":"76561198370695025","relationship":"friend","friend_since":1664934882}]}}`))
		case "/ISteamUser/GetPlayerBans/v1/":
			if got := query.Get("steamids"); got != "76561198856448829,76561198370695025" {
				t.Fatalf("unexpected steamids: %s", got)
			}
			_, _ = w.Write([]byte(`{"players":[{"SteamId":"76561198370695025","CommunityBanned":false,"VACBanned":false,"NumberOfVACBans":0,"DaysSinceLastBan":766,"NumberOfGameBans":1,"EconomyBan":"none"},{"SteamId":"76561198856448829","CommunityBanned":false,"VACBanned":true,"NumberOfVACBans":1,"DaysSinceLastBan":1049,"NumberOfGameBans":0,"EconomyBan":"none"}]}`))
		case "/ISteamUser/GetUserGroupList/v1/":
			if got := query.Get("steamid"); got != "76561198856448829" {
				t.Fatalf("unexpected steamid: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"success":true,"groups":[{"gid":"32222334"},{"gid":"35210443"},{"gid":"38546535"}]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	friends, err := client.API.SteamUser.GetFriendList(
		context.Background(),
		"76561198370695025",
		&steamuser.GetFriendListOptions{Relationship: "friend"},
	)
	if err != nil {
		t.Fatalf("GetFriendList returned error: %v", err)
	}
	if len(friends.FriendsList.Friends) != 2 || friends.FriendsList.Friends[0].FriendSince != 1712215394 {
		t.Fatalf("unexpected friend list: %#v", friends.FriendsList.Friends)
	}

	bans, err := client.API.SteamUser.GetPlayerBans(context.Background(), []string{"76561198856448829", "76561198370695025"})
	if err != nil {
		t.Fatalf("GetPlayerBans returned error: %v", err)
	}
	if len(bans.Players) != 2 || !bans.Players[1].VACBanned {
		t.Fatalf("unexpected player bans: %#v", bans.Players)
	}

	groupList, err := client.API.SteamUser.GetUserGroupList(context.Background(), "76561198856448829")
	if err != nil {
		t.Fatalf("GetUserGroupList returned error: %v", err)
	}
	if !groupList.Response.Success || len(groupList.Response.Groups) != 3 || groupList.Response.Groups[0].GID != "32222334" {
		t.Fatalf("unexpected group list: %#v", groupList.Response)
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

func TestSteamUserOAuthUsesExplicitAccessToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		switch r.URL.Path {
		case "/ISteamUserOAuth/GetUserSummaries/v1/":
			if got := query.Get("steamids"); got != "76561198856448829,76561198370695025" {
				t.Fatalf("unexpected steamids: %s", got)
			}
			_, _ = w.Write([]byte(`{"players":[{"steamid":"76561198370695025","personaname":"百兽发布"},{"steamid":"76561198856448829","personaname":"-"}]}`))
		case "/ISteamUserOAuth/GetFriendList/v1/":
			if got := query.Get("access_token"); got != "user-token" {
				t.Fatalf("unexpected access token: %s", got)
			}
			_, _ = w.Write([]byte(`{"friends":[{"steamid":"76561198095859886","relationship":"friend","friend_since":1615463685},{"steamid":"76561198856448829","relationship":"friend","friend_since":1664934883}]}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithAPIKey("test-key"),
		steam.WithAccessToken("global-token"),
		steam.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	summaries, err := client.API.SteamUserOAuth.GetUserSummaries(context.Background(), []string{"76561198856448829", "76561198370695025"})
	if err != nil {
		t.Fatalf("GetUserSummaries returned error: %v", err)
	}
	if len(summaries.Players) != 2 || summaries.Players[0].PersonaName != "百兽发布" {
		t.Fatalf("unexpected summaries: %#v", summaries.Players)
	}

	friends, err := client.API.SteamUserOAuth.GetFriendList(context.Background(), "user-token")
	if err != nil {
		t.Fatalf("GetFriendList returned error: %v", err)
	}
	if len(friends.Friends) != 2 || friends.Friends[0].FriendSince != 1615463685 {
		t.Fatalf("unexpected friend list: %#v", friends.Friends)
	}
}

func TestSteamUserOAuthValidation(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient(steam.WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.API.SteamUserOAuth.GetUserSummaries(context.Background(), nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	tooMany := make([]string, 101)
	for i := range tooMany {
		tooMany[i] = "1"
	}
	_, err = client.API.SteamUserOAuth.GetUserSummaries(context.Background(), tooMany)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SteamUserOAuth.GetFriendList(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)
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

	_, err = client.API.SteamUser.GetFriendList(context.Background(), "", nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SteamUser.GetPlayerBans(context.Background(), nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SteamUser.GetUserGroupList(context.Background(), "")
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

func TestSteamNotificationServiceUsesExplicitAccessToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ISteamNotificationService/GetPreferences/v1/":
			query := r.URL.Query()
			if got := query.Get("access_token"); got != "user-token" {
				t.Fatalf("unexpected access token: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"preferences":[{"notification_type":2,"notification_targets":11},{"notification_type":3,"notification_targets":1}]}}`))
		case "/ISteamNotificationService/GetSteamNotifications/v1/":
			query := r.URL.Query()
			if got := query.Get("access_token"); got != "user-token" {
				t.Fatalf("unexpected access token: %s", got)
			}
			if got := query.Get("include_hidden"); got != "true" {
				t.Fatalf("unexpected include_hidden: %s", got)
			}
			if got := query.Get("language"); got != "6" {
				t.Fatalf("unexpected language: %s", got)
			}
			if got := query.Get("include_confirmation_count"); got != "true" {
				t.Fatalf("unexpected include_confirmation_count: %s", got)
			}
			if got := query.Get("include_pinned_counts"); got != "true" {
				t.Fatalf("unexpected include_pinned_counts: %s", got)
			}
			if got := query.Get("include_read"); got != "true" {
				t.Fatalf("unexpected include_read: %s", got)
			}
			if got := query.Get("count_only"); got != "false" {
				t.Fatalf("unexpected count_only: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"notifications":[{"notification_id":"147813875091","notification_targets":3,"notification_type":8,"body_data":"{\"appid\":1139900,\"count\":1}","read":false,"timestamp":1777309268,"hidden":false,"expiry":1778515200,"viewed":1778159038}],"confirmation_count":0,"pending_gift_count":0,"pending_friend_count":0,"unread_count":19,"pending_family_invite_count":0}}`))
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

	preferences, err := client.API.SteamNotificationService.GetPreferences(context.Background(), "user-token")
	if err != nil {
		t.Fatalf("GetPreferences returned error: %v", err)
	}
	if len(preferences.Response.Preferences) != 2 || preferences.Response.Preferences[0].NotificationTargets != 11 {
		t.Fatalf("unexpected preferences: %#v", preferences.Response.Preferences)
	}

	trueValue := true
	falseValue := false
	language := int32(6)
	notifications, err := client.API.SteamNotificationService.GetSteamNotifications(
		context.Background(),
		"user-token",
		&steamnotificationservice.GetSteamNotificationsOptions{
			IncludeHidden:            &trueValue,
			Language:                 &language,
			IncludeConfirmationCount: &trueValue,
			IncludePinnedCounts:      &trueValue,
			IncludeRead:              &trueValue,
			CountOnly:                &falseValue,
		},
	)
	if err != nil {
		t.Fatalf("GetSteamNotifications returned error: %v", err)
	}
	if notifications.Response.UnreadCount != 19 || len(notifications.Response.Notifications) != 1 {
		t.Fatalf("unexpected notifications payload: %#v", notifications.Response)
	}
	if notifications.Response.Notifications[0].NotificationID != "147813875091" {
		t.Fatalf("unexpected notification id: %#v", notifications.Response.Notifications[0])
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

func TestPlayerServicePurchasedAndRecentlyPlayedEndpoints(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/IPlayerService/GetPurchasedAndUpgradedProfileCustomizations/v1/":
			query := r.URL.Query()
			if got := query.Get("access_token"); got != "user-token" {
				t.Fatalf("unexpected access token: %s", got)
			}
			if got := query.Get("steamid"); got != "76561198370695025" {
				t.Fatalf("unexpected steamid: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"purchased_customizations":[{"customization_type":6,"count":1},{"customization_type":24,"count":1}],"upgraded_customizations":[{"customization_type":1,"level":3},{"customization_type":17,"level":2}]}}`))
		case "/IPlayerService/GetPurchasedProfileCustomizations/v1/":
			if got := r.URL.Query().Get("steamid"); got != "76561198370695025" {
				t.Fatalf("unexpected steamid: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"purchased_customizations":[{"purchaseid":"916672","customization_type":6},{"purchaseid":"7442062","customization_type":24}]}}`))
		case "/IPlayerService/GetRecentlyPlayedGames/v1/":
			query := r.URL.Query()
			if got := query.Get("access_token"); got != "user-token" {
				t.Fatalf("unexpected access token: %s", got)
			}
			if got := query.Get("steamid"); got != "76561198370695025" {
				t.Fatalf("unexpected steamid: %s", got)
			}
			if got := query.Get("count"); got != "5" {
				t.Fatalf("unexpected count: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"total_count":1,"games":[{"appid":550,"name":"Left 4 Dead 2","playtime_2weeks":177,"playtime_forever":51162,"img_icon_url":"7d5a243f9500d2f8467312822f8af2a2928777ed","playtime_windows_forever":16852,"playtime_mac_forever":0,"playtime_linux_forever":0,"playtime_deck_forever":0}]}}`))
		case "/IPlayerService/GetSteamLevel/v1/":
			if got := r.URL.Query().Get("steamid"); got != "76561198370695025" {
				t.Fatalf("unexpected steamid: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"player_level":67}}`))
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

	purchasedAndUpgraded, err := client.API.PlayerService.GetPurchasedAndUpgradedProfileCustomizations(
		context.Background(),
		"user-token",
		"76561198370695025",
	)
	if err != nil {
		t.Fatalf("GetPurchasedAndUpgradedProfileCustomizations returned error: %v", err)
	}
	if len(purchasedAndUpgraded.Response.UpgradedCustomizations) != 2 || purchasedAndUpgraded.Response.UpgradedCustomizations[0].Level != 3 {
		t.Fatalf("unexpected upgraded customizations: %#v", purchasedAndUpgraded.Response.UpgradedCustomizations)
	}

	purchased, err := client.API.PlayerService.GetPurchasedProfileCustomizations(context.Background(), "76561198370695025")
	if err != nil {
		t.Fatalf("GetPurchasedProfileCustomizations returned error: %v", err)
	}
	if len(purchased.Response.PurchasedCustomizations) != 2 || purchased.Response.PurchasedCustomizations[1].PurchaseID != "7442062" {
		t.Fatalf("unexpected purchased customizations: %#v", purchased.Response.PurchasedCustomizations)
	}

	recentlyPlayed, err := client.API.PlayerService.GetRecentlyPlayedGames(
		context.Background(),
		"user-token",
		"76561198370695025",
		&playerservice.GetRecentlyPlayedGamesOptions{Count: 5},
	)
	if err != nil {
		t.Fatalf("GetRecentlyPlayedGames returned error: %v", err)
	}
	if recentlyPlayed.Response.TotalCount != 1 || len(recentlyPlayed.Response.Games) != 1 || recentlyPlayed.Response.Games[0].AppID != 550 {
		t.Fatalf("unexpected recently played response: %#v", recentlyPlayed.Response)
	}

	level, err := client.API.PlayerService.GetSteamLevel(context.Background(), "76561198370695025")
	if err != nil {
		t.Fatalf("GetSteamLevel returned error: %v", err)
	}
	if level.Response.PlayerLevel != 67 {
		t.Fatalf("unexpected steam level: %#v", level.Response)
	}
}

func TestPlayerServiceSteamLevelDistributionAndTopAchievementsEndpoints(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/IPlayerService/GetSteamLevelDistribution/v1/":
			if got := r.URL.Query().Get("player_level"); got != "10" {
				t.Fatalf("unexpected player_level: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"player_level_percentile":89.1774597167969}}`))
		case "/IPlayerService/GetTopAchievementsForGames/v1/":
			query := r.URL.Query()
			if got := query.Get("steamid"); got != "76561198370695025" {
				t.Fatalf("unexpected steamid: %s", got)
			}
			if got := query.Get("language"); got != "en" {
				t.Fatalf("unexpected language: %s", got)
			}
			if got := query.Get("max_achievements"); got != "3" {
				t.Fatalf("unexpected max_achievements: %s", got)
			}
			if got := query.Get("appids[0]"); got != "550" {
				t.Fatalf("unexpected appids[0]: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"games":[{"appid":550,"total_achievements":101,"achievements":[{"statid":517,"bit":6,"name":"Valve Gift Grab 2011 - L4D2","desc":"Collect three gifts dropped by Special Infected in Versus Mode.","icon":"6378d5648f017f2c1039b927f7e4995fd4cc87ab.jpg","icon_gray":"bbdbc3e9dde37cf3e5dacc5b38e6c30f9e7410dd.jpg","hidden":false,"player_percent_unlocked":"7.1"}]}]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := steam.NewClient(steam.WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	distribution, err := client.API.PlayerService.GetSteamLevelDistribution(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetSteamLevelDistribution returned error: %v", err)
	}
	if distribution.Response.PlayerLevelPercentile != 89.1774597167969 {
		t.Fatalf("unexpected distribution: %#v", distribution.Response)
	}

	topAchievements, err := client.API.PlayerService.GetTopAchievementsForGames(
		context.Background(),
		"76561198370695025",
		&playerservice.GetTopAchievementsForGamesOptions{
			Language:        "en",
			MaxAchievements: 3,
			AppIDs:          []uint32{550},
		},
	)
	if err != nil {
		t.Fatalf("GetTopAchievementsForGames returned error: %v", err)
	}
	if len(topAchievements.Response.Games) != 1 || len(topAchievements.Response.Games[0].Achievements) != 1 || topAchievements.Response.Games[0].Achievements[0].PlayerPercentUnlocked != "7.1" {
		t.Fatalf("unexpected top achievements response: %#v", topAchievements.Response)
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

	_, err = client.API.SteamNotificationService.GetPreferences(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SteamNotificationService.GetSteamNotifications(context.Background(), "", nil)
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

	_, err = client.API.PlayerService.GetPurchasedAndUpgradedProfileCustomizations(context.Background(), "", "76561198370695025")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetPurchasedAndUpgradedProfileCustomizations(context.Background(), "user-token", "")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetPurchasedProfileCustomizations(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetRecentlyPlayedGames(context.Background(), "", "76561198370695025", nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetRecentlyPlayedGames(context.Background(), "user-token", "", nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetSteamLevel(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetTopAchievementsForGames(context.Background(), "", nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetTopAchievementsForGames(context.Background(), "76561198370695025", nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetTopAchievementsForGames(context.Background(), "76561198370695025", &playerservice.GetTopAchievementsForGamesOptions{})
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetTopAchievementsForGames(context.Background(), "76561198370695025", &playerservice.GetTopAchievementsForGamesOptions{AppIDs: []uint32{0}})
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.PlayerService.GetTopAchievementsForGames(context.Background(), "76561198370695025", &playerservice.GetTopAchievementsForGamesOptions{MaxAchievements: 9, AppIDs: []uint32{550}})
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
		if got := query.Get("tags"); got != "patchnotes,events" {
			t.Fatalf("unexpected tags query: %s", got)
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
			Tags:      []string{"patchnotes", "events"},
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

func TestSteamUserStatsAdditionalEndpoints(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		switch r.URL.Path {
		case "/ISteamUserStats/GetGlobalAchievementPercentagesForApp/v2/":
			if got := query.Get("gameid"); got != "550" {
				t.Fatalf("unexpected gameid: %s", got)
			}
			_, _ = w.Write([]byte(`{"achievementpercentages":{"achievements":[{"name":"GLOBAL_GNOME_ALONE","percent":"68.6"}]}}`))
		case "/ISteamUserStats/GetSchemaForGame/v2/":
			if got := query.Get("appid"); got != "550" {
				t.Fatalf("unexpected appid: %s", got)
			}
			_, _ = w.Write([]byte(`{"game":{"gameName":"Left 4 Dead 2","gameVersion":"143","availableGameStats":{"achievements":[{"name":"ACH_HONK_A_CLOWNS_NOSE","defaultvalue":0,"displayName":"CL0WND","hidden":0,"description":"Honk the noses of 10 Clowns.","icon":"icon","icongray":"gray"}],"stats":[{"name":"Stat.GamesPlayed.Total","defaultvalue":0,"displayName":"txt.Stat.GamesPlayed.Total"}]}}}`))
		case "/ISteamUserStats/GetNumberOfCurrentPlayers/v1/":
			if got := query.Get("appid"); got != "550" {
				t.Fatalf("unexpected appid: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"player_count":21495,"result":1}}`))
		case "/ISteamUserStats/GetUserStatsForGame/v2/":
			if got := query.Get("steamid"); got != "76561198370695025" {
				t.Fatalf("unexpected steamid: %s", got)
			}
			if got := query.Get("appid"); got != "550" {
				t.Fatalf("unexpected appid: %s", got)
			}
			_, _ = w.Write([]byte(`{"playerstats":{"steamID":"76561198370695025","gameName":"Left 4 Dead 2","achievements":[{"name":"ACH_HONK_A_CLOWNS_NOSE","achieved":1}],"stats":[{"name":"Stat.GamesPlayed.Total","value":2552},{"name":"Stat.KitsUsed.Avg","value":0.8538401126861572}]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	global, err := client.API.SteamUserStats.GetGlobalAchievementPercentagesForApp(context.Background(), 550)
	if err != nil {
		t.Fatalf("GetGlobalAchievementPercentagesForApp returned error: %v", err)
	}
	if len(global.AchievementPercentages.Achievements) != 1 || global.AchievementPercentages.Achievements[0].Percent != "68.6" {
		t.Fatalf("unexpected global achievements: %#v", global.AchievementPercentages.Achievements)
	}

	schema, err := client.API.SteamUserStats.GetSchemaForGame(context.Background(), 550)
	if err != nil {
		t.Fatalf("GetSchemaForGame returned error: %v", err)
	}
	if schema.Game.GameVersion != "143" || len(schema.Game.AvailableGameStats.Achievements) != 1 {
		t.Fatalf("unexpected schema: %#v", schema.Game)
	}

	currentPlayers, err := client.API.SteamUserStats.GetNumberOfCurrentPlayers(context.Background(), 550)
	if err != nil {
		t.Fatalf("GetNumberOfCurrentPlayers returned error: %v", err)
	}
	if currentPlayers.Response.PlayerCount != 21495 || currentPlayers.Response.Result != 1 {
		t.Fatalf("unexpected current players response: %#v", currentPlayers.Response)
	}

	userStats, err := client.API.SteamUserStats.GetUserStatsForGame(context.Background(), "76561198370695025", 550)
	if err != nil {
		t.Fatalf("GetUserStatsForGame returned error: %v", err)
	}
	if len(userStats.PlayerStats.Achievements) != 1 || len(userStats.PlayerStats.Stats) != 2 || userStats.PlayerStats.Stats[0].Value != 2552 {
		t.Fatalf("unexpected user stats: %#v", userStats.PlayerStats)
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

	_, err = client.API.SteamUserStats.GetGlobalAchievementPercentagesForApp(context.Background(), 0)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SteamUserStats.GetSchemaForGame(context.Background(), 0)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SteamUserStats.GetNumberOfCurrentPlayers(context.Background(), 0)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SteamUserStats.GetUserStatsForGame(context.Background(), "", 550)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.SteamUserStats.GetUserStatsForGame(context.Background(), "1", 0)
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

func TestSteamWebAPIUtilEndpoints(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ISteamWebAPIUtil/GetServerInfo/v1/":
			_, _ = w.Write([]byte(`{"servertime":1778308821,"servertimestring":"Fri May  8 23:40:21 2026"}`))
		case "/ISteamWebAPIUtil/GetSupportedAPIList/v1/":
			if got := r.URL.Query().Get("key"); got != "test-key" {
				t.Fatalf("unexpected api key: %s", got)
			}
			_, _ = w.Write([]byte(`{"apilist":{"interfaces":[{"name":"ISteamWebAPIUtil","methods":[{"name":"GetServerInfo","version":1,"httpmethod":"GET","parameters":[]},{"name":"GetSupportedAPIList","version":1,"httpmethod":"GET","parameters":[{"name":"key","type":"string","optional":true,"description":"access key"}]}]}]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	info, err := client.API.SteamWebAPIUtil.GetServerInfo(context.Background())
	if err != nil {
		t.Fatalf("GetServerInfo returned error: %v", err)
	}
	if info.ServerTime != 1778308821 || info.ServerTimeString != "Fri May  8 23:40:21 2026" {
		t.Fatalf("unexpected server info: %#v", info)
	}

	apiList, err := client.API.SteamWebAPIUtil.GetSupportedAPIList(context.Background())
	if err != nil {
		t.Fatalf("GetSupportedAPIList returned error: %v", err)
	}
	if len(apiList.APIList.Interfaces) != 1 || len(apiList.APIList.Interfaces[0].Methods) != 2 {
		t.Fatalf("unexpected api list: %#v", apiList.APIList)
	}
}

func TestStoreBrowseServiceGetContentHubConfig(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/IStoreBrowseService/GetContentHubConfig/v1/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"response":{"hubconfigs":[{"hubcategoryid":1,"type":"tagids","handle":"action","display_name":"Action","url_path":"category/action","replaces_tags":[19],"must_have_tags":[19]},{"hubcategoryid":86,"type":"contenthub","handle":"adultonly","display_name":"Adult Only","url_path":"adultonly"}]}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	resp, err := client.API.StoreBrowseService.GetContentHubConfig(context.Background())
	if err != nil {
		t.Fatalf("GetContentHubConfig returned error: %v", err)
	}
	if len(resp.Response.HubConfigs) != 2 || resp.Response.HubConfigs[0].MustHaveTags[0] != 19 {
		t.Fatalf("unexpected hub configs: %#v", resp.Response.HubConfigs)
	}
}

func TestStoreCatalogServiceGetDevPageLinks(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/IStoreCatalogService/GetDevPageLinks/v1/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("appid"); got != "550" {
			t.Fatalf("unexpected appid: %s", got)
		}
		_, _ = w.Write([]byte(`{"response":{"links":[{"appid":550,"clan_steamid":"103582791429521412","relation":0,"linkname":"Valve","json":"{\"link_url\":\"https://www.youtube.com/watch?v=Jz6FCFoL3k4&t=4s\"}"}]}}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	resp, err := client.API.StoreCatalogService.GetDevPageLinks(context.Background(), 550)
	if err != nil {
		t.Fatalf("GetDevPageLinks returned error: %v", err)
	}
	if len(resp.Response.Links) != 1 || resp.Response.Links[0].LinkName != "Valve" {
		t.Fatalf("unexpected links: %#v", resp.Response.Links)
	}
}

func TestStorePreferencesServiceUsesExplicitAccessToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/IStorePreferencesService/GetIgnoreList/v1/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		query := r.URL.Query()
		if got := query.Get("access_token"); got != "user-token" {
			t.Fatalf("unexpected access token: %s", got)
		}
		if got := query.Get("key"); got != "test-key" {
			t.Fatalf("unexpected api key: %s", got)
		}
		_, _ = w.Write([]byte(`{"response":{"ignore_list":[{"appid":1005490,"reason":0},{"appid":1128960,"reason":0}]}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithAPIKey("test-key"),
		steam.WithAccessToken("global-token"),
		steam.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	resp, err := client.API.StorePreferencesService.GetIgnoreList(context.Background(), "user-token")
	if err != nil {
		t.Fatalf("GetIgnoreList returned error: %v", err)
	}
	if len(resp.Response.IgnoreList) != 2 || resp.Response.IgnoreList[0].AppID != 1005490 {
		t.Fatalf("unexpected ignore list: %#v", resp.Response.IgnoreList)
	}
}

func TestStoreTopSellersService(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/IStoreTopSellersService/GetCountryList/v1/":
			if got := r.URL.Query().Get("key"); got != "test-key" {
				t.Fatalf("unexpected api key: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"countries":[{"country_code":"CN","name":"China"},{"country_code":"US","name":"United States"}]}}`))
		case "/IStoreTopSellersService/GetWeeklyTopSellers/v1/":
			if got := r.URL.Query().Get("key"); got != "test-key" {
				t.Fatalf("unexpected api key: %s", got)
			}
			var payload struct {
				Context struct {
					CountryCode string `json:"country_code"`
				} `json:"context"`
			}
			if err := json.Unmarshal([]byte(r.URL.Query().Get("input_json")), &payload); err != nil {
				t.Fatalf("unmarshal input_json failed: %v", err)
			}
			if payload.Context.CountryCode != "CN" {
				t.Fatalf("unexpected country code: %s", payload.Context.CountryCode)
			}
			_, _ = w.Write([]byte(`{"response":{"country_code":"CN","top_sellers":[{"appid":550}]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	countryList, err := client.API.StoreTopSellersService.GetCountryList(context.Background())
	if err != nil {
		t.Fatalf("GetCountryList returned error: %v", err)
	}
	if len(countryList.Response.Countries) != 2 || countryList.Response.Countries[0].CountryCode != "CN" {
		t.Fatalf("unexpected countries: %#v", countryList.Response.Countries)
	}

	weekly, err := client.API.StoreTopSellersService.GetWeeklyTopSellers(context.Background(), "CN")
	if err != nil {
		t.Fatalf("GetWeeklyTopSellers returned error: %v", err)
	}
	if !strings.Contains(string(weekly.Response), `"top_sellers"`) {
		t.Fatalf("unexpected weekly response: %s", string(weekly.Response))
	}
}

func TestStoreTopSellersServiceValidation(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient(steam.WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.API.StoreTopSellersService.GetWeeklyTopSellers(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)
}

func TestUserAccountServiceUsesExplicitAccessToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/IUserAccountService/GetUserCountry/v1/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		query := r.URL.Query()
		if got := query.Get("access_token"); got != "user-token" {
			t.Fatalf("unexpected access token: %s", got)
		}
		if got := query.Get("steamid"); got != "76561198370695025" {
			t.Fatalf("unexpected steamid: %s", got)
		}
		if got := query.Get("key"); got != "test-key" {
			t.Fatalf("unexpected api key: %s", got)
		}
		_, _ = w.Write([]byte(`{"response":{"country":"HK"}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithAPIKey("test-key"),
		steam.WithAccessToken("global-token"),
		steam.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	resp, err := client.API.UserAccountService.GetUserCountry(context.Background(), "user-token", "76561198370695025")
	if err != nil {
		t.Fatalf("GetUserCountry returned error: %v", err)
	}
	if resp.Response.Country != "HK" {
		t.Fatalf("unexpected country: %#v", resp.Response)
	}
}

func TestUserReviewsServiceUsesExplicitAccessToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/IUserReviewsService/GetFriendsRecommendedApp/v1/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		query := r.URL.Query()
		if got := query.Get("access_token"); got != "user-token" {
			t.Fatalf("unexpected access token: %s", got)
		}
		if got := query.Get("appid"); got != "550" {
			t.Fatalf("unexpected appid: %s", got)
		}
		if got := query.Get("key"); got != "test-key" {
			t.Fatalf("unexpected api key: %s", got)
		}
		_, _ = w.Write([]byte(`{"response":{"accountids_recommended":[907748460,1728978895]}}`))
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithAPIKey("test-key"),
		steam.WithAccessToken("global-token"),
		steam.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	resp, err := client.API.UserReviewsService.GetFriendsRecommendedApp(context.Background(), "user-token", 550)
	if err != nil {
		t.Fatalf("GetFriendsRecommendedApp returned error: %v", err)
	}
	if len(resp.Response.AccountIDsRecommended) != 2 || resp.Response.AccountIDsRecommended[0] != 907748460 {
		t.Fatalf("unexpected recommended accounts: %#v", resp.Response.AccountIDsRecommended)
	}
}

func TestUserStoreVisitService(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/IUserStoreVisitService/GetFrequentlyVisitedPages/v1/":
			query := r.URL.Query()
			if got := query.Get("access_token"); got != "user-token" {
				t.Fatalf("unexpected access token: %s", got)
			}
			if got := query.Get("key"); got != "test-key" {
				t.Fatalf("unexpected api key: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"visit_data":{"recent_apps":[{"item_id":{"appid":2483190},"time_visit":1778310751}]},"frequent_hubs":[{"item_id":{"tagid":11014},"time_visit":1777123729,"visit_count":1},{"item_id":{"hubcategoryid":151},"time_visit":1777123302,"visit_count":1}]}}`))
		case "/IUserStoreVisitService/GetMostVisitedItemsOnStore/v1/":
			if got := r.URL.Query().Get("key"); got != "test-key" {
				t.Fatalf("unexpected api key: %s", got)
			}
			var payload struct {
				Context struct {
					CountryCode string `json:"country_code"`
				} `json:"context"`
				DataRequest map[string]any `json:"data_request"`
			}
			if err := json.Unmarshal([]byte(r.URL.Query().Get("input_json")), &payload); err != nil {
				t.Fatalf("unmarshal input_json failed: %v", err)
			}
			if payload.Context.CountryCode != "CN" {
				t.Fatalf("unexpected country code: %s", payload.Context.CountryCode)
			}
			if got, ok := payload.DataRequest["include_assets"].(bool); !ok || !got {
				t.Fatalf("unexpected include_assets: %#v", payload.DataRequest["include_assets"])
			}
			if got, ok := payload.DataRequest["include_tag_count"].(string); !ok || got != "5" {
				t.Fatalf("unexpected include_tag_count: %#v", payload.DataRequest["include_tag_count"])
			}
			_, _ = w.Write([]byte(`{"response":{"item_ids":[{"appid":3526710}],"items":[{"appid":3526710,"name":"Everything is Crab"}]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithAPIKey("test-key"),
		steam.WithAccessToken("global-token"),
		steam.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	visits, err := client.API.UserStoreVisitService.GetFrequentlyVisitedPages(context.Background(), "user-token")
	if err != nil {
		t.Fatalf("GetFrequentlyVisitedPages returned error: %v", err)
	}
	if len(visits.Response.VisitData.RecentApps) != 1 || visits.Response.VisitData.RecentApps[0].ItemID.AppID != 2483190 {
		t.Fatalf("unexpected recent apps: %#v", visits.Response.VisitData.RecentApps)
	}
	if len(visits.Response.FrequentHubs) != 2 || visits.Response.FrequentHubs[1].ItemID.HubCategoryID != 151 {
		t.Fatalf("unexpected frequent hubs: %#v", visits.Response.FrequentHubs)
	}

	trueValue := true
	mostVisited, err := client.API.UserStoreVisitService.GetMostVisitedItemsOnStore(
		context.Background(),
		"CN",
		&userstorevisitservice.GetMostVisitedItemsOnStoreOptions{
			IncludeAssets:                 &trueValue,
			IncludeRelease:                &trueValue,
			IncludePlatforms:              &trueValue,
			IncludeAllPurchaseOptions:     &trueValue,
			IncludeScreenshots:            &trueValue,
			IncludeTrailers:               &trueValue,
			IncludeRatings:                &trueValue,
			IncludeTagCount:               "5",
			IncludeReviews:                &trueValue,
			IncludeBasicInfo:              &trueValue,
			IncludeSupportedLanguages:     &trueValue,
			IncludeFullDescription:        &trueValue,
			IncludeIncludedItems:          &trueValue,
			IncludeAssetsWithoutOverrides: &trueValue,
			ApplyUserFilters:              &trueValue,
			IncludeLinks:                  &trueValue,
		},
	)
	if err != nil {
		t.Fatalf("GetMostVisitedItemsOnStore returned error: %v", err)
	}
	if len(mostVisited.Response.ItemIDs) != 1 || mostVisited.Response.ItemIDs[0].AppID != 3526710 {
		t.Fatalf("unexpected item ids: %#v", mostVisited.Response.ItemIDs)
	}
	if len(mostVisited.Response.Items) != 1 {
		t.Fatalf("unexpected items: %d", len(mostVisited.Response.Items))
	}
}

func TestWishlistService(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/IWishlistService/GetWishlist/v1/":
			if got := r.URL.Query().Get("steamid"); got != "76561198370695025" {
				t.Fatalf("unexpected steamid: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"items":[{"appid":4000,"priority":2,"date_added":1494672274},{"appid":25000,"priority":0,"date_added":1717842914}]}}`))
		case "/IWishlistService/GetWishlistItemCount/v1/":
			if got := r.URL.Query().Get("steamid"); got != "76561198370695025" {
				t.Fatalf("unexpected steamid: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"count":1119}}`))
		case "/IWishlistService/GetWishlistItemsOnSale/v1/":
			query := r.URL.Query()
			if got := query.Get("access_token"); got != "user-token" {
				t.Fatalf("unexpected access token: %s", got)
			}
			if got := query.Get("key"); got != "test-key" {
				t.Fatalf("unexpected api key: %s", got)
			}
			var payload struct {
				Context struct {
					CountryCode string `json:"country_code"`
				} `json:"context"`
				DataRequest map[string]any `json:"data_request"`
			}
			if err := json.Unmarshal([]byte(query.Get("input_json")), &payload); err != nil {
				t.Fatalf("unmarshal input_json failed: %v", err)
			}
			if payload.Context.CountryCode != "CN" {
				t.Fatalf("unexpected country code: %s", payload.Context.CountryCode)
			}
			if got, ok := payload.DataRequest["include_assets"].(bool); !ok || !got {
				t.Fatalf("unexpected include_assets: %#v", payload.DataRequest["include_assets"])
			}
			if got, ok := payload.DataRequest["include_tag_count"].(string); !ok || got != "5" {
				t.Fatalf("unexpected include_tag_count: %#v", payload.DataRequest["include_tag_count"])
			}
			_, _ = w.Write([]byte(`{"response":{"items":[{"appid":1086940,"store_item":{"name":"Baldur's Gate 3"}},{"appid":813230,"store_item":{"name":"ANIMAL WELL"}}],"total_items_on_sale":43}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithAPIKey("test-key"),
		steam.WithAccessToken("global-token"),
		steam.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	wishlist, err := client.API.WishlistService.GetWishlist(context.Background(), "76561198370695025")
	if err != nil {
		t.Fatalf("GetWishlist returned error: %v", err)
	}
	if len(wishlist.Response.Items) != 2 || wishlist.Response.Items[0].AppID != 4000 {
		t.Fatalf("unexpected wishlist items: %#v", wishlist.Response.Items)
	}

	count, err := client.API.WishlistService.GetWishlistItemCount(context.Background(), "76561198370695025")
	if err != nil {
		t.Fatalf("GetWishlistItemCount returned error: %v", err)
	}
	if count.Response.Count != 1119 {
		t.Fatalf("unexpected wishlist count: %#v", count.Response)
	}

	trueValue := true
	onSale, err := client.API.WishlistService.GetWishlistItemsOnSale(
		context.Background(),
		"user-token",
		"CN",
		&wishlistservice.GetWishlistItemsOnSaleOptions{
			IncludeAssets:                 &trueValue,
			IncludeRelease:                &trueValue,
			IncludePlatforms:              &trueValue,
			IncludeAllPurchaseOptions:     &trueValue,
			IncludeScreenshots:            &trueValue,
			IncludeTrailers:               &trueValue,
			IncludeRatings:                &trueValue,
			IncludeTagCount:               "5",
			IncludeReviews:                &trueValue,
			IncludeBasicInfo:              &trueValue,
			IncludeSupportedLanguages:     &trueValue,
			IncludeFullDescription:        &trueValue,
			IncludeIncludedItems:          &trueValue,
			IncludeAssetsWithoutOverrides: &trueValue,
			ApplyUserFilters:              &trueValue,
			IncludeLinks:                  &trueValue,
		},
	)
	if err != nil {
		t.Fatalf("GetWishlistItemsOnSale returned error: %v", err)
	}
	if len(onSale.Response.Items) != 2 || onSale.Response.Items[1].AppID != 813230 {
		t.Fatalf("unexpected on-sale items: %#v", onSale.Response.Items)
	}
	if onSale.Response.TotalItemsOnSale != 43 {
		t.Fatalf("unexpected total_items_on_sale: %d", onSale.Response.TotalItemsOnSale)
	}
}

func TestUserScopedServicesValidation(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient(steam.WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.API.UserAccountService.GetUserCountry(context.Background(), "", "76561198370695025")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.UserAccountService.GetUserCountry(context.Background(), "user-token", "")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.UserReviewsService.GetFriendsRecommendedApp(context.Background(), "", 550)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.UserReviewsService.GetFriendsRecommendedApp(context.Background(), "user-token", 0)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.UserStoreVisitService.GetFrequentlyVisitedPages(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.UserStoreVisitService.GetMostVisitedItemsOnStore(context.Background(), "", nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.WishlistService.GetWishlist(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.WishlistService.GetWishlistItemCount(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.WishlistService.GetWishlistItemsOnSale(context.Background(), "", "CN", nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.WishlistService.GetWishlistItemsOnSale(context.Background(), "user-token", "", nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)
}

func TestStoreServiceEndpoints(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		switch r.URL.Path {
		case "/IStoreService/GetAppList/v1/":
			if got := query.Get("include_games"); got != "true" {
				t.Fatalf("unexpected include_games: %s", got)
			}
			if got := query.Get("include_dlc"); got != "true" {
				t.Fatalf("unexpected include_dlc: %s", got)
			}
			if got := query.Get("include_software"); got != "true" {
				t.Fatalf("unexpected include_software: %s", got)
			}
			if got := query.Get("include_videos"); got != "true" {
				t.Fatalf("unexpected include_videos: %s", got)
			}
			if got := query.Get("include_hardware"); got != "true" {
				t.Fatalf("unexpected include_hardware: %s", got)
			}
			if got := query.Get("last_appid"); got != "550" {
				t.Fatalf("unexpected last_appid: %s", got)
			}
			if got := query.Get("max_results"); got != "10" {
				t.Fatalf("unexpected max_results: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"apps":[{"appid":570,"name":"Dota 2","last_modified":1769535998,"price_change_number":23683736},{"appid":620,"name":"Portal 2","last_modified":1745363004,"price_change_number":34672985}],"have_more_results":true,"last_appid":620}}`))
		case "/IStoreService/GetGamesFollowed/v1/":
			if got := query.Get("steamid"); got != "76561198370695025" {
				t.Fatalf("unexpected steamid: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"appids":[550,219740,431960]}}`))
		case "/IStoreService/GetGamesFollowedCount/v1/":
			if got := query.Get("steamid"); got != "76561198370695025" {
				t.Fatalf("unexpected steamid: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"followed_game_count":33}}`))
		case "/IStoreService/GetMostPopularTags/v1/":
			_, _ = w.Write([]byte(`{"response":{"tags":[{"tagid":492,"name":"Indie"},{"tagid":19,"name":"Action"}]}}`))
		case "/IStoreService/GetUserGameInterestState/v1/":
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			if got := query.Get("access_token"); got != "user-token" {
				t.Fatalf("unexpected access token: %s", got)
			}
			if got := query.Get("appid"); got != "550" {
				t.Fatalf("unexpected appid: %s", got)
			}
			if got := query.Get("store_appid"); got != "551" {
				t.Fatalf("unexpected store_appid: %s", got)
			}
			if got := query.Get("beta_appid"); got != "552" {
				t.Fatalf("unexpected beta_appid: %s", got)
			}
			_, _ = w.Write([]byte(`{"response":{"owned":true,"following":true,"in_queues":[1],"queue_items_remaining":[12],"queue_items_next_appid":[2288340],"queues":[{"type":1,"skipped":false,"items_remaining":12,"next_appid":2288340,"experimental_cohort":4}]}}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := steam.NewClient(
		steam.WithAPIKey("test-key"),
		steam.WithAccessToken("global-token"),
		steam.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	trueValue := true
	lastAppID := uint32(550)
	maxResults := uint32(10)
	appList, err := client.API.StoreService.GetAppList(context.Background(), &storeservice.GetAppListOptions{
		IncludeGames:    &trueValue,
		IncludeDLC:      &trueValue,
		IncludeSoftware: &trueValue,
		IncludeVideos:   &trueValue,
		IncludeHardware: &trueValue,
		LastAppID:       &lastAppID,
		MaxResults:      &maxResults,
	})
	if err != nil {
		t.Fatalf("GetAppList returned error: %v", err)
	}
	if len(appList.Response.Apps) != 2 || !appList.Response.HaveMoreResults || appList.Response.LastAppID != 620 {
		t.Fatalf("unexpected app list: %#v", appList.Response)
	}

	followed, err := client.API.StoreService.GetGamesFollowed(context.Background(), "76561198370695025")
	if err != nil {
		t.Fatalf("GetGamesFollowed returned error: %v", err)
	}
	if len(followed.Response.AppIDs) != 3 || followed.Response.AppIDs[0] != 550 {
		t.Fatalf("unexpected followed games: %#v", followed.Response.AppIDs)
	}

	followedCount, err := client.API.StoreService.GetGamesFollowedCount(context.Background(), "76561198370695025")
	if err != nil {
		t.Fatalf("GetGamesFollowedCount returned error: %v", err)
	}
	if followedCount.Response.FollowedGameCount != 33 {
		t.Fatalf("unexpected followed game count: %#v", followedCount.Response)
	}

	tags, err := client.API.StoreService.GetMostPopularTags(context.Background())
	if err != nil {
		t.Fatalf("GetMostPopularTags returned error: %v", err)
	}
	if len(tags.Response.Tags) != 2 || tags.Response.Tags[1].TagID != 19 {
		t.Fatalf("unexpected tags: %#v", tags.Response.Tags)
	}

	storeAppID := uint32(551)
	betaAppID := uint32(552)
	interestState, err := client.API.StoreService.GetUserGameInterestState(
		context.Background(),
		"user-token",
		550,
		&storeservice.GetUserGameInterestStateOptions{
			StoreAppID: &storeAppID,
			BetaAppID:  &betaAppID,
		},
	)
	if err != nil {
		t.Fatalf("GetUserGameInterestState returned error: %v", err)
	}
	if !interestState.Response.Owned || !interestState.Response.Following || len(interestState.Response.Queues) != 1 {
		t.Fatalf("unexpected interest state: %#v", interestState.Response)
	}
}

func TestStoreServiceValidation(t *testing.T) {
	t.Parallel()

	client, err := steam.NewClient(steam.WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	_, err = client.API.StoreCatalogService.GetDevPageLinks(context.Background(), 0)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.StorePreferencesService.GetIgnoreList(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.StoreService.GetGamesFollowed(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.StoreService.GetGamesFollowedCount(context.Background(), "")
	expectKind(t, err, steam.ErrorKindRequestBuild)

	maxResults := uint32(50001)
	_, err = client.API.StoreService.GetAppList(context.Background(), &storeservice.GetAppListOptions{MaxResults: &maxResults})
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.StoreService.GetUserGameInterestState(context.Background(), "", 550, nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)

	_, err = client.API.StoreService.GetUserGameInterestState(context.Background(), "user-token", 0, nil)
	expectKind(t, err, steam.ErrorKindRequestBuild)
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
