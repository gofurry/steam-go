package storefront

import (
	"context"
	"sync"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
)

const defaultBatchMaxConcurrent = 4

// ListAppReviewsOptions controls Storefront review pagination.
type ListAppReviewsOptions struct {
	Query    GetAppReviewsOptions
	MaxPages int
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
