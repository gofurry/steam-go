package community

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestListInventoryPaginatesAndDoesNotMutateOptions(t *testing.T) {
	t.Parallel()

	communityTransport := &recordingTransport{
		statuses: []int{http.StatusOK, http.StatusOK},
		responseBodies: []string{
			`{"success":1,"assets":[{"appid":730,"contextid":"2","assetid":"10","classid":"20","instanceid":"0","amount":"1"}],"descriptions":[],"total_inventory_count":2,"more_items":true,"last_assetid":"11"}`,
			`{"success":1,"assets":[{"appid":730,"contextid":"2","assetid":"11","classid":"21","instanceid":"0","amount":"1"}],"descriptions":[],"total_inventory_count":2,"more_items":false,"last_assetid":"11"}`,
		},
	}
	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, communityTransport, 1024)
	opts := &ListInventoryOptions{
		Query: GetInventoryOptions{
			Count:        1,
			StartAssetID: "10",
		},
	}

	var pages []InventoryPage
	err := service.ListInventory(context.Background(), "76561198370695025", 730, "2", opts, func(page InventoryPage) error {
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		t.Fatalf("ListInventory returned error: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(pages))
	}
	if pages[0].Cursor != "10" || pages[1].Cursor != "11" {
		t.Fatalf("unexpected page cursors: %#v", pages)
	}
	if opts.Query.StartAssetID != "10" {
		t.Fatalf("ListInventory mutated options start asset id: %q", opts.Query.StartAssetID)
	}

	communityTransport.mu.Lock()
	defer communityTransport.mu.Unlock()
	if len(communityTransport.requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(communityTransport.requests))
	}
	assertQuery(t, communityTransport.requests[0].query, "start_assetid", "10")
	assertQuery(t, communityTransport.requests[1].query, "start_assetid", "11")
}

func TestListInventoryStopsOnHandlerError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("stop")
	communityTransport := &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: `{"success":1,"assets":[],"descriptions":[],"total_inventory_count":0,"more_items":false}`,
	}
	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, communityTransport, 1024)

	err := service.ListInventory(context.Background(), "76561198370695025", 730, "2", nil, func(InventoryPage) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected handler error, got %v", err)
	}
}
