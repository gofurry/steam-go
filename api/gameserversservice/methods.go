package gameserversservice

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofurry/steam-go/internal/endpoint"
	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/request"
	"github.com/gofurry/steam-go/internal/response"
)

// GetServerListOptions controls optional GetServerList query parameters.
type GetServerListOptions struct {
	// Filter is a Master Server style \key\value filter string.
	//
	// Use NewFilter for common filters. Raw strings are accepted for advanced
	// upstream filters and validated before the request is sent.
	Filter string
	// Limit caps the number of returned server candidates. Zero omits the
	// parameter so upstream defaults apply.
	Limit uint32
}

// GetServerList returns server candidates discovered by Steam's server list endpoint.
//
// The endpoint returns candidate servers from Steam's Web API side. For strict
// real-time data, query each returned server with A2S afterward.
func (s *Service) GetServerList(ctx context.Context, opts *GetServerListOptions) (GetServerListResponse, error) {
	body, err := s.GetServerListRaw(ctx, opts)
	if err != nil {
		return GetServerListResponse{}, err
	}
	return response.DecodeJSON[GetServerListResponse](body)
}

// GetServerListRaw returns the raw JSON response body for server discovery.
func (s *Service) GetServerListRaw(ctx context.Context, opts *GetServerListOptions) ([]byte, error) {
	query, err := buildGetServerListQuery(opts)
	if err != nil {
		return nil, err
	}

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.GameServersServiceGetServerList,
		Query:  query,
	})
}

func buildGetServerListQuery(opts *GetServerListOptions) (url.Values, error) {
	query := url.Values{}
	if opts == nil {
		return query, nil
	}

	filter := strings.TrimSpace(opts.Filter)
	if filter != "" {
		if err := validateFilterString(filter); err != nil {
			return nil, err
		}
		query.Set("filter", filter)
	}
	if opts.Limit > 0 {
		query.Set("limit", strconv.FormatUint(uint64(opts.Limit), 10))
	}
	return query, nil
}

func validateFilterString(filter string) error {
	if strings.ContainsRune(filter, '\x00') {
		return requestBuildError("filter must not contain NUL bytes")
	}
	if !strings.HasPrefix(filter, `\`) {
		return requestBuildError(`filter must start with "\"`)
	}

	parts := strings.Split(filter, `\`)
	if len(parts) < 3 || parts[0] != "" || (len(parts)-1)%2 != 0 {
		return requestBuildError(`filter must use continuous "\key\value" pairs`)
	}

	hasNotEmpty := false
	hasNoPlayers := false
	for i := 1; i < len(parts); i += 2 {
		key := parts[i]
		value := parts[i+1]
		if err := validateFilterPart("filter key", key); err != nil {
			return err
		}
		if err := validateFilterPart("filter value", value); err != nil {
			return err
		}
		if err := validateKnownFilterValue(key, value); err != nil {
			return err
		}

		switch key {
		case "empty":
			hasNotEmpty = true
		case "noplayers":
			hasNoPlayers = true
		}
	}
	if hasNotEmpty && hasNoPlayers {
		return requestBuildError(`filter must not combine "empty" and "noplayers"`)
	}
	return nil
}

func validateFilterPart(label, value string) error {
	if value == "" {
		return requestBuildError(label + " must not be empty")
	}
	if strings.TrimSpace(value) != value {
		return requestBuildError(label + " must not have leading or trailing whitespace")
	}
	if strings.ContainsRune(value, '\x00') {
		return requestBuildError(label + " must not contain NUL bytes")
	}
	return nil
}

func validateKnownFilterValue(key, value string) error {
	switch key {
	case "appid", "napp":
		return validatePositiveUint32Filter(key, value)
	case "secure", "linux", "empty", "full", "noplayers", "proxy", "white", "dedicated", "collapse_addr_hash":
		if value != "1" {
			return requestBuildError(fmt.Sprintf("filter %s must be 1", key))
		}
	case "password":
		if value != "0" && value != "1" {
			return requestBuildError("filter password must be 0 or 1")
		}
	case "type":
		switch value {
		case ServerTypeDedicated, ServerTypeListen, ServerTypeSourceTV:
		default:
			return requestBuildError("filter type must be d, l, or p")
		}
	}
	return nil
}

func validatePositiveUint32Filter(key, value string) error {
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil || parsed == 0 {
		return requestBuildError(fmt.Sprintf("filter %s must be a positive uint32", key))
	}
	return nil
}

func requestBuildError(message string) error {
	return sdkerrors.New(sdkerrors.KindRequestBuild, 0, message, nil, nil)
}
