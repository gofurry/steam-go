package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/addons/websession"
	"github.com/gofurry/steam-go/api/authenticationservice"
)

func main() {
	var (
		accountFlag  = flag.String("account", "", "Steam account name; falls back to STEAM_ACCOUNT_NAME")
		passwordFlag = flag.String("password", "", "Steam account password; falls back to STEAM_PASSWORD")
		guardFlag    = flag.String("guard-code", "", "optional Steam Guard code; falls back to STEAM_GUARD_CODE")
		guardType    = flag.String("guard-type", "device_code", "guard code type: email_code, device_code, device_confirmation, or email_confirmation")
		deviceName   = flag.String("device-name", "steam-go example", "device friendly name sent to Steam")
		websiteID    = flag.String("website-id", "Store", "website ID used for BeginAuthSessionViaCredentials")
		language     = flag.Uint64("language", 0, "optional language id passed to BeginAuthSessionViaCredentials")
		proxyRaw     = flag.String("proxy", "", "optional HTTP proxy URL for Steam requests, e.g. http://127.0.0.1:7897")
		timeout      = flag.Duration("timeout", 15*time.Second, "per-request timeout")
		pollTimeout  = flag.Duration("poll-timeout", 2*time.Minute, "maximum time to wait for Steam approval or tokens")
		skipValidate = flag.Bool("skip-validate", false, "skip Store / Community cookie validation after token exchange")
	)
	flag.Parse()

	accountName := firstNonEmpty(*accountFlag, os.Getenv("STEAM_ACCOUNT_NAME"))
	password := firstNonEmpty(*passwordFlag, os.Getenv("STEAM_PASSWORD"))
	guardCode := firstNonEmpty(*guardFlag, os.Getenv("STEAM_GUARD_CODE"))

	if strings.TrimSpace(accountName) == "" || password == "" {
		log.Fatal("missing credentials: provide -account/-password or set STEAM_ACCOUNT_NAME and STEAM_PASSWORD")
	}

	selector, err := steam.NewStaticProxySelector(*proxyRaw)
	if err != nil {
		log.Fatal(err)
	}
	httpClient, err := steam.NewHTTPClientWithProxySelector(selector, *timeout)
	if err != nil {
		log.Fatal(err)
	}

	sdk, err := steam.NewClient(
		steam.WithTimeout(*timeout),
		steam.WithProxySelector(selector),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer sdk.Close()

	sessionClient, err := websession.NewClient(
		sdk.API.AuthenticationService,
		websession.WithHTTPClient(httpClient),
		websession.WithTimeout(*timeout),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	challenge, err := sessionClient.StartWithCredentials(ctx, websession.StartWithCredentialsRequest{
		AccountName: accountName,
		Password:    password,
		DeviceName:  *deviceName,
		WebsiteID:   *websiteID,
		Language:    *language,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("steamid=%s client_id=%d poll_interval=%s\n", challenge.SteamID, challenge.ClientID, effectivePollInterval(challenge.PollInterval))
	fmt.Println("allowed_confirmations:")
	for _, confirmation := range challenge.AllowedConfirmations {
		fmt.Printf("- %s: %s\n", confirmationTypeName(confirmation.ConfirmationType), strings.TrimSpace(confirmation.AssociatedMessage))
	}

	requiresTypedCode := hasTypedCodeConfirmation(challenge.AllowedConfirmations)
	allowsApprovalFlow := hasApprovalConfirmation(challenge.AllowedConfirmations)
	if guardCode != "" {
		typ, err := parseGuardCodeType(*guardType)
		if err != nil {
			log.Fatal(err)
		}
		if err := sessionClient.SubmitSteamGuardCode(ctx, challenge, guardCode, typ); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("submitted guard code as %s\n", guardTypeName(typ))
	} else if requiresTypedCode && !allowsApprovalFlow {
		log.Fatal("this login challenge requires a guard code; rerun with -guard-code or set STEAM_GUARD_CODE")
	} else if allowsApprovalFlow {
		fmt.Println("waiting for approval in Steam Guard or email confirmation flow")
	}

	loginResult, err := pollUntilReady(ctx, sessionClient, challenge, *pollTimeout)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(
		"login_ready account=%s steamid=%s refresh_token_len=%d access_token_len=%d\n",
		loginResult.AccountName,
		loginResult.SteamID,
		len(loginResult.RefreshToken),
		len(loginResult.AccessToken),
	)

	cookies, err := sessionClient.RefreshTokenToWebCookies(ctx, loginResult.RefreshToken)
	if err != nil {
		log.Fatal(err)
	}
	if !*skipValidate {
		if err := sessionClient.ValidateWebCookies(ctx, cookies); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf(
		"web_cookies_ready domains=%s sessionid=%s\n",
		strings.Join(cookies.Domains, ","),
		maskTail(cookies.SessionID, 6),
	)
	fmt.Println("tokens were not printed; use addon APIs directly if you need to store them yourself")
}

func pollUntilReady(ctx context.Context, client *websession.Client, challenge *websession.LoginChallenge, maxWait time.Duration) (*websession.LoginResult, error) {
	deadline := time.Now().Add(maxWait)
	interval := effectivePollInterval(challenge.PollInterval)
	for {
		result, err := client.Poll(ctx, challenge)
		if err != nil {
			var resultErr *authenticationservice.EResultError
			if errors.As(err, &resultErr) {
				return nil, fmt.Errorf("poll auth session failed with %s (%d): %w", firstNonEmpty(resultErr.Name, "EResult"), resultErr.Code, err)
			}
			return nil, err
		}
		if result.RefreshToken != "" || result.AccessToken != "" {
			return result, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out after %s while waiting for Steam auth session", maxWait)
		}
		time.Sleep(interval)
	}
}

func parseGuardCodeType(raw string) (authenticationservice.GuardCodeType, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "email_code":
		return authenticationservice.GuardCodeTypeEmailCode, nil
	case "device_code", "":
		return authenticationservice.GuardCodeTypeDeviceCode, nil
	case "device_confirmation":
		return authenticationservice.GuardCodeTypeDeviceConfirmation, nil
	case "email_confirmation":
		return authenticationservice.GuardCodeTypeEmailConfirmation, nil
	default:
		return 0, fmt.Errorf("unsupported guard type %q", raw)
	}
}

func guardTypeName(value authenticationservice.GuardCodeType) string {
	switch value {
	case authenticationservice.GuardCodeTypeEmailCode:
		return "email_code"
	case authenticationservice.GuardCodeTypeDeviceCode:
		return "device_code"
	case authenticationservice.GuardCodeTypeDeviceConfirmation:
		return "device_confirmation"
	case authenticationservice.GuardCodeTypeEmailConfirmation:
		return "email_confirmation"
	default:
		return fmt.Sprintf("unknown(%d)", value)
	}
}

func confirmationTypeName(value uint32) string {
	switch authenticationservice.GuardCodeType(value) {
	case authenticationservice.GuardCodeTypeEmailCode:
		return "email_code"
	case authenticationservice.GuardCodeTypeDeviceCode:
		return "device_code"
	case authenticationservice.GuardCodeTypeDeviceConfirmation:
		return "device_confirmation"
	case authenticationservice.GuardCodeTypeEmailConfirmation:
		return "email_confirmation"
	default:
		return fmt.Sprintf("unknown(%d)", value)
	}
}

func hasTypedCodeConfirmation(confirmations []authenticationservice.AuthConfirmation) bool {
	for _, confirmation := range confirmations {
		switch authenticationservice.GuardCodeType(confirmation.ConfirmationType) {
		case authenticationservice.GuardCodeTypeEmailCode, authenticationservice.GuardCodeTypeDeviceCode:
			return true
		}
	}
	return false
}

func hasApprovalConfirmation(confirmations []authenticationservice.AuthConfirmation) bool {
	for _, confirmation := range confirmations {
		switch authenticationservice.GuardCodeType(confirmation.ConfirmationType) {
		case authenticationservice.GuardCodeTypeDeviceConfirmation, authenticationservice.GuardCodeTypeEmailConfirmation:
			return true
		}
	}
	return false
}

func effectivePollInterval(interval time.Duration) time.Duration {
	if interval <= 0 {
		return 3 * time.Second
	}
	return interval
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func maskTail(value string, tail int) string {
	if len(value) <= tail {
		return value
	}
	return strings.Repeat("*", len(value)-tail) + value[len(value)-tail:]
}
