package storefront

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/request"
	"github.com/gofurry/steam-go/internal/response"
	itraffic "github.com/gofurry/steam-go/internal/traffic"
	"github.com/gofurry/steam-go/internal/webendpoint"
)

const (
	defaultReviewFilter       = "recent"
	defaultReviewLanguage     = "all"
	defaultReviewCursor       = "*"
	defaultReviewType         = "all"
	defaultReviewPurchaseType = "all"
	defaultReviewNumPerPage   = 100
	defaultReviewDayRange     = 365
)

var (
	validReviewFilters = map[string]struct{}{
		"recent":  {},
		"updated": {},
		"all":     {},
	}
	validReviewTypes = map[string]struct{}{
		"all":      {},
		"positive": {},
		"negative": {},
	}
	validPurchaseTypes = map[string]struct{}{
		"all":                {},
		"steam":              {},
		"non_steam_purchase": {},
	}
)

// GetAppDetailsOptions controls optional query parameters for app details.
type GetAppDetailsOptions struct {
	CountryCode string
	Language    string
	Filters     []string
}

// GetPackageDetailsOptions controls optional query parameters for package details.
type GetPackageDetailsOptions struct {
	CountryCode string
	Language    string
}

// GetAppReviewsOptions controls optional query parameters for app reviews.
type GetAppReviewsOptions struct {
	Filter                 string
	Language               string
	DayRange               int
	Cursor                 string
	ReviewType             string
	PurchaseType           string
	NumPerPage             int
	FilterOfftopicActivity *int
}

// GetAppDetails returns Storefront app details for one AppID.
func (s *Service) GetAppDetails(ctx context.Context, appID uint32, opts *GetAppDetailsOptions) (AppDetailsEnvelope, error) {
	body, err := s.GetAppDetailsRaw(ctx, appID, opts)
	if err != nil {
		return nil, err
	}
	return response.DecodeJSON[AppDetailsEnvelope](body)
}

// GetAppDetailsRaw returns the raw JSON response body for app details.
func (s *Service) GetAppDetailsRaw(ctx context.Context, appID uint32, opts *GetAppDetailsOptions) ([]byte, error) {
	if appID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}

	query := url.Values{}
	query.Set("appids", strconv.FormatUint(uint64(appID), 10))
	if opts != nil {
		if countryCode := strings.TrimSpace(opts.CountryCode); countryCode != "" {
			query.Set("cc", countryCode)
		}
		if language := strings.TrimSpace(opts.Language); language != "" {
			query.Set("l", language)
		}
		filters := normalizeFilters(opts.Filters)
		if filters != "" {
			query.Set("filters", filters)
		}
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method:       http.MethodGet,
		Path:         webendpoint.StoreAppDetailsPath,
		Query:        query,
		TrafficClass: itraffic.ClassPublicStorePage,
	})
}

// GetPackageDetails returns Storefront package details for one package ID.
func (s *Service) GetPackageDetails(ctx context.Context, packageID uint32, opts *GetPackageDetailsOptions) (PackageDetailsEnvelope, error) {
	body, err := s.GetPackageDetailsRaw(ctx, packageID, opts)
	if err != nil {
		return nil, err
	}
	return response.DecodeJSON[PackageDetailsEnvelope](body)
}

// GetPackageDetailsRaw returns the raw JSON response body for package details.
func (s *Service) GetPackageDetailsRaw(ctx context.Context, packageID uint32, opts *GetPackageDetailsOptions) ([]byte, error) {
	if packageID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "package id must be greater than zero", nil, nil)
	}

	query := url.Values{}
	query.Set("packageids", strconv.FormatUint(uint64(packageID), 10))
	if opts != nil {
		if countryCode := strings.TrimSpace(opts.CountryCode); countryCode != "" {
			query.Set("cc", countryCode)
		}
		if language := strings.TrimSpace(opts.Language); language != "" {
			query.Set("l", language)
		}
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method:       http.MethodGet,
		Path:         webendpoint.StorePackageDetailsPath,
		Query:        query,
		TrafficClass: itraffic.ClassPublicStorePage,
	})
}

// GetAppReviews returns Storefront reviews for one AppID.
func (s *Service) GetAppReviews(ctx context.Context, appID uint32, opts *GetAppReviewsOptions) (AppReviewsResponse, error) {
	body, err := s.GetAppReviewsRaw(ctx, appID, opts)
	if err != nil {
		return AppReviewsResponse{}, err
	}
	return response.DecodeJSON[AppReviewsResponse](body)
}

// GetAppReviewsRaw returns the raw JSON response body for app reviews.
func (s *Service) GetAppReviewsRaw(ctx context.Context, appID uint32, opts *GetAppReviewsOptions) ([]byte, error) {
	if appID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}

	query := url.Values{}
	query.Set("json", "1")

	filter := defaultReviewFilter
	language := defaultReviewLanguage
	dayRange := defaultReviewDayRange
	cursor := defaultReviewCursor
	reviewType := defaultReviewType
	purchaseType := defaultReviewPurchaseType
	numPerPage := defaultReviewNumPerPage

	if opts != nil {
		var err error
		filter, err = normalizeReviewEnum("filter", opts.Filter, validReviewFilters, defaultReviewFilter)
		if err != nil {
			return nil, err
		}
		language = strings.TrimSpace(opts.Language)
		if language == "" {
			language = defaultReviewLanguage
		}
		if opts.DayRange != 0 {
			if opts.DayRange < 1 || opts.DayRange > 365 {
				return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "day range must be between 1 and 365", nil, nil)
			}
			dayRange = opts.DayRange
		}
		cursor = strings.TrimSpace(opts.Cursor)
		if cursor == "" {
			cursor = defaultReviewCursor
		}
		reviewType, err = normalizeReviewEnum("review type", opts.ReviewType, validReviewTypes, defaultReviewType)
		if err != nil {
			return nil, err
		}
		purchaseType, err = normalizeReviewEnum("purchase type", opts.PurchaseType, validPurchaseTypes, defaultReviewPurchaseType)
		if err != nil {
			return nil, err
		}
		if opts.NumPerPage != 0 {
			if opts.NumPerPage < 1 || opts.NumPerPage > 100 {
				return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "num per page must be between 1 and 100", nil, nil)
			}
			numPerPage = opts.NumPerPage
		}
		if opts.FilterOfftopicActivity != nil {
			if *opts.FilterOfftopicActivity != 0 && *opts.FilterOfftopicActivity != 1 {
				return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "filter offtopic activity must be 0 or 1", nil, nil)
			}
			query.Set("filter_offtopic_activity", strconv.Itoa(*opts.FilterOfftopicActivity))
		}
	}

	query.Set("filter", filter)
	query.Set("language", language)
	query.Set("day_range", strconv.Itoa(dayRange))
	query.Set("cursor", cursor)
	query.Set("review_type", reviewType)
	query.Set("purchase_type", purchaseType)
	query.Set("num_per_page", strconv.Itoa(numPerPage))

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method:       http.MethodGet,
		Path:         fmt.Sprintf("%s/%d", webendpoint.StoreAppReviewsPath, appID),
		Query:        query,
		TrafficClass: itraffic.ClassPublicStorePage,
	})
}

func normalizeFilters(filters []string) string {
	if len(filters) == 0 {
		return ""
	}
	normalized := make([]string, 0, len(filters))
	for _, filter := range filters {
		filter = strings.TrimSpace(filter)
		if filter == "" {
			continue
		}
		normalized = append(normalized, filter)
	}
	return strings.Join(normalized, ",")
}

func normalizeReviewEnum(name, value string, allowed map[string]struct{}, fallback string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	if _, ok := allowed[value]; !ok {
		return "", sdkerrors.New(sdkerrors.KindRequestBuild, 0, fmt.Sprintf("unsupported %s %q", name, value), nil, nil)
	}
	return value, nil
}
