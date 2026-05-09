package openid_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/GoFurry/steam-go/addons/openid"
)

func TestNewVerifierValidatesConfig(t *testing.T) {
	t.Parallel()

	_, err := openid.NewVerifier(openid.Config{})
	expectCode(t, err, openid.ErrorCodeConfig)

	_, err = openid.NewVerifier(
		openid.Config{
			Realm:    "https://example.com",
			ReturnTo: "https://example.com/callback",
		},
		openid.WithEndpoint("://bad"),
	)
	expectCode(t, err, openid.ErrorCodeConfig)

	_, err = openid.NewVerifier(
		openid.Config{
			Realm:    "https://example.com",
			ReturnTo: "https://example.com/callback",
		},
		openid.WithTimeout(0),
	)
	expectCode(t, err, openid.ErrorCodeConfig)

	_, err = openid.NewVerifier(
		openid.Config{
			Realm:    "https://example.com",
			ReturnTo: "https://example.com/callback",
		},
		openid.WithMaxResponseBodyBytes(0),
	)
	expectCode(t, err, openid.ErrorCodeConfig)
}

func TestLoginURLIncludesOpenIDParams(t *testing.T) {
	t.Parallel()

	verifier := newVerifier(t)

	raw, err := verifier.LoginURL("")
	if err != nil {
		t.Fatalf("LoginURL returned error: %v", err)
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	query := parsed.Query()
	if got := query.Get("openid.ns"); got != "http://specs.openid.net/auth/2.0" {
		t.Fatalf("unexpected openid.ns: %s", got)
	}
	if got := query.Get("openid.mode"); got != "checkid_setup" {
		t.Fatalf("unexpected openid.mode: %s", got)
	}
	if got := query.Get("openid.claimed_id"); got != "http://specs.openid.net/auth/2.0/identifier_select" {
		t.Fatalf("unexpected openid.claimed_id: %s", got)
	}
	if got := query.Get("openid.identity"); got != "http://specs.openid.net/auth/2.0/identifier_select" {
		t.Fatalf("unexpected openid.identity: %s", got)
	}
	if got := query.Get("openid.realm"); got != "https://example.com" {
		t.Fatalf("unexpected openid.realm: %s", got)
	}
	if got := query.Get("openid.return_to"); got != "https://example.com/auth/steam/callback" {
		t.Fatalf("unexpected openid.return_to: %s", got)
	}
}

func TestLoginURLIncludesStateInReturnTo(t *testing.T) {
	t.Parallel()

	verifier := newVerifier(t)

	raw, err := verifier.LoginURL("demo-state")
	if err != nil {
		t.Fatalf("LoginURL returned error: %v", err)
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	returnTo, err := url.Parse(parsed.Query().Get("openid.return_to"))
	if err != nil {
		t.Fatalf("Parse return_to returned error: %v", err)
	}
	if got := returnTo.Query().Get("state"); got != "demo-state" {
		t.Fatalf("unexpected state: %s", got)
	}
}

func TestVerifyValuesSuccess(t *testing.T) {
	t.Parallel()

	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioReadAllString(r)
		gotBody = body
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("Content-Type"); !strings.Contains(got, "application/x-www-form-urlencoded") {
			t.Fatalf("unexpected content type: %s", got)
		}
		_, _ = w.Write([]byte("ns:http://specs.openid.net/auth/2.0\nis_valid:true\n"))
	}))
	defer server.Close()

	verifier, err := openid.NewVerifier(
		openid.Config{
			Realm:    "https://example.com",
			ReturnTo: "https://example.com/auth/steam/callback",
		},
		openid.WithEndpoint(server.URL),
	)
	if err != nil {
		t.Fatalf("NewVerifier returned error: %v", err)
	}

	values := sampleValues("demo-state")
	identity, err := verifier.VerifyValues(context.Background(), values)
	if err != nil {
		t.Fatalf("VerifyValues returned error: %v", err)
	}
	if identity.SteamID != "76561198000000000" {
		t.Fatalf("unexpected steam id: %s", identity.SteamID)
	}
	if identity.ClaimedID != "https://steamcommunity.com/openid/id/76561198000000000" {
		t.Fatalf("unexpected claimed_id: %s", identity.ClaimedID)
	}
	if identity.State != "demo-state" {
		t.Fatalf("unexpected state: %s", identity.State)
	}
	if !strings.Contains(gotBody, "openid.mode=check_authentication") {
		t.Fatalf("expected check_authentication body, got: %s", gotBody)
	}
}

func TestVerifyValuesAcceptsHTTPClaimedID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("is_valid:true\n"))
	}))
	defer server.Close()

	verifier := newVerifierWithEndpoint(t, server.URL)
	values := sampleValues("")
	values.Set("openid.claimed_id", "http://steamcommunity.com/openid/id/76561198000000000")
	values.Set("openid.identity", "http://steamcommunity.com/openid/id/76561198000000000")

	identity, err := verifier.VerifyValues(context.Background(), values)
	if err != nil {
		t.Fatalf("VerifyValues returned error: %v", err)
	}
	if identity.SteamID != "76561198000000000" {
		t.Fatalf("unexpected steam id: %s", identity.SteamID)
	}
	if identity.ClaimedID != "http://steamcommunity.com/openid/id/76561198000000000" {
		t.Fatalf("unexpected claimed_id: %s", identity.ClaimedID)
	}
}

func TestVerifyRequestDelegatesToValues(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("is_valid:true\n"))
	}))
	defer server.Close()

	verifier, err := openid.NewVerifier(
		openid.Config{
			Realm:    "https://example.com",
			ReturnTo: "https://example.com/auth/steam/callback",
		},
		openid.WithEndpoint(server.URL),
	)
	if err != nil {
		t.Fatalf("NewVerifier returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/callback?"+sampleValues("demo-state").Encode(), nil)
	identity, err := verifier.VerifyRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("VerifyRequest returned error: %v", err)
	}
	if identity.State != "demo-state" {
		t.Fatalf("unexpected state: %s", identity.State)
	}
}

func TestVerifyValuesHandlesProviderInvalid(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("is_valid:false\n"))
	}))
	defer server.Close()

	verifier := newVerifierWithEndpoint(t, server.URL)
	_, err := verifier.VerifyValues(context.Background(), sampleValues(""))
	expectCode(t, err, openid.ErrorCodeVerify)
}

func TestVerifyValuesHandlesHTTPStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer server.Close()

	verifier := newVerifierWithEndpoint(t, server.URL)
	_, err := verifier.VerifyValues(context.Background(), sampleValues(""))
	expectCode(t, err, openid.ErrorCodeHTTPStatus)
}

func TestVerifyValuesHandlesTransportError(t *testing.T) {
	t.Parallel()

	verifier := newVerifierWithEndpoint(t, "http://127.0.0.1:1")
	_, err := verifier.VerifyValues(context.Background(), sampleValues(""))
	expectCode(t, err, openid.ErrorCodeTransport)
}

func TestVerifyValuesHandlesTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte("is_valid:true\n"))
	}))
	defer server.Close()

	verifier, err := openid.NewVerifier(
		openid.Config{
			Realm:    "https://example.com",
			ReturnTo: "https://example.com/auth/steam/callback",
		},
		openid.WithEndpoint(server.URL),
		openid.WithTimeout(20*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("NewVerifier returned error: %v", err)
	}

	_, err = verifier.VerifyValues(context.Background(), sampleValues(""))
	expectCode(t, err, openid.ErrorCodeTransport)
}

func TestVerifyValuesRejectsOversizedResponseBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("a", 128)))
	}))
	defer server.Close()

	verifier, err := openid.NewVerifier(
		openid.Config{
			Realm:    "https://example.com",
			ReturnTo: "https://example.com/auth/steam/callback",
		},
		openid.WithEndpoint(server.URL),
		openid.WithMaxResponseBodyBytes(32),
	)
	if err != nil {
		t.Fatalf("NewVerifier returned error: %v", err)
	}

	_, err = verifier.VerifyValues(context.Background(), sampleValues(""))
	expectCode(t, err, openid.ErrorCodeTransport)
}

func TestVerifyValuesRejectsBadMode(t *testing.T) {
	t.Parallel()

	verifier := newVerifier(t)
	values := sampleValues("")
	values.Set("openid.mode", "cancel")

	_, err := verifier.VerifyValues(context.Background(), values)
	expectCode(t, err, openid.ErrorCodeVerify)
}

func TestVerifyValuesRejectsBadClaimedID(t *testing.T) {
	t.Parallel()

	verifier := newVerifier(t)
	values := sampleValues("")
	values.Set("openid.claimed_id", "https://example.com/not-steam")

	_, err := verifier.VerifyValues(context.Background(), values)
	expectCode(t, err, openid.ErrorCodeIdentity)
}

func TestVerifyValuesRejectsReturnToMismatch(t *testing.T) {
	t.Parallel()

	verifier := newVerifier(t)
	values := sampleValues("")
	values.Set("openid.return_to", "https://example.com/other")

	_, err := verifier.VerifyValues(context.Background(), values)
	expectCode(t, err, openid.ErrorCodeVerify)
}

func TestVerifyValuesSupportsEndpointOverride(t *testing.T) {
	t.Parallel()

	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		_, _ = w.Write([]byte("is_valid:true\n"))
	}))
	defer server.Close()

	verifier := newVerifierWithEndpoint(t, server.URL)
	if _, err := verifier.VerifyValues(context.Background(), sampleValues("")); err != nil {
		t.Fatalf("VerifyValues returned error: %v", err)
	}
	if !called {
		t.Fatal("expected endpoint override server to be called")
	}
}

func newVerifier(t *testing.T) *openid.Verifier {
	t.Helper()
	verifier, err := openid.NewVerifier(openid.Config{
		Realm:    "https://example.com",
		ReturnTo: "https://example.com/auth/steam/callback",
	})
	if err != nil {
		t.Fatalf("NewVerifier returned error: %v", err)
	}
	return verifier
}

func newVerifierWithEndpoint(t *testing.T, endpoint string) *openid.Verifier {
	t.Helper()
	verifier, err := openid.NewVerifier(
		openid.Config{
			Realm:    "https://example.com",
			ReturnTo: "https://example.com/auth/steam/callback",
		},
		openid.WithEndpoint(endpoint),
	)
	if err != nil {
		t.Fatalf("NewVerifier returned error: %v", err)
	}
	return verifier
}

func sampleValues(state string) url.Values {
	returnTo := "https://example.com/auth/steam/callback"
	if state != "" {
		returnTo += "?state=" + url.QueryEscape(state)
	}
	return url.Values{
		"openid.ns":             []string{"http://specs.openid.net/auth/2.0"},
		"openid.mode":           []string{"id_res"},
		"openid.op_endpoint":    []string{"https://steamcommunity.com/openid/login"},
		"openid.claimed_id":     []string{"https://steamcommunity.com/openid/id/76561198000000000"},
		"openid.identity":       []string{"https://steamcommunity.com/openid/id/76561198000000000"},
		"openid.return_to":      []string{returnTo},
		"openid.response_nonce": []string{"2026-04-16T00:00:00Zxyz"},
		"openid.assoc_handle":   []string{"1234567890"},
		"openid.signed":         []string{"signed,op_endpoint,claimed_id,identity,return_to,response_nonce,assoc_handle"},
		"openid.sig":            []string{"signature"},
	}
}

func expectCode(t *testing.T, err error, code openid.ErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %s, got nil", code)
	}
	var openIDErr *openid.Error
	if !errors.As(err, &openIDErr) {
		t.Fatalf("expected *openid.Error, got %T", err)
	}
	if openIDErr.Code != code {
		t.Fatalf("expected error code %s, got %s", code, openIDErr.Code)
	}
}

func ioReadAllString(r *http.Request) (string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
