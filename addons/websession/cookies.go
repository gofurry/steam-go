package websession

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type WebCookieResult struct {
	Jar       http.CookieJar
	SessionID string
	SteamID   string
	Domains   []string
}

type finalizeLoginResponse struct {
	Error        int            `json:"error"`
	Message      string         `json:"message"`
	TransferInfo []transferInfo `json:"transfer_info"`
}

type transferInfo struct {
	URL    string                     `json:"url"`
	Params map[string]json.RawMessage `json:"params"`
}

func (c *Client) RefreshTokenToWebCookies(ctx context.Context, refreshToken string) (*WebCookieResult, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, &Error{Code: ErrorCodeRequestBuild, Op: "refresh_token_to_web_cookies", Message: "refresh token must not be empty"}
	}
	steamID, err := steamIDFromJWT(refreshToken)
	if err != nil {
		return nil, &Error{Code: ErrorCodeIdentity, Op: "refresh_token_to_web_cookies", Message: "decode steam id from refresh token failed", Err: err}
	}
	sessionID, err := randomSessionID()
	if err != nil {
		return nil, &Error{Code: ErrorCodeRequestBuild, Op: "refresh_token_to_web_cookies", Message: "generate session id failed", Err: err}
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, &Error{Code: ErrorCodeConfig, Op: "refresh_token_to_web_cookies", Message: "create cookie jar failed", Err: err}
	}
	httpClient := cloneHTTPClient(c.httpClient)
	httpClient.Jar = jar

	finalizeURL := resolveURL(c.loginBaseURL, "/jwt/finalizelogin")
	redirURL := resolveURL(c.communityBaseURL, "/login/home/")
	redirQuery := redirURL.Query()
	redirQuery.Set("goto", "")
	redirURL.RawQuery = redirQuery.Encode()
	form := url.Values{}
	form.Set("nonce", refreshToken)
	form.Set("sessionid", sessionID)
	form.Set("redir", redirURL.String())
	body, err := c.postForm(ctx, httpClient, finalizeURL.String(), form)
	if err != nil {
		return nil, err
	}

	var finalized finalizeLoginResponse
	if err := json.Unmarshal(body, &finalized); err != nil {
		return nil, &Error{Code: ErrorCodeDecode, Op: "refresh_token_to_web_cookies", Message: "decode finalizelogin response failed", Err: err}
	}
	if finalized.Error != 0 {
		return nil, &Error{Code: ErrorCodeVerify, Op: "refresh_token_to_web_cookies", Message: fmt.Sprintf("finalizelogin returned error %d: %s", finalized.Error, finalized.Message)}
	}

	domains := map[string]struct{}{}
	addDomain(domains, c.loginBaseURL)
	for _, transfer := range finalized.TransferInfo {
		transferURL, err := parseAbsoluteURL(transfer.URL)
		if err != nil {
			return nil, &Error{Code: ErrorCodeDecode, Op: "refresh_token_to_web_cookies", Message: "invalid transfer url", Err: err}
		}
		transferForm := url.Values{}
		transferForm.Set("steamID", steamID)
		for key, raw := range transfer.Params {
			value, err := rawJSONToString(raw)
			if err != nil {
				return nil, &Error{Code: ErrorCodeDecode, Op: "refresh_token_to_web_cookies", Message: "invalid transfer param", Err: err}
			}
			transferForm.Set(key, value)
		}
		if _, err := c.postForm(ctx, httpClient, transferURL.String(), transferForm); err != nil {
			return nil, err
		}
		addDomain(domains, transferURL)
	}

	return &WebCookieResult{
		Jar:       jar,
		SessionID: sessionID,
		SteamID:   steamID,
		Domains:   sortedDomains(domains),
	}, nil
}

func (c *Client) postForm(ctx context.Context, client *http.Client, rawURL string, form url.Values) ([]byte, error) {
	reqCtx, cancel := c.withTimeout(ctx)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, rawURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, &Error{Code: ErrorCodeRequestBuild, Op: "post_form", Message: "build request failed", Err: err}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, &Error{Code: ErrorCodeTransport, Op: "post_form", Message: "request failed", Err: err}
	}
	defer resp.Body.Close()

	body, readErr := readBodyLimited(resp.Body, c.maxResponseBodyBytes)
	if readErr != nil {
		return nil, &Error{Code: ErrorCodeTransport, Op: "post_form", Message: "read response failed", Err: readErr}
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &Error{Code: ErrorCodeHTTPStatus, Op: "post_form", Message: fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))}
	}
	return body, nil
}

func cloneHTTPClient(base *http.Client) *http.Client {
	if base == nil {
		return &http.Client{}
	}
	cloned := *base
	return &cloned
}

func randomSessionID() (string, error) {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}

func rawJSONToString(raw json.RawMessage) (string, error) {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}
	var number json.Number
	if err := json.Unmarshal(raw, &number); err == nil {
		return number.String(), nil
	}
	var boolean bool
	if err := json.Unmarshal(raw, &boolean); err == nil {
		if boolean {
			return "true", nil
		}
		return "false", nil
	}
	return "", fmt.Errorf("unsupported json value %s", string(raw))
}
