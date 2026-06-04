package community

import (
	"context"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
)

// ListInventoryOptions controls Community inventory pagination.
type ListInventoryOptions struct {
	Query    GetInventoryOptions
	MaxPages int
}

// InventoryPage is one Community inventory page.
type InventoryPage struct {
	Page     int
	Cursor   string
	Response InventoryResponse
	Assets   []InventoryAsset
	Items    []InventoryItem
}

// InventoryPageHandler handles one Community inventory page.
type InventoryPageHandler func(InventoryPage) error

// ListInventory walks Community inventory pages without accumulating all items in memory.
//
// This helper is read-only. It does not perform login, refresh cookies, or
// guarantee access to private inventories.
func (s *Service) ListInventory(ctx context.Context, steamID string, appID uint32, contextID string, opts *ListInventoryOptions, handler InventoryPageHandler) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if handler == nil {
		return sdkerrors.New(sdkerrors.KindRequestBuild, 0, "inventory handler is required", nil, nil)
	}
	if opts != nil && opts.MaxPages < 0 {
		return sdkerrors.New(sdkerrors.KindRequestBuild, 0, "max pages must not be negative", nil, nil)
	}

	query := GetInventoryOptions{}
	maxPages := 0
	if opts != nil {
		query = opts.Query
		maxPages = opts.MaxPages
	}
	cursor := query.StartAssetID

	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query.StartAssetID = cursor
		resp, err := s.GetInventory(ctx, steamID, appID, contextID, &query)
		if err != nil {
			return err
		}
		if err := handler(InventoryPage{
			Page:     page,
			Cursor:   cursor,
			Response: resp,
			Assets:   resp.Assets,
			Items:    resp.Descriptions,
		}); err != nil {
			return err
		}

		nextCursor := resp.LastAssetID
		if !resp.MoreItems.Bool() || nextCursor == "" || nextCursor == cursor {
			return nil
		}
		cursor = nextCursor
	}
	return nil
}
