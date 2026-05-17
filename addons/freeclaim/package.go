package freeclaim

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gofurry/steam-go/web/storefront"
)

type ResolveFreePackagesOptions struct {
	CountryCode string
	Language    string
}

type FreePackage struct {
	AppID      uint32
	PackageID  uint32
	Title      string
	OptionText string
}

type packageGroup struct {
	Name  string             `json:"name"`
	Title string             `json:"title"`
	Subs  []packageGroupSubs `json:"subs"`
}

type packageGroupSubs struct {
	PackageID                uint32 `json:"packageid"`
	PercentSavingsText       string `json:"percent_savings_text"`
	OptionText               string `json:"option_text"`
	IsFreeLicense            bool   `json:"is_free_license"`
	PriceInCentsWithDiscount int    `json:"price_in_cents_with_discount"`
}

func (c *Client) ResolveFreePackages(ctx context.Context, appID uint32, opts *ResolveFreePackagesOptions) ([]FreePackage, error) {
	if appID == 0 {
		return nil, &Error{Code: ErrorCodeRequestBuild, Op: "resolve_free_packages", Message: "app id must be greater than zero"}
	}

	appDetails, err := c.storefront.GetAppDetails(ctx, appID, &storefront.GetAppDetailsOptions{
		CountryCode: fieldOrEmpty(opts, func(value *ResolveFreePackagesOptions) string { return value.CountryCode }),
		Language:    fieldOrEmpty(opts, func(value *ResolveFreePackagesOptions) string { return value.Language }),
	})
	if err != nil {
		return nil, err
	}

	result, ok := appDetails[strconv.FormatUint(uint64(appID), 10)]
	if !ok {
		return nil, &Error{Code: ErrorCodeVerify, Op: "resolve_free_packages", Message: "app details response missing target app"}
	}
	if !result.Success {
		return nil, &Error{Code: ErrorCodeVerify, Op: "resolve_free_packages", Message: "app details returned success=false"}
	}

	return parseFreePackages(appID, result.Data.Name, result.Data.PackageGroups)
}

func parseFreePackages(appID uint32, appName string, raw json.RawMessage) ([]FreePackage, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var groups []packageGroup
	if err := json.Unmarshal(raw, &groups); err != nil {
		return nil, &Error{Code: ErrorCodeDecode, Op: "resolve_free_packages", Message: "decode package_groups failed", Err: err}
	}

	packages := make([]FreePackage, 0)
	seen := make(map[uint32]struct{})
	for _, group := range groups {
		for _, sub := range group.Subs {
			if !looksLikeFreePackage(sub) {
				continue
			}
			if _, ok := seen[sub.PackageID]; ok {
				continue
			}
			seen[sub.PackageID] = struct{}{}
			packages = append(packages, FreePackage{
				AppID:      appID,
				PackageID:  sub.PackageID,
				Title:      appName,
				OptionText: strings.TrimSpace(sub.OptionText),
			})
		}
	}
	return packages, nil
}

func looksLikeFreePackage(sub packageGroupSubs) bool {
	if sub.PackageID == 0 || sub.PriceInCentsWithDiscount != 0 {
		return false
	}
	if sub.IsFreeLicense {
		return true
	}
	percentSavings := strings.ToLower(strings.TrimSpace(sub.PercentSavingsText))
	optionText := strings.ToLower(strings.TrimSpace(sub.OptionText))
	return strings.Contains(percentSavings, "100") || strings.Contains(optionText, "free")
}

func fieldOrEmpty[T any](value *T, getter func(*T) string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(getter(value))
}
