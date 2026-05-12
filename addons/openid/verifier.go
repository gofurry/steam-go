package openid

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	openIDNamespace  = "http://specs.openid.net/auth/2.0"
	identifierSelect = "http://specs.openid.net/auth/2.0/identifier_select"
	claimedIDHost    = "steamcommunity.com"
	claimedIDPrefix  = "/openid/id/"
)

type Config struct {
	Realm    string
	ReturnTo string
}

type Identity struct {
	SteamID   string
	ClaimedID string
	State     string
}

type Verifier struct {
	realm                string
	returnTo             *url.URL
	endpoint             *url.URL
	httpClient           *http.Client
	timeout              time.Duration
	stateParam           string
	maxResponseBodyBytes int64
}

func NewVerifier(cfg Config, opts ...Option) (*Verifier, error) {
	realmURL, err := parseAbsoluteURL(cfg.Realm)
	if err != nil {
		return nil, &Error{
			Code:    ErrorCodeConfig,
			Op:      "new_verifier",
			Message: "invalid realm",
			Err:     err,
		}
	}

	returnToURL, err := parseAbsoluteURL(cfg.ReturnTo)
	if err != nil {
		return nil, &Error{
			Code:    ErrorCodeConfig,
			Op:      "new_verifier",
			Message: "invalid return_to",
			Err:     err,
		}
	}

	options := defaultVerifierOptions()
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	if options.endpoint == nil {
		options.endpoint, err = parseAbsoluteURL(defaultEndpoint)
		if err != nil {
			return nil, &Error{
				Code:    ErrorCodeConfig,
				Op:      "new_verifier",
				Message: "invalid default endpoint",
				Err:     err,
			}
		}
	}

	return &Verifier{
		realm:                realmURL.String(),
		returnTo:             cloneURL(returnToURL),
		endpoint:             cloneURL(options.endpoint),
		httpClient:           options.httpClient,
		timeout:              options.timeout,
		stateParam:           options.stateParam,
		maxResponseBodyBytes: options.maxResponseBodyBytes,
	}, nil
}

func (v *Verifier) LoginURL(state string) (string, error) {
	if v == nil {
		return "", &Error{
			Code:    ErrorCodeRequestBuild,
			Op:      "login_url",
			Message: "verifier must not be nil",
		}
	}

	returnTo := v.returnToWithState(state)
	endpoint := cloneURL(v.endpoint)
	query := endpoint.Query()
	query.Set("openid.ns", openIDNamespace)
	query.Set("openid.mode", "checkid_setup")
	query.Set("openid.claimed_id", identifierSelect)
	query.Set("openid.identity", identifierSelect)
	query.Set("openid.realm", v.realm)
	query.Set("openid.return_to", returnTo.String())
	endpoint.RawQuery = query.Encode()
	return endpoint.String(), nil
}

func (v *Verifier) VerifyRequest(ctx context.Context, req *http.Request) (*Identity, error) {
	if req == nil || req.URL == nil {
		return nil, &Error{
			Code:    ErrorCodeRequestBuild,
			Op:      "verify_request",
			Message: "request must not be nil",
		}
	}

	return v.VerifyValues(ctx, req.URL.Query())
}

func (v *Verifier) VerifyValues(ctx context.Context, values url.Values) (*Identity, error) {
	if v == nil {
		return nil, &Error{
			Code:    ErrorCodeRequestBuild,
			Op:      "verify_values",
			Message: "verifier must not be nil",
		}
	}
	if values == nil {
		return nil, &Error{
			Code:    ErrorCodeRequestBuild,
			Op:      "verify_values",
			Message: "values must not be nil",
		}
	}
	if mode := values.Get("openid.mode"); mode != "id_res" {
		return nil, &Error{
			Code:    ErrorCodeVerify,
			Op:      "verify_values",
			Message: fmt.Sprintf("unexpected openid.mode %q", mode),
		}
	}

	claimedID := values.Get("openid.claimed_id")
	steamID, err := parseSteamID(claimedID)
	if err != nil {
		return nil, err
	}

	returnToRaw := values.Get("openid.return_to")
	state, err := v.verifyReturnTo(returnToRaw)
	if err != nil {
		return nil, err
	}

	checkValues := cloneValues(values)
	checkValues.Set("openid.mode", "check_authentication")

	checkCtx, cancel := withTimeout(ctx, v.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(checkCtx, http.MethodPost, v.endpoint.String(), strings.NewReader(checkValues.Encode()))
	if err != nil {
		return nil, &Error{
			Code:    ErrorCodeRequestBuild,
			Op:      "verify_values",
			Message: "build verification request failed",
			Err:     err,
		}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, &Error{
			Code:    ErrorCodeTransport,
			Op:      "verify_values",
			Message: "openid verification request failed",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	body, readErr := readBodyLimited(resp.Body, v.maxResponseBodyBytes)
	if readErr != nil {
		return nil, &Error{
			Code:    ErrorCodeTransport,
			Op:      "verify_values",
			Message: "read verification response failed",
			Err:     readErr,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &Error{
			Code:    ErrorCodeHTTPStatus,
			Op:      "verify_values",
			Message: fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body))),
		}
	}

	if !providerResponseValid(string(body)) {
		return nil, &Error{
			Code:    ErrorCodeVerify,
			Op:      "verify_values",
			Message: "provider rejected authentication response",
		}
	}

	return &Identity{
		SteamID:   steamID,
		ClaimedID: claimedID,
		State:     state,
	}, nil
}

func (v *Verifier) verifyReturnTo(raw string) (string, error) {
	returnToURL, err := parseAbsoluteURL(raw)
	if err != nil {
		return "", &Error{
			Code:    ErrorCodeVerify,
			Op:      "verify_values",
			Message: "invalid openid.return_to",
			Err:     err,
		}
	}

	state := returnToURL.Query().Get(v.stateParam)
	expected := v.returnToWithState(state)
	if expected.String() != returnToURL.String() {
		return "", &Error{
			Code:    ErrorCodeVerify,
			Op:      "verify_values",
			Message: "openid.return_to mismatch",
		}
	}

	return state, nil
}

func (v *Verifier) returnToWithState(state string) *url.URL {
	returnTo := cloneURL(v.returnTo)
	query := returnTo.Query()
	if state == "" {
		query.Del(v.stateParam)
	} else {
		query.Set(v.stateParam, state)
	}
	returnTo.RawQuery = query.Encode()
	return returnTo
}

func parseAbsoluteURL(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, fmt.Errorf("url must not be empty")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("url must be absolute")
	}
	return parsed, nil
}

func parseSteamID(claimedID string) (string, error) {
	parsed, err := url.Parse(claimedID)
	if err != nil {
		return "", &Error{
			Code:    ErrorCodeIdentity,
			Op:      "verify_values",
			Message: "invalid claimed_id",
		}
	}

	if parsed.Scheme != "https" || !strings.EqualFold(parsed.Host, claimedIDHost) {
		return "", &Error{
			Code:    ErrorCodeIdentity,
			Op:      "verify_values",
			Message: "invalid claimed_id",
		}
	}

	if !strings.HasPrefix(parsed.Path, claimedIDPrefix) {
		return "", &Error{
			Code:    ErrorCodeIdentity,
			Op:      "verify_values",
			Message: "invalid claimed_id",
		}
	}

	steamID := strings.TrimPrefix(parsed.Path, claimedIDPrefix)
	if steamID == "" || strings.Contains(steamID, "/") {
		return "", &Error{
			Code:    ErrorCodeIdentity,
			Op:      "verify_values",
			Message: "invalid claimed_id",
		}
	}

	if _, err := strconv.ParseUint(steamID, 10, 64); err != nil {
		return "", &Error{
			Code:    ErrorCodeIdentity,
			Op:      "verify_values",
			Message: "invalid steamid in claimed_id",
			Err:     err,
		}
	}

	return steamID, nil
}

func providerResponseValid(body string) bool {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.EqualFold(line, "is_valid:true") {
			return true
		}
	}
	return false
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	cloned := *u
	return &cloned
}

func cloneValues(values url.Values) url.Values {
	if values == nil {
		return nil
	}
	cloned := make(url.Values, len(values))
	for key, items := range values {
		copied := make([]string, len(items))
		copy(copied, items)
		cloned[key] = copied
	}
	return cloned
}

func withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}

func readBodyLimited(r io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return io.ReadAll(r)
	}

	reader := &io.LimitedReader{R: r, N: maxBytes + 1}
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("response body exceeds limit of %d bytes", maxBytes)
	}
	return body, nil
}
