package main

import (
	"context"
	"strings"
	"time"

	"github.com/gofurry/steam-go/api/authenticationservice"
	"github.com/gofurry/steam-go/examples/internal/secretinput"
)

func resolvePassword(resolver secretinput.Resolver) (string, error) {
	return resolver.ResolveSensitive("STEAM_PASSWORD", "Steam password (input hidden): ")
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
