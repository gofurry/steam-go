package freeclaim

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type ClaimStatus string

const (
	ClaimStatusClaimed       ClaimStatus = "claimed"
	ClaimStatusAlreadyOwned  ClaimStatus = "already_owned"
	ClaimStatusLoginRequired ClaimStatus = "login_required"
	ClaimStatusRateLimited   ClaimStatus = "rate_limited"
	ClaimStatusUnknown       ClaimStatus = "unknown"
)

type ClaimPackageRequest struct {
	AppID     uint32
	PackageID uint32
}

type ClaimResult struct {
	Status    ClaimStatus
	AppID     uint32
	PackageID uint32
	Owned     bool
	Message   string
}

func (c *Client) ClaimPackage(ctx context.Context, provider CookieJarProvider, req ClaimPackageRequest) (ClaimResult, error) {
	jar, err := c.resolveProviderJar(ctx, provider, "claim_package")
	if err != nil {
		return ClaimResult{}, err
	}
	return c.claimPackageWithJar(ctx, jar, req)
}

func (c *Client) claimPackageWithJar(ctx context.Context, jar http.CookieJar, req ClaimPackageRequest) (ClaimResult, error) {
	if req.AppID == 0 {
		return ClaimResult{}, &Error{Code: ErrorCodeRequestBuild, Op: "claim_package", Message: "app id must be greater than zero"}
	}
	if req.PackageID == 0 {
		return ClaimResult{}, &Error{Code: ErrorCodeRequestBuild, Op: "claim_package", Message: "package id must be greater than zero"}
	}

	preOwned, preOwnedErr := c.isAppOwnedWithJar(ctx, jar, req.AppID)
	if preOwnedErr == nil && preOwned {
		return ClaimResult{
			Status:    ClaimStatusAlreadyOwned,
			AppID:     req.AppID,
			PackageID: req.PackageID,
			Owned:     true,
			Message:   "app already owned before claim",
		}, nil
	}

	appURL := resolveURL(c.storeBaseURL, "/app/"+strconv.FormatUint(uint64(req.AppID), 10)+"/")
	pageResult, err := c.doRequestWithJar(ctx, jar, http.MethodGet, appURL.String(), nil, "", nil, "claim_package")
	if err != nil {
		return ClaimResult{}, err
	}
	if pageResult.Block != nil {
		return ClaimResult{
			Status:    ClaimStatusRateLimited,
			AppID:     req.AppID,
			PackageID: req.PackageID,
			Message:   pageResult.Block.Message,
		}, nil
	}
	if endedOnLoginPage(pageResult.FinalURL) {
		return ClaimResult{
			Status:    ClaimStatusLoginRequired,
			AppID:     req.AppID,
			PackageID: req.PackageID,
			Message:   "store app page redirected to login",
		}, nil
	}
	if pageResult.StatusCode == http.StatusForbidden || pageResult.StatusCode == http.StatusTooManyRequests {
		return ClaimResult{
			Status:    ClaimStatusRateLimited,
			AppID:     req.AppID,
			PackageID: req.PackageID,
			Message:   fmt.Sprintf("store app page returned status %d", pageResult.StatusCode),
		}, nil
	}
	if pageResult.StatusCode < http.StatusOK || pageResult.StatusCode >= http.StatusMultipleChoices {
		return ClaimResult{}, &Error{Code: ErrorCodeHTTPStatus, Op: "claim_package", Message: fmt.Sprintf("unexpected status %d: %s", pageResult.StatusCode, strings.TrimSpace(string(pageResult.Body)))}
	}

	formValues, err := parseAddToCartForm(pageResult.Body, req.PackageID)
	if err != nil {
		return ClaimResult{}, err
	}
	if formValues.Get("action") == "" {
		formValues.Set("action", "add_to_cart")
	}
	formValues.Set("subid", strconv.FormatUint(uint64(req.PackageID), 10))
	if formValues.Get("sessionid") == "" {
		sessionID := storeCookieValue(jar, c.storeBaseURL, "sessionid")
		if sessionID == "" {
			return ClaimResult{}, &Error{Code: ErrorCodeVerify, Op: "claim_package", Message: "store sessionid cookie is missing"}
		}
		formValues.Set("sessionid", sessionID)
	}

	claimURL := resolveURL(c.storeBaseURL, "/checkout/addfreelicense")
	claimResult, err := c.doRequestWithJar(ctx, jar, http.MethodPost, claimURL.String(), strings.NewReader(formValues.Encode()), "application/x-www-form-urlencoded; charset=UTF-8", map[string]string{
		"Origin":           strings.TrimRight(c.storeBaseURL.String(), "/"),
		"Referer":          appURL.String(),
		"X-Requested-With": "XMLHttpRequest",
	}, "claim_package")
	if err != nil {
		return ClaimResult{}, err
	}

	outcome := ClaimResult{
		AppID:     req.AppID,
		PackageID: req.PackageID,
		Message:   classifyClaimMessage(claimResult.Body),
	}
	switch {
	case claimResult.Block != nil:
		outcome.Status = ClaimStatusRateLimited
		if outcome.Message == "" {
			outcome.Message = claimResult.Block.Message
		}
		return outcome, nil
	case endedOnLoginPage(claimResult.FinalURL):
		outcome.Status = ClaimStatusLoginRequired
		if outcome.Message == "" {
			outcome.Message = "claim request ended on login page"
		}
		return outcome, nil
	case claimResult.StatusCode == http.StatusForbidden || claimResult.StatusCode == http.StatusTooManyRequests:
		outcome.Status = ClaimStatusRateLimited
		if outcome.Message == "" {
			outcome.Message = fmt.Sprintf("claim request returned status %d", claimResult.StatusCode)
		}
		return outcome, nil
	case claimBodyContainsSuccess(claimResult.Body):
		outcome.Status = ClaimStatusClaimed
		outcome.Owned = true
		return outcome, nil
	case claimBodyContainsAlreadyOwned(claimResult.Body):
		outcome.Status = ClaimStatusAlreadyOwned
		outcome.Owned = true
		return outcome, nil
	}

	postOwned, postOwnedErr := c.isAppOwnedWithJar(ctx, jar, req.AppID)
	if postOwnedErr == nil && postOwned {
		outcome.Status = ClaimStatusClaimed
		outcome.Owned = true
		if outcome.Message == "" {
			outcome.Message = "ownership fallback confirmed app ownership after claim"
		}
		return outcome, nil
	}

	outcome.Status = ClaimStatusUnknown
	if outcome.Message == "" {
		outcome.Message = "claim response could not be classified"
	}
	return outcome, nil
}

func parseAddToCartForm(body []byte, packageID uint32) (url.Values, error) {
	document, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, &Error{Code: ErrorCodeParse, Op: "claim_package", Message: "parse app page html failed", Err: err}
	}

	formName := fmt.Sprintf("add_to_cart_%d", packageID)
	form := firstNode(document, func(node *html.Node) bool {
		if node.Type != html.ElementNode || node.Data != "form" {
			return false
		}
		return nodeAttr(node, "name") == formName || nodeAttr(node, "id") == formName
	})
	if form == nil {
		return nil, &Error{Code: ErrorCodeParse, Op: "claim_package", Message: "add_to_cart form not found"}
	}

	values := url.Values{}
	visitNodes(form, func(node *html.Node) {
		if node.Type != html.ElementNode || node.Data != "input" {
			return
		}
		if inputType := strings.ToLower(strings.TrimSpace(nodeAttr(node, "type"))); inputType != "" && inputType != "hidden" {
			return
		}
		name := strings.TrimSpace(nodeAttr(node, "name"))
		if name == "" {
			return
		}
		values.Set(name, nodeAttr(node, "value"))
	})
	return values, nil
}

func claimBodyContainsSuccess(body []byte) bool {
	lowered := strings.ToLower(string(body))
	return strings.Contains(lowered, "success!") ||
		strings.Contains(lowered, "成功！") ||
		strings.Contains(lowered, "已被绑定至您的 steam 帐户") ||
		strings.Contains(lowered, "已被绑定至您的 steam 账户")
}

func claimBodyContainsAlreadyOwned(body []byte) bool {
	lowered := strings.ToLower(string(body))
	return strings.Contains(lowered, "already own") ||
		strings.Contains(lowered, "already have") ||
		strings.Contains(lowered, "已拥有") ||
		strings.Contains(lowered, "已经拥有")
}

func classifyClaimMessage(body []byte) string {
	snippet := strings.TrimSpace(string(body))
	if snippet == "" {
		return ""
	}
	if len(snippet) > 160 {
		snippet = snippet[:160]
	}
	return snippet
}

func endedOnLoginPage(finalURL *url.URL) bool {
	if finalURL == nil {
		return false
	}
	return strings.Contains(strings.ToLower(finalURL.Path), "/login")
}

func storeCookieValue(jar http.CookieJar, baseURL *url.URL, name string) string {
	if jar == nil || baseURL == nil || strings.TrimSpace(name) == "" {
		return ""
	}
	requestURL := resolveURL(baseURL, "/")
	for _, cookie := range jar.Cookies(requestURL) {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}
