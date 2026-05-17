package freeclaim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

const (
	defaultSearchCount = 50
	defaultSearchOS    = "win"
	defaultSearchSNR   = "1_7_7_7000_7"
)

type SearchPromotionsOptions struct {
	Query string
	Start int
	Count int
	OS    string
	SNR   string
}

type FreePromotion struct {
	AppID         uint32
	Title         string
	StoreURL      string
	CapsuleURL    string
	OriginalPrice string
	FinalPrice    string
	Discount      string
}

type searchResultsEnvelope struct {
	Success     int    `json:"success"`
	ResultsHTML string `json:"results_html"`
	TotalCount  int    `json:"total_count"`
}

func (c *Client) SearchPromotions(ctx context.Context, opts *SearchPromotionsOptions) ([]FreePromotion, error) {
	query := url.Values{}
	query.Set("query", "")
	query.Set("start", "0")
	query.Set("count", strconv.Itoa(defaultSearchCount))
	query.Set("dynamic_data", "")
	query.Set("force_infinite", "1")
	query.Set("specials", "1")
	query.Set("maxprice", "free")
	query.Set("os", defaultSearchOS)
	query.Set("snr", defaultSearchSNR)
	query.Set("infinite", "1")

	if opts != nil {
		query.Set("query", strings.TrimSpace(opts.Query))
		if opts.Start < 0 {
			return nil, &Error{Code: ErrorCodeRequestBuild, Op: "search_promotions", Message: "start must not be negative"}
		}
		if opts.Start > 0 {
			query.Set("start", strconv.Itoa(opts.Start))
		}
		if opts.Count != 0 {
			if opts.Count < 1 {
				return nil, &Error{Code: ErrorCodeRequestBuild, Op: "search_promotions", Message: "count must be greater than zero"}
			}
			query.Set("count", strconv.Itoa(opts.Count))
		}
		if osName := strings.TrimSpace(opts.OS); osName != "" {
			query.Set("os", osName)
		}
		if snr := strings.TrimSpace(opts.SNR); snr != "" {
			query.Set("snr", snr)
		}
	}

	rawURL := resolveURL(c.storeBaseURL, "/search/results/")
	rawURL.RawQuery = query.Encode()
	result, err := c.doRequestWithJar(ctx, nil, http.MethodGet, rawURL.String(), nil, "", nil, "search_promotions")
	if err != nil {
		return nil, err
	}
	if result.Block != nil {
		return nil, &Error{Code: ErrorCodeVerify, Op: "search_promotions", Message: result.Block.Message}
	}
	if result.StatusCode < http.StatusOK || result.StatusCode >= http.StatusMultipleChoices {
		return nil, &Error{Code: ErrorCodeHTTPStatus, Op: "search_promotions", Message: fmt.Sprintf("unexpected status %d: %s", result.StatusCode, strings.TrimSpace(string(result.Body)))}
	}

	var envelope searchResultsEnvelope
	if err := json.Unmarshal(result.Body, &envelope); err != nil {
		return nil, &Error{Code: ErrorCodeDecode, Op: "search_promotions", Message: "decode search response failed", Err: err}
	}
	if envelope.Success != 1 {
		return nil, &Error{Code: ErrorCodeVerify, Op: "search_promotions", Message: fmt.Sprintf("store search returned success=%d", envelope.Success)}
	}
	return parseSearchResultsHTML(envelope.ResultsHTML, c.storeBaseURL)
}

func parseSearchResultsHTML(markup string, base *url.URL) ([]FreePromotion, error) {
	document, err := html.Parse(strings.NewReader("<html><body>" + markup + "</body></html>"))
	if err != nil {
		return nil, err
	}

	promotions := make([]FreePromotion, 0)
	visitNodes(document, func(node *html.Node) {
		if node.Type != html.ElementNode || node.Data != "a" || !hasClass(node, "search_result_row") {
			return
		}

		appID, ok := parseNodeAppID(node)
		if !ok {
			return
		}
		discountBlock := firstNode(node, func(child *html.Node) bool {
			return child.Type == html.ElementNode && hasClass(child, "discount_block")
		})
		if discountBlock == nil {
			return
		}
		if nodeAttr(discountBlock, "data-discount") != "100" || nodeAttr(discountBlock, "data-price-final") != "0" {
			return
		}
		originalPrice := textContent(firstNode(node, func(child *html.Node) bool {
			return child.Type == html.ElementNode && hasClass(child, "discount_original_price")
		}))
		if strings.TrimSpace(originalPrice) == "" {
			return
		}

		storeURL := nodeAttr(node, "href")
		if base != nil && storeURL != "" {
			if parsed, err := url.Parse(storeURL); err == nil {
				storeURL = base.ResolveReference(parsed).String()
			}
		}
		capsuleURL := ""
		if image := firstNode(node, func(child *html.Node) bool {
			return child.Type == html.ElementNode && child.Data == "img"
		}); image != nil {
			capsuleURL = nodeAttr(image, "src")
		}
		discount := textContent(firstNode(node, func(child *html.Node) bool {
			return child.Type == html.ElementNode && hasClass(child, "discount_pct")
		}))
		if discount == "" {
			discount = "-" + nodeAttr(discountBlock, "data-discount") + "%"
		}
		finalPrice := textContent(firstNode(node, func(child *html.Node) bool {
			return child.Type == html.ElementNode && hasClass(child, "discount_final_price")
		}))
		title := textContent(firstNode(node, func(child *html.Node) bool {
			return child.Type == html.ElementNode && hasClass(child, "title")
		}))

		promotions = append(promotions, FreePromotion{
			AppID:         appID,
			Title:         title,
			StoreURL:      storeURL,
			CapsuleURL:    capsuleURL,
			OriginalPrice: strings.TrimSpace(originalPrice),
			FinalPrice:    strings.TrimSpace(finalPrice),
			Discount:      strings.TrimSpace(discount),
		})
	})
	return promotions, nil
}

func parseNodeAppID(node *html.Node) (uint32, bool) {
	raw := strings.TrimSpace(nodeAttr(node, "data-ds-appid"))
	if raw == "" {
		return 0, false
	}
	if comma := strings.IndexByte(raw, ','); comma >= 0 {
		raw = raw[:comma]
	}
	value, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 32)
	if err != nil || value == 0 {
		return 0, false
	}
	return uint32(value), true
}
