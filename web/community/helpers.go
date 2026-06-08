package community

import (
	"context"
	"strconv"

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

// JoinedInventoryItem pairs one owned asset row with its matching description.
type JoinedInventoryItem struct {
	Asset       InventoryAsset
	Description *InventoryItem
}

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

// JoinInventoryDescriptions pairs response assets with descriptions by appid,
// classid, and instanceid while preserving the original asset order.
//
// Missing descriptions are represented by a nil Description pointer. The helper
// is local-only and does not perform network requests, login, pricing, market
// lookups, trading checks, or account automation.
func JoinInventoryDescriptions(resp InventoryResponse) []JoinedInventoryItem {
	if len(resp.Assets) == 0 {
		return nil
	}

	descriptions := make(map[string]*InventoryItem, len(resp.Descriptions))
	for i := range resp.Descriptions {
		description := &resp.Descriptions[i]
		key := inventoryDescriptionKey(description.AppID, description.ClassID, description.InstanceID)
		if _, exists := descriptions[key]; !exists {
			descriptions[key] = description
		}
	}

	joined := make([]JoinedInventoryItem, len(resp.Assets))
	for i, asset := range resp.Assets {
		joined[i] = JoinedInventoryItem{
			Asset:       asset,
			Description: descriptions[inventoryDescriptionKey(asset.AppID, asset.ClassID, asset.InstanceID)],
		}
	}
	return joined
}

func inventoryDescriptionKey(appID uint32, classID, instanceID string) string {
	return strconv.FormatUint(uint64(appID), 10) + "\x00" + classID + "\x00" + instanceID
}
