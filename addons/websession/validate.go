package websession

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

func (c *Client) ValidateWebCookies(ctx context.Context, result *WebCookieResult) error {
	if result == nil {
		return &Error{Code: ErrorCodeRequestBuild, Op: "validate_web_cookies", Message: "web cookie result must not be nil"}
	}
	if err := c.ValidateCommunitySession(ctx, result.Jar, result.SteamID); err != nil {
		return err
	}
	return c.ValidateStoreSession(ctx, result.Jar)
}

func (c *Client) ValidateCommunitySession(ctx context.Context, jar http.CookieJar, steamID string) error {
	if jar == nil {
		return &Error{Code: ErrorCodeRequestBuild, Op: "validate_community_session", Message: "cookie jar must not be nil"}
	}
	if _, err := strconv.ParseUint(strings.TrimSpace(steamID), 10, 64); err != nil {
		return &Error{Code: ErrorCodeRequestBuild, Op: "validate_community_session", Message: "steam id must be a uint64 string", Err: err}
	}
	parsed := resolveURL(c.communityBaseURL, "/profiles/"+steamID+"/")
	query := parsed.Query()
	query.Set("xml", "1")
	parsed.RawQuery = query.Encode()
	body, err := c.getWithJar(ctx, jar, parsed.String(), "validate_community_session")
	if err != nil {
		return err
	}
	if !strings.Contains(string(body), steamID) {
		return &Error{Code: ErrorCodeVerify, Op: "validate_community_session", Message: "community profile did not match steam id"}
	}
	return nil
}

func (c *Client) ValidateStoreSession(ctx context.Context, jar http.CookieJar) error {
	if jar == nil {
		return &Error{Code: ErrorCodeRequestBuild, Op: "validate_store_session", Message: "cookie jar must not be nil"}
	}
	parsed := resolveURL(c.storeBaseURL, "/account/")
	query := parsed.Query()
	query.Set("l", "english")
	parsed.RawQuery = query.Encode()
	body, err := c.getWithJar(ctx, jar, parsed.String(), "validate_store_session")
	if err != nil {
		return err
	}
	if strings.Contains(strings.ToLower(string(body)), "login") {
		return &Error{Code: ErrorCodeVerify, Op: "validate_store_session", Message: "store account page appears to require login"}
	}
	return nil
}

func (c *Client) getWithJar(ctx context.Context, jar http.CookieJar, rawURL, op string) ([]byte, error) {
	reqCtx, cancel := c.withTimeout(ctx)
	defer cancel()

	client := cloneHTTPClient(c.httpClient)
	client.Jar = jar
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, &Error{Code: ErrorCodeRequestBuild, Op: op, Message: "build request failed", Err: err}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &Error{Code: ErrorCodeTransport, Op: op, Message: "request failed", Err: err}
	}
	defer resp.Body.Close()
	body, readErr := readBodyLimited(resp.Body, c.maxResponseBodyBytes)
	if readErr != nil {
		return nil, &Error{Code: ErrorCodeTransport, Op: op, Message: "read response failed", Err: readErr}
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &Error{Code: ErrorCodeHTTPStatus, Op: op, Message: fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))}
	}
	if resp.Request != nil && strings.Contains(strings.ToLower(resp.Request.URL.Path), "/login") {
		return nil, &Error{Code: ErrorCodeVerify, Op: op, Message: "request ended on a login page"}
	}
	return body, nil
}

func steamIDFromJWT(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("jwt must contain a payload")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	var claims struct {
		Subject string `json:"sub"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", err
	}
	if _, err := strconv.ParseUint(claims.Subject, 10, 64); err != nil {
		return "", err
	}
	return claims.Subject, nil
}

func addDomain(domains map[string]struct{}, u *url.URL) {
	if u == nil || u.Hostname() == "" {
		return
	}
	domains[u.Hostname()] = struct{}{}
}

func sortedDomains(domains map[string]struct{}) []string {
	out := make([]string, 0, len(domains))
	for domain := range domains {
		out = append(out, domain)
	}
	sort.Strings(out)
	return out
}
