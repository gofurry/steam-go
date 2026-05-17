package freeclaim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ownershipEnvelope struct {
	OwnedApps []uint32 `json:"rgOwnedApps"`
}

func (c *Client) IsAppOwned(ctx context.Context, provider CookieJarProvider, appID uint32) (bool, error) {
	jar, err := c.resolveProviderJar(ctx, provider, "is_app_owned")
	if err != nil {
		return false, err
	}
	return c.isAppOwnedWithJar(ctx, jar, appID)
}

func (c *Client) isAppOwnedWithJar(ctx context.Context, jar http.CookieJar, appID uint32) (bool, error) {
	if appID == 0 {
		return false, &Error{Code: ErrorCodeRequestBuild, Op: "is_app_owned", Message: "app id must be greater than zero"}
	}

	requestURL := resolveURL(c.storeBaseURL, "/dynamicstore/userdata/")
	query := requestURL.Query()
	query.Set("t", strconv.FormatInt(time.Now().UnixMilli(), 10))
	requestURL.RawQuery = query.Encode()

	result, err := c.doRequestWithJar(ctx, jar, http.MethodGet, requestURL.String(), nil, "", map[string]string{
		"Accept": "application/json",
	}, "is_app_owned")
	if err != nil {
		return false, err
	}
	if result.Block != nil {
		return false, &Error{Code: ErrorCodeVerify, Op: "is_app_owned", Message: result.Block.Message}
	}
	if endedOnLoginPage(result.FinalURL) {
		return false, &Error{Code: ErrorCodeVerify, Op: "is_app_owned", Message: "request ended on a login page"}
	}
	if result.StatusCode < http.StatusOK || result.StatusCode >= http.StatusMultipleChoices {
		return false, &Error{Code: ErrorCodeHTTPStatus, Op: "is_app_owned", Message: fmt.Sprintf("unexpected status %d: %s", result.StatusCode, strings.TrimSpace(string(result.Body)))}
	}

	var envelope ownershipEnvelope
	if err := json.Unmarshal(result.Body, &envelope); err != nil {
		return false, &Error{Code: ErrorCodeDecode, Op: "is_app_owned", Message: "decode dynamicstore userdata failed", Err: err}
	}
	for _, ownedAppID := range envelope.OwnedApps {
		if ownedAppID == appID {
			return true, nil
		}
	}
	return false, nil
}
