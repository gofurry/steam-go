package market

import (
	"context"
	"net/http"
	"testing"
)

func TestGetPriceOverviewBatchPreservesOrderAndPerItemErrors(t *testing.T) {
	t.Parallel()

	marketTransport := &recordingTransport{
		statuses:     []int{http.StatusOK, http.StatusTooManyRequests, http.StatusOK},
		responseBody: `{"success":true,"lowest_price":"$2.30","volume":"1,234","median_price":"$2.40"}`,
	}
	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, marketTransport, 1024)
	items := []PriceOverviewBatchItem{
		{AppID: 440, MarketHashName: "Key"},
		{AppID: 730, MarketHashName: "Case"},
		{AppID: 570, MarketHashName: "Bundle"},
	}

	results, err := service.GetPriceOverviewBatch(context.Background(), items, &GetPriceOverviewBatchOptions{
		Currency:      1,
		MaxConcurrent: 1,
	})
	if err != nil {
		t.Fatalf("GetPriceOverviewBatch returned top-level error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for idx := range items {
		if results[idx].Item != items[idx] {
			t.Fatalf("result %d out of order: %#v", idx, results[idx])
		}
	}
	if results[0].Err != nil || results[2].Err != nil {
		t.Fatalf("expected first and third results to succeed: %#v", results)
	}
	if results[1].Err == nil {
		t.Fatal("expected per-item error for second result")
	}
}

func TestGetPriceOverviewBatchValidatesConcurrency(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{statuses: []int{http.StatusOK}}, 1024)
	_, err := service.GetPriceOverviewBatch(context.Background(), []PriceOverviewBatchItem{{AppID: 440, MarketHashName: "Key"}}, &GetPriceOverviewBatchOptions{MaxConcurrent: -1})
	if err == nil {
		t.Fatal("expected max concurrent validation error")
	}
}
