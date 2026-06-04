package storefront

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestListAppReviewsPaginatesAndDoesNotMutateOptions(t *testing.T) {
	t.Parallel()

	store := &recordingTransport{
		statuses: []int{http.StatusOK, http.StatusOK},
		responseBodies: []string{
			`{"success":1,"query_summary":{"num_reviews":1,"total_reviews":2},"cursor":"next","reviews":[{"recommendationid":"1","author":{"steamid":"2"},"review":"good"}]}`,
			`{"success":1,"query_summary":{"num_reviews":0,"total_reviews":2},"cursor":"next","reviews":[]}`,
		},
	}
	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, store, 1024)
	opts := &ListAppReviewsOptions{
		Query: GetAppReviewsOptions{
			Cursor:     "start",
			NumPerPage: 10,
		},
	}

	var pages []AppReviewsPage
	err := service.ListAppReviews(context.Background(), 550, opts, func(page AppReviewsPage) error {
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		t.Fatalf("ListAppReviews returned error: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(pages))
	}
	if pages[0].Cursor != "start" || pages[1].Cursor != "next" {
		t.Fatalf("unexpected page cursors: %#v", pages)
	}
	if opts.Query.Cursor != "start" {
		t.Fatalf("ListAppReviews mutated options cursor: %q", opts.Query.Cursor)
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	if len(store.requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(store.requests))
	}
	assertQuery(t, store.requests[0].query, "cursor", "start")
	assertQuery(t, store.requests[1].query, "cursor", "next")
}

func TestListAppReviewsStopsOnHandlerError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("stop")
	store := &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: `{"success":1,"query_summary":{"num_reviews":1},"cursor":"next","reviews":[{"recommendationid":"1","author":{"steamid":"2"}}]}`,
	}
	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, store, 1024)

	err := service.ListAppReviews(context.Background(), 550, nil, func(AppReviewsPage) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected handler error, got %v", err)
	}
}

func TestListAppReviewsHonorsContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{statuses: []int{http.StatusOK}}, 1024)

	err := service.ListAppReviews(ctx, 550, nil, func(AppReviewsPage) error {
		t.Fatal("handler should not run after context cancellation")
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestGetAppDetailsBatchPreservesOrderAndPerItemErrors(t *testing.T) {
	t.Parallel()

	store := &recordingTransport{
		statuses:     []int{http.StatusOK, http.StatusOK},
		responseBody: `{"550":{"success":true,"data":{"type":"game","name":"Left 4 Dead 2","steam_appid":550,"required_age":0,"is_free":false,"short_description":"coop","header_image":"header.jpg","developers":[],"publishers":[],"platforms":{"windows":true,"mac":false,"linux":true}}}}`,
	}
	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, store, 1024)

	results, err := service.GetAppDetailsBatch(context.Background(), []uint32{550, 0, 10}, &GetAppDetailsBatchOptions{MaxConcurrent: 1})
	if err != nil {
		t.Fatalf("GetAppDetailsBatch returned top-level error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].AppID != 550 || results[1].AppID != 0 || results[2].AppID != 10 {
		t.Fatalf("results are out of order: %#v", results)
	}
	if results[0].Err != nil || results[2].Err != nil {
		t.Fatalf("expected successful first and third result: %#v", results)
	}
	if results[1].Err == nil {
		t.Fatal("expected per-item validation error for app id 0")
	}
}

func TestGetAppDetailsBatchHonorsContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}}, &recordingTransport{statuses: []int{http.StatusOK}}, 1024)

	_, err := service.GetAppDetailsBatch(ctx, []uint32{550}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
