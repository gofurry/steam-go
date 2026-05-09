package wishlistservice

import "encoding/json"

// GetWishlistResponse matches IWishlistService/GetWishlist/v1.
type GetWishlistResponse struct {
	Response WishlistPayload `json:"response"`
}

// WishlistPayload is the top-level wishlist payload.
type WishlistPayload struct {
	Items []WishlistItem `json:"items"`
}

// WishlistItem matches one wishlist row.
type WishlistItem struct {
	AppID     uint32 `json:"appid"`
	Priority  int    `json:"priority"`
	DateAdded int64  `json:"date_added"`
}

// GetWishlistItemCountResponse matches IWishlistService/GetWishlistItemCount/v1.
type GetWishlistItemCountResponse struct {
	Response WishlistItemCountPayload `json:"response"`
}

// WishlistItemCountPayload is the wishlist count payload.
type WishlistItemCountPayload struct {
	Count int `json:"count"`
}

// GetWishlistItemsOnSaleResponse matches IWishlistService/GetWishlistItemsOnSale/v1.
type GetWishlistItemsOnSaleResponse struct {
	Response WishlistItemsOnSalePayload `json:"response"`
}

// WishlistItemsOnSalePayload is the top-level wishlist on-sale payload.
type WishlistItemsOnSalePayload struct {
	Items            []WishlistItemsOnSaleItem `json:"items"`
	TotalItemsOnSale int                       `json:"total_items_on_sale"`
}

// WishlistItemsOnSaleItem matches one on-sale wishlist row.
type WishlistItemsOnSaleItem struct {
	AppID     uint32          `json:"appid"`
	StoreItem json.RawMessage `json:"store_item"`
}
