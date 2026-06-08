package storefront

import (
	"context"
	"errors"
	"sync"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
)

const defaultBatchMaxConcurrent = 4

// ListAppReviewsOptions controls Storefront review pagination.
type ListAppReviewsOptions struct {
	Query    GetAppReviewsOptions
	MaxPages int
}

// CollectAppReviewsOptions controls bounded Storefront review collection.
type CollectAppReviewsOptions struct {
	Query      GetAppReviewsOptions
	MaxPages   int
	MaxReviews int
}

// AppReviewsPage is one Storefront reviews page.
type AppReviewsPage struct {
	Page     int
	Cursor   string
	Response AppReviewsResponse
	Reviews  []AppReview
}

// AppReviewsPageHandler handles one Storefront reviews page.
type AppReviewsPageHandler func(AppReviewsPage) error

// AppReviewsCollection is the accumulated result from CollectAppReviews.
type AppReviewsCollection struct {
	Reviews      []AppReview
	Pages        int
	LastCursor   string
	QuerySummary StoreReviewQuerySummary
	Truncated    bool
}

// GetAppDetailsBatchOptions controls batch app details lookups.
type GetAppDetailsBatchOptions struct {
	Query         GetAppDetailsOptions
	MaxConcurrent int
}

// AppDetailsBatchResult is one app details batch result.
type AppDetailsBatchResult struct {
	AppID    uint32
	Response AppDetailsEnvelope
	Err      error
}

// ListAppReviews walks Storefront app review pages without accumulating all reviews in memory.
func (s *Service) ListAppReviews(ctx context.Context, appID uint32, opts *ListAppReviewsOptions, handler AppReviewsPageHandler) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if handler == nil {
		return sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app reviews handler is required", nil, nil)
	}
	if opts != nil && opts.MaxPages < 0 {
		return sdkerrors.New(sdkerrors.KindRequestBuild, 0, "max pages must not be negative", nil, nil)
	}

	query := GetAppReviewsOptions{}
	maxPages := 0
	if opts != nil {
		query = cloneGetAppReviewsOptions(opts.Query)
		maxPages = opts.MaxPages
	}
	cursor := query.Cursor
	if cursor == "" {
		cursor = defaultReviewCursor
	}

	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query.Cursor = cursor
		resp, err := s.GetAppReviews(ctx, appID, &query)
		if err != nil {
			return err
		}
		if err := handler(AppReviewsPage{
			Page:     page,
			Cursor:   cursor,
			Response: resp,
			Reviews:  resp.Reviews,
		}); err != nil {
			return err
		}

		nextCursor := resp.Cursor
		if len(resp.Reviews) == 0 || nextCursor == "" || nextCursor == cursor {
			return nil
		}
		cursor = nextCursor
	}
	return nil
}

var errStopReviewCollection = errors.New("stop app review collection")

// CollectAppReviews accumulates Storefront app review pages into memory with an
// explicit caller-provided bound.
//
// Callers must set MaxPages or MaxReviews. This helper is read-only, but it can
// still issue multiple Storefront requests, so production callers should pair it
// with conservative traffic policy, rate limits, and context deadlines.
func (s *Service) CollectAppReviews(ctx context.Context, appID uint32, opts *CollectAppReviewsOptions) (AppReviewsCollection, error) {
	if opts == nil {
		return AppReviewsCollection{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app reviews collection limits are required", nil, nil)
	}
	if opts.MaxPages < 0 {
		return AppReviewsCollection{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "max pages must not be negative", nil, nil)
	}
	if opts.MaxReviews < 0 {
		return AppReviewsCollection{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "max reviews must not be negative", nil, nil)
	}
	if opts.MaxPages == 0 && opts.MaxReviews == 0 {
		return AppReviewsCollection{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "max pages or max reviews is required", nil, nil)
	}

	query := cloneGetAppReviewsOptions(opts.Query)
	if opts.MaxReviews > 0 && query.NumPerPage == 0 && opts.MaxReviews < defaultReviewNumPerPage {
		query.NumPerPage = opts.MaxReviews
	}

	collection := AppReviewsCollection{}
	err := s.ListAppReviews(ctx, appID, &ListAppReviewsOptions{
		Query:    query,
		MaxPages: opts.MaxPages,
	}, func(page AppReviewsPage) error {
		collection.Pages = page.Page
		collection.LastCursor = page.Response.Cursor
		collection.QuerySummary = page.Response.QuerySummary

		pageReviews := page.Reviews
		if opts.MaxReviews > 0 {
			remaining := opts.MaxReviews - len(collection.Reviews)
			if remaining <= 0 {
				collection.Truncated = true
				return errStopReviewCollection
			}
			if len(pageReviews) > remaining {
				collection.Reviews = append(collection.Reviews, pageReviews[:remaining]...)
				collection.Truncated = true
				return errStopReviewCollection
			}
		}

		collection.Reviews = append(collection.Reviews, pageReviews...)
		if opts.MaxReviews > 0 && len(collection.Reviews) >= opts.MaxReviews {
			collection.Truncated = true
			return errStopReviewCollection
		}
		if opts.MaxPages > 0 && page.Page >= opts.MaxPages && len(page.Reviews) > 0 && page.Response.Cursor != "" && page.Response.Cursor != page.Cursor {
			collection.Truncated = true
		}
		return nil
	})
	if errors.Is(err, errStopReviewCollection) {
		return collection, nil
	}
	return collection, err
}

// GetAppDetailsBatch fetches app details while preserving input order and per-item errors.
func (s *Service) GetAppDetailsBatch(ctx context.Context, appIDs []uint32, opts *GetAppDetailsBatchOptions) ([]AppDetailsBatchResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	maxConcurrent := defaultBatchMaxConcurrent
	query := GetAppDetailsOptions{}
	if opts != nil {
		if opts.MaxConcurrent < 0 {
			return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "max concurrent must not be negative", nil, nil)
		}
		if opts.MaxConcurrent > 0 {
			maxConcurrent = opts.MaxConcurrent
		}
		query = cloneGetAppDetailsOptions(opts.Query)
	}
	if len(appIDs) == 0 {
		return nil, nil
	}

	results := make([]AppDetailsBatchResult, len(appIDs))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for worker := 0; worker < maxConcurrent; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				appID := appIDs[idx]
				resp, err := s.GetAppDetails(ctx, appID, &query)
				results[idx] = AppDetailsBatchResult{AppID: appID, Response: resp, Err: err}
			}
		}()
	}

	for idx := range appIDs {
		if err := ctx.Err(); err != nil {
			close(jobs)
			wg.Wait()
			return results, err
		}
		select {
		case jobs <- idx:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return results, ctx.Err()
		}
	}
	close(jobs)
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return results, err
	}
	return results, nil
}

func cloneGetAppReviewsOptions(src GetAppReviewsOptions) GetAppReviewsOptions {
	return src
}

func cloneGetAppDetailsOptions(src GetAppDetailsOptions) GetAppDetailsOptions {
	if len(src.Filters) == 0 {
		return src
	}
	cloned := src
	cloned.Filters = append([]string(nil), src.Filters...)
	return cloned
}
