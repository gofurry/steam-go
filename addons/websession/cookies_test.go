package websession

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRefreshTokenToWebCookiesAndValidateSessions(t *testing.T) {
	t.Parallel()

	steamID := "76561198000000001"
	refreshToken := testJWTWithSubject(t, steamID)

	var sawStoreTransfer bool
	var sawCommunityTransfer bool
	var sawStoreValidation bool
	var sawCommunityValidation bool

	storeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login/settoken":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm returned error: %v", err)
			}
			if got := r.Form.Get("steamID"); got != steamID {
				t.Fatalf("unexpected store steamID: %q", got)
			}
			http.SetCookie(w, &http.Cookie{Name: "sessionid", Value: "store-session", Path: "/"})
			http.SetCookie(w, &http.Cookie{Name: "steamLoginSecure", Value: "store-secure", Path: "/"})
			sawStoreTransfer = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		case "/account/":
			if r.URL.Query().Get("l") != "english" {
				t.Fatalf("unexpected store language query: %q", r.URL.RawQuery)
			}
			if _, err := r.Cookie("sessionid"); err != nil {
				t.Fatalf("expected store session cookie: %v", err)
			}
			sawStoreValidation = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("account overview"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer storeServer.Close()

	communityServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login/transfer":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm returned error: %v", err)
			}
			if got := r.Form.Get("steamID"); got != steamID {
				t.Fatalf("unexpected community steamID: %q", got)
			}
			http.SetCookie(w, &http.Cookie{Name: "steamLoginSecure", Value: "community-secure", Path: "/"})
			sawCommunityTransfer = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		case "/profiles/" + steamID + "/":
			if r.URL.Query().Get("xml") != "1" {
				t.Fatalf("unexpected community query: %q", r.URL.RawQuery)
			}
			if _, err := r.Cookie("steamLoginSecure"); err != nil {
				t.Fatalf("expected community login cookie: %v", err)
			}
			sawCommunityValidation = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<profile><steamID64>" + steamID + "</steamID64></profile>"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer communityServer.Close()

	loginServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jwt/finalizelogin" {
			http.NotFound(w, r)
			return
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm returned error: %v", err)
		}
		if got := r.Form.Get("nonce"); got != refreshToken {
			t.Fatalf("unexpected nonce: %q", got)
		}
		if got := r.Form.Get("redir"); got != communityServer.URL+"/login/home/?goto=" {
			t.Fatalf("unexpected redir: %q", got)
		}
		if r.Form.Get("sessionid") == "" {
			t.Fatal("expected sessionid form field")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":0,"transfer_info":[{"url":"` + storeServer.URL + `/login/settoken","params":{"nonce":"` + refreshToken + `","auth":"demo"}},{"url":"` + communityServer.URL + `/login/transfer","params":{"nonce":"` + refreshToken + `"}}]}`))
	}))
	defer loginServer.Close()

	client := newTestClient(
		t,
		newTestAuthService(t, &sequenceAuthTransport{}),
		WithHTTPClient(&http.Client{}),
		WithLoginBaseURL(loginServer.URL),
		WithStoreBaseURL(storeServer.URL),
		WithCommunityBaseURL(communityServer.URL),
	)

	result, err := client.RefreshTokenToWebCookies(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("RefreshTokenToWebCookies returned error: %v", err)
	}
	if result.Jar == nil {
		t.Fatal("expected cookie jar")
	}
	if result.SteamID != steamID {
		t.Fatalf("unexpected steam id: %s", result.SteamID)
	}
	if strings.TrimSpace(result.SessionID) == "" {
		t.Fatal("expected session id")
	}

	if err := client.ValidateWebCookies(context.Background(), result); err != nil {
		t.Fatalf("ValidateWebCookies returned error: %v", err)
	}
	if !sawStoreTransfer || !sawCommunityTransfer || !sawStoreValidation || !sawCommunityValidation {
		t.Fatalf("unexpected transfer/validation state: storeTransfer=%v communityTransfer=%v storeValidation=%v communityValidation=%v", sawStoreTransfer, sawCommunityTransfer, sawStoreValidation, sawCommunityValidation)
	}
}

func TestRefreshTokenToWebCookiesRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, newTestAuthService(t, &sequenceAuthTransport{}))

	_, err := client.RefreshTokenToWebCookies(context.Background(), "not-a-jwt")
	if err == nil {
		t.Fatal("expected invalid token error")
	}
	var clientErr *Error
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected websession error, got %T", err)
	}
	if clientErr.Code != ErrorCodeIdentity {
		t.Fatalf("unexpected error code: %s", clientErr.Code)
	}
}

func TestRefreshTokenToWebCookiesHonorsBodyLimit(t *testing.T) {
	t.Parallel()

	loginServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":0,"transfer_info":[],"padding":"abcdefghijklmnopqrstuvwxyz"}`))
	}))
	defer loginServer.Close()

	client := newTestClient(
		t,
		newTestAuthService(t, &sequenceAuthTransport{}),
		WithLoginBaseURL(loginServer.URL),
		WithMaxResponseBodyBytes(8),
	)

	_, err := client.RefreshTokenToWebCookies(context.Background(), testJWTWithSubject(t, "76561198000000001"))
	if err == nil {
		t.Fatal("expected body limit error")
	}
	var clientErr *Error
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected websession error, got %T", err)
	}
	if clientErr.Code != ErrorCodeTransport {
		t.Fatalf("unexpected error code: %s", clientErr.Code)
	}
}

func TestValidateStoreSessionDetectsLoginRedirect(t *testing.T) {
	t.Parallel()

	storeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/account/":
			http.Redirect(w, r, "/login/", http.StatusFound)
		case "/login/":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("login"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer storeServer.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New returned error: %v", err)
	}
	client := newTestClient(
		t,
		newTestAuthService(t, &sequenceAuthTransport{}),
		WithStoreBaseURL(storeServer.URL),
	)

	err = client.ValidateStoreSession(context.Background(), jar)
	if err == nil {
		t.Fatal("expected login redirect error")
	}
	var clientErr *Error
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected websession error, got %T", err)
	}
	if clientErr.Code != ErrorCodeVerify {
		t.Fatalf("unexpected error code: %s", clientErr.Code)
	}
}

func TestValidateCommunitySessionHonorsTimeout(t *testing.T) {
	t.Parallel()

	communityServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<profile><steamID64>76561198000000001</steamID64></profile>"))
	}))
	defer communityServer.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New returned error: %v", err)
	}
	client := newTestClient(
		t,
		newTestAuthService(t, &sequenceAuthTransport{}),
		WithCommunityBaseURL(communityServer.URL),
		WithTimeout(10*time.Millisecond),
	)

	err = client.ValidateCommunitySession(context.Background(), jar, "76561198000000001")
	if err == nil {
		t.Fatal("expected timeout error")
	}
	var clientErr *Error
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected websession error, got %T", err)
	}
	if clientErr.Code != ErrorCodeTransport {
		t.Fatalf("unexpected error code: %s", clientErr.Code)
	}
}
