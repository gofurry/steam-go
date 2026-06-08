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

func TestListInventoryStopsOnRepeatedLastAssetID(t *testing.T) {
	t.Parallel()

	communityTransport := &recordingTransport{
		statuses: []int{http.StatusOK, http.StatusOK, http.StatusOK},
		responseBodies: []string{
			`{"success":1,"assets":[{"appid":730,"contextid":"2","assetid":"10","classid":"20","instanceid":"0","amount":"1"}],"descriptions":[],"total_inventory_count":3,"more_items":true,"last_assetid":"11"}`,
			`{"success":1,"assets":[{"appid":730,"contextid":"2","assetid":"11","classid":"21","instanceid":"0","amount":"1"}],"descriptions":[],"total_inventory_count":3,"more_items":true,"last_assetid":"11"}`,
			`{"success":1,"assets":[{"appid":730,"contextid":"2","assetid":"12","classid":"22","instanceid":"0","amount":"1"}],"descriptions":[],"total_inventory_count":3,"more_items":false,"last_assetid":"12"}`,
		},
	}
	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, communityTransport, 1024)

	var pages []InventoryPage
	err := service.ListInventory(context.Background(), "76561198370695025", 730, "2", nil, func(page InventoryPage) error {
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		t.Fatalf("ListInventory returned error: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages before repeated last_assetid stop, got %d", len(pages))
	}

	communityTransport.mu.Lock()
	defer communityTransport.mu.Unlock()
	if len(communityTransport.requests) != 2 {
		t.Fatalf("expected repeated last_assetid to stop without third request, got %d requests", len(communityTransport.requests))
	}
}

func TestListInventoryHonorsContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{statuses: []int{http.StatusOK}}, 1024)

	err := service.ListInventory(ctx, "76561198370695025", 730, "2", nil, func(InventoryPage) error {
		t.Fatal("handler should not run after context cancellation")
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestListInventoryRejectsNegativeMaxPages(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{statuses: []int{http.StatusOK}}, 1024)
	err := service.ListInventory(context.Background(), "76561198370695025", 730, "2", &ListInventoryOptions{MaxPages: -1}, func(InventoryPage) error {
		t.Fatal("handler should not run for invalid max pages")
		return nil
	})
	if err == nil {
		t.Fatal("expected max pages validation error")
	}
}

func TestJoinInventoryDescriptionsPreservesOrderAndMatchesDescriptions(t *testing.T) {
	t.Parallel()

	resp := InventoryResponse{
		Assets: []InventoryAsset{
			{AppID: 730, AssetID: "asset-1", ClassID: "class-a", InstanceID: "0"},
			{AppID: 730, AssetID: "asset-2", ClassID: "class-missing", InstanceID: "0"},
			{AppID: 730, AssetID: "asset-3", ClassID: "class-b", InstanceID: "1"},
		},
		Descriptions: []InventoryItem{
			{AppID: 730, ClassID: "class-b", InstanceID: "1", Name: "Second"},
			{AppID: 730, ClassID: "class-a", InstanceID: "0", Name: "First"},
			{AppID: 730, ClassID: "class-a", InstanceID: "0", Name: "Duplicate ignored"},
		},
	}

	joined := JoinInventoryDescriptions(resp)
	if len(joined) != 3 {
		t.Fatalf("expected 3 joined items, got %d", len(joined))
	}
	if joined[0].Asset.AssetID != "asset-1" || joined[0].Description == nil || joined[0].Description.Name != "First" {
		t.Fatalf("unexpected first joined item: %#v", joined[0])
	}
	if joined[1].Asset.AssetID != "asset-2" || joined[1].Description != nil {
		t.Fatalf("expected missing description for second item, got %#v", joined[1])
	}
	if joined[2].Asset.AssetID != "asset-3" || joined[2].Description == nil || joined[2].Description.Name != "Second" {
		t.Fatalf("unexpected third joined item: %#v", joined[2])
	}
}

func TestJoinInventoryDescriptionsReturnsNilForEmptyAssets(t *testing.T) {
	t.Parallel()

	if got := JoinInventoryDescriptions(InventoryResponse{}); got != nil {
		t.Fatalf("expected nil for empty assets, got %#v", got)
	}
}
