package storefront

import (
	"context"
	"net/http"
	"testing"
)

func TestGetAdjacentPartnerEventsBuildsRequestAndDecodesResponse(t *testing.T) {
	t.Parallel()

	store := &recordingTransport{
		statuses: []int{http.StatusOK},
		responseBody: `{
			"success": 1,
			"events": [
				{
					"gid": "123",
					"clan_steamid": "103582791432902485",
					"appid": 550,
					"event_name": "Update",
					"event_type": 28,
					"comment_count": 13,
					"forum_topic_id": "topic-123",
					"rtime_created": 1700000001,
					"rtime32_start_time": 1700000002,
					"rtime32_end_time": 1700000003,
					"rtime32_last_modified": 1700000004,
					"published": 1,
					"announcement_body": {
						"gid": "456",
						"event_gid": "123",
						"headline": "Patch Notes",
						"body": "[b]Hello[/b]",
						"posttime": 1700000000,
						"updatetime": 1700000100,
						"url": "https://store.steampowered.com/news/app/550/view/456",
						"commentcount": 12,
						"clanid": "3381077",
						"forum_topic_id": "topic-456",
						"tags": ["patch"],
						"language": 0,
						"voteupcount": 20,
						"votedowncount": 1
					}
				}
			]
		}`,
	}
	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, store, 4096)

	resp, err := service.GetAdjacentPartnerEvents(context.Background(), 550, &GetAdjacentPartnerEventsOptions{
		CountBefore:  1,
		CountAfter:   10,
		LanguageList: "6_0",
	})
	if err != nil {
		t.Fatalf("GetAdjacentPartnerEvents returned error: %v", err)
	}
	if len(resp.Events) != 1 {
		t.Fatalf("unexpected event count: %d", len(resp.Events))
	}
	if resp.Success != 1 {
		t.Fatalf("unexpected success value: %d", resp.Success)
	}
	if got := resp.Events[0].ClanSteamID; got != "103582791432902485" {
		t.Fatalf("unexpected clan steam id: %q", got)
	}
	if got := resp.Events[0].RTimeStart; got != 1700000002 {
		t.Fatalf("unexpected event start time: %d", got)
	}
	if got := resp.Events[0].AnnouncementBody.Headline; got != "Patch Notes" {
		t.Fatalf("unexpected headline: %q", got)
	}
	if got := resp.Events[0].AnnouncementBody.ClanID; got != "3381077" {
		t.Fatalf("unexpected announcement clan id: %q", got)
	}
	if len(resp.Events[0].Raw) == 0 || len(resp.Events[0].AnnouncementBody.Raw) == 0 {
		t.Fatal("expected raw payloads to be preserved")
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	if len(store.requests) != 1 {
		t.Fatalf("expected 1 store request, got %d", len(store.requests))
	}
	req := store.requests[0]
	if req.path != "/events/ajaxgetadjacentpartnerevents" {
		t.Fatalf("unexpected path: %s", req.path)
	}
	assertQuery(t, req.query, "appid", "550")
	assertQuery(t, req.query, "count_before", "1")
	assertQuery(t, req.query, "count_after", "10")
	assertQuery(t, req.query, "lang_list", "6_0")
}

func TestGetAdjacentPartnerEventsValidation(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{statuses: []int{http.StatusOK}}, 1024)

	if _, err := service.GetAdjacentPartnerEvents(context.Background(), 0, nil); err == nil {
		t.Fatal("expected validation error for app id")
	}
	if _, err := service.GetAdjacentPartnerEvents(context.Background(), 550, &GetAdjacentPartnerEventsOptions{CountBefore: -1}); err == nil {
		t.Fatal("expected validation error for count before")
	}
	if _, err := service.GetAdjacentPartnerEvents(context.Background(), 550, &GetAdjacentPartnerEventsOptions{CountAfter: -1}); err == nil {
		t.Fatal("expected validation error for count after")
	}
}
