package market

import (
	"context"
	"sync"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
)

const defaultBatchMaxConcurrent = 4

// PriceOverviewBatchItem identifies one market price overview lookup.
type PriceOverviewBatchItem struct {
	AppID          uint32
	MarketHashName string
}

// GetPriceOverviewBatchOptions controls batch market price overview lookups.
type GetPriceOverviewBatchOptions struct {
	Currency      int
	MaxConcurrent int
}

// PriceOverviewBatchResult is one market price overview batch result.
type PriceOverviewBatchResult struct {
	Item     PriceOverviewBatchItem
	Response PriceOverviewResponse
	Err      error
}

// GetPriceOverviewBatch fetches market price overviews while preserving input order and per-item errors.
func (s *Service) GetPriceOverviewBatch(ctx context.Context, items []PriceOverviewBatchItem, opts *GetPriceOverviewBatchOptions) ([]PriceOverviewBatchResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	maxConcurrent := defaultBatchMaxConcurrent
	query := GetPriceOverviewOptions{}
	if opts != nil {
		if opts.MaxConcurrent < 0 {
			return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "max concurrent must not be negative", nil, nil)
		}
		if opts.MaxConcurrent > 0 {
			maxConcurrent = opts.MaxConcurrent
		}
		query.Currency = opts.Currency
	}
	if len(items) == 0 {
		return nil, nil
	}

	results := make([]PriceOverviewBatchResult, len(items))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for worker := 0; worker < maxConcurrent; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				item := items[idx]
				resp, err := s.GetPriceOverview(ctx, item.AppID, item.MarketHashName, &query)
				results[idx] = PriceOverviewBatchResult{Item: item, Response: resp, Err: err}
			}
		}()
	}

	for idx := range items {
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
