package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofurry/steam-go/addons/websession"
	"github.com/gofurry/steam-go/api/authenticationservice"
	"github.com/gofurry/steam-go/examples/internal/secretinput"
)

type loginOptions struct {
	AccountName string
	DeviceName  string
	WebsiteID   string
	Language    uint64
	GuardType   string
	PollTimeout time.Duration
}

func resolveRefreshToken(resolver secretinput.Resolver) (string, error) {
	return resolver.ResolveSensitive("STEAM_REFRESH_TOKEN", "Steam refresh token (input hidden): ")
}

func loginToWebCookies(ctx context.Context, resolver secretinput.Resolver, client *websession.Client, opts loginOptions) (*websession.WebCookieResult, error) {
	if client == nil {
		return nil, fmt.Errorf("websession client must not be nil")
	}
	accountName := strings.TrimSpace(firstNonEmpty(opts.AccountName, resolver.Env("STEAM_ACCOUNT_NAME")))
	if accountName == "" {
		return nil, fmt.Errorf("missing account name: provide -account or set STEAM_ACCOUNT_NAME")
	}
	password, err := resolver.ResolveSensitive("STEAM_PASSWORD", "Steam password (input hidden): ")
	if err != nil {
		return nil, err
	}

	challenge, err := client.StartWithCredentials(ctx, websession.StartWithCredentialsRequest{
		AccountName: accountName,
		Password:    password,
		DeviceName:  opts.DeviceName,
		WebsiteID:   opts.WebsiteID,
		Language:    opts.Language,
	})
	if err != nil {
		return nil, err
	}

	guardCode, err := resolveGuardCode(resolver, challenge.AllowedConfirmations)
	if err != nil {
		return nil, err
	}
	if guardCode != "" {
		typ, err := parseGuardCodeType(opts.GuardType)
		if err != nil {
			return nil, err
		}
		if err := client.SubmitSteamGuardCode(ctx, challenge, guardCode, typ); err != nil {
			return nil, err
		}
	}

	loginResult, err := pollUntilReady(ctx, client, challenge, opts.PollTimeout)
	if err != nil {
		return nil, err
	}
	return client.RefreshTokenToWebCookies(ctx, loginResult.RefreshToken)
}

func resolveGuardCode(resolver secretinput.Resolver, confirmations []authenticationservice.AuthConfirmation) (string, error) {
	if code := resolver.Env("STEAM_GUARD_CODE"); strings.TrimSpace(code) != "" {
		return code, nil
	}
	if hasTypedCodeConfirmation(confirmations) && !hasApprovalConfirmation(confirmations) {
		return resolver.ResolveSensitive("STEAM_GUARD_CODE", "Steam Guard code (input hidden): ")
	}
	return "", nil
}

func pollUntilReady(ctx context.Context, client *websession.Client, challenge *websession.LoginChallenge, maxWait time.Duration) (*websession.LoginResult, error) {
	if maxWait <= 0 {
		maxWait = 2 * time.Minute
	}
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
		if err := waitForPollInterval(ctx, interval); err != nil {
			return nil, err
		}
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

func waitForPollInterval(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		return nil
	}
	timer := time.NewTimer(interval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
