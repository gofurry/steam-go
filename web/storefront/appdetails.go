package storefront

import (
	"context"
	"fmt"
	"strconv"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
)

// AppDetailsMatch identifies the result for a requested AppID while preserving
// Steam's original response key, which need not be numeric or equal to the AppID.
//
// Warning: ResponseKey is upstream metadata, not a substitute for the verified
// application identity in RequestedAppID and Result.Data.SteamAppID.
type AppDetailsMatch struct {
	RequestedAppID uint32
	ResponseKey    string
	Result         AppDetailsResult
}

// ResolveAppDetails locally resolves a requested AppID using data.steam_appid.
// It performs no network requests and does not modify the upstream envelope.
//
// Warning: AppDetails is an undocumented, volatile Steam surface. An invalid
// direct key, missing identity, unsuccessful result, or multiple identity matches
// fails closed. A valid alternate entry cannot override an invalid direct entry.
// No DLC, parent, child, or other application relationship is inferred.
func ResolveAppDetails(envelope AppDetailsEnvelope, appID uint32) (AppDetailsMatch, error) {
	if appID == 0 {
		return AppDetailsMatch{}, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "app id must be greater than zero", nil, nil)
	}

	directKey := strconv.FormatUint(uint64(appID), 10)
	if direct, ok := envelope[directKey]; ok {
		switch {
		case !direct.Success:
			return AppDetailsMatch{}, sdkerrors.New(sdkerrors.KindAPIResponse, 0, fmt.Sprintf("appdetails direct entry for app id %d was not successful", appID), nil, nil)
		case direct.Data.SteamAppID == 0:
			return AppDetailsMatch{}, sdkerrors.New(sdkerrors.KindAPIResponse, 0, fmt.Sprintf("appdetails direct entry for app id %d has no application identity", appID), nil, nil)
		case direct.Data.SteamAppID != appID:
			return AppDetailsMatch{}, sdkerrors.New(sdkerrors.KindAPIResponse, 0, fmt.Sprintf("appdetails direct entry has an identity conflict for app id %d", appID), nil, nil)
		}
	}

	var match AppDetailsMatch
	count := 0
	for key, result := range envelope {
		if result.Data.SteamAppID != appID {
			continue
		}
		count++
		if count > 1 {
			return AppDetailsMatch{}, sdkerrors.New(sdkerrors.KindAPIResponse, 0, fmt.Sprintf("appdetails response has ambiguous identity for app id %d", appID), nil, nil)
		}
		match = AppDetailsMatch{RequestedAppID: appID, ResponseKey: key, Result: result}
	}
	if count == 0 {
		return AppDetailsMatch{}, sdkerrors.New(sdkerrors.KindAPIResponse, 0, fmt.Sprintf("appdetails response did not contain requested app id %d", appID), nil, nil)
	}
	if !match.Result.Success {
		return AppDetailsMatch{}, sdkerrors.New(sdkerrors.KindAPIResponse, 0, fmt.Sprintf("appdetails result for app id %d was not successful", appID), nil, nil)
	}
	return match, nil
}

// GetResolvedAppDetails fetches and resolves one AppID through GetAppDetails and
// ResolveAppDetails, using data.steam_appid as the requested application identity.
//
// Warning: This undocumented, volatile Steam surface can return mismatched keys.
// Direct-key conflicts, missing identity, unsuccessful results, and multiple
// matches fail closed. DLC, parent, and child relationships are never inferred.
// ResponseKey preserves Steam's original key without normalization.
func (s *Service) GetResolvedAppDetails(ctx context.Context, appID uint32, opts *GetAppDetailsOptions) (AppDetailsMatch, error) {
	envelope, err := s.GetAppDetails(ctx, appID, opts)
	if err != nil {
		return AppDetailsMatch{}, err
	}
	return ResolveAppDetails(envelope, appID)
}
