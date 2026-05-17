package main

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/steam-go/api/authenticationservice"
	"github.com/gofurry/steam-go/examples/internal/secretinput"
)

type fakeFD struct {
	fd uintptr
}

func (f fakeFD) Fd() uintptr {
	return f.fd
}

func TestResolvePasswordPrefersEnv(t *testing.T) {
	t.Parallel()

	var promptCalls int
	password, err := resolvePassword(secretinput.Resolver{
		Input: fakeFD{fd: 0},
		Getenv: func(key string) string {
			if key == "STEAM_PASSWORD" {
				return "env-password"
			}
			return ""
		},
		IsTerminal: func(int) bool { return true },
		ReadPassword: func(int) ([]byte, error) {
			promptCalls++
			return []byte("prompt-password"), nil
		},
		PromptOutput: io.Discard,
	})
	if err != nil {
		t.Fatalf("resolvePassword returned error: %v", err)
	}
	if password != "env-password" {
		t.Fatalf("unexpected password: %q", password)
	}
	if promptCalls != 0 {
		t.Fatalf("expected no terminal prompt, got %d calls", promptCalls)
	}
}

func TestResolvePasswordRejectsMissingSecretOnNonInteractiveInput(t *testing.T) {
	t.Parallel()

	_, err := resolvePassword(secretinput.Resolver{
		Input:        fakeFD{fd: 0},
		Getenv:       func(string) string { return "" },
		IsTerminal:   func(int) bool { return false },
		ReadPassword: func(int) ([]byte, error) { return []byte("unused"), nil },
		PromptOutput: io.Discard,
	})
	if err == nil {
		t.Fatal("expected missing password error")
	}
	if !strings.Contains(err.Error(), "STEAM_PASSWORD") {
		t.Fatalf("expected STEAM_PASSWORD in error, got %q", err)
	}
}

func TestResolveGuardCodeOnlyPromptsWhenTypedCodeIsRequired(t *testing.T) {
	t.Parallel()

	typedOnly := []authenticationservice.AuthConfirmation{{
		ConfirmationType: uint32(authenticationservice.GuardCodeTypeDeviceCode),
	}}
	typedWithApproval := []authenticationservice.AuthConfirmation{
		{ConfirmationType: uint32(authenticationservice.GuardCodeTypeDeviceCode)},
		{ConfirmationType: uint32(authenticationservice.GuardCodeTypeDeviceConfirmation)},
	}

	t.Run("prompts for typed-only challenge", func(t *testing.T) {
		var promptCalls int
		code, err := resolveGuardCode(secretinput.Resolver{
			Input:        fakeFD{fd: 0},
			Getenv:       func(string) string { return "" },
			IsTerminal:   func(int) bool { return true },
			PromptOutput: io.Discard,
			ReadPassword: func(int) ([]byte, error) {
				promptCalls++
				return []byte("123456"), nil
			},
		}, typedOnly)
		if err != nil {
			t.Fatalf("resolveGuardCode returned error: %v", err)
		}
		if code != "123456" {
			t.Fatalf("unexpected guard code: %q", code)
		}
		if promptCalls != 1 {
			t.Fatalf("expected one prompt, got %d", promptCalls)
		}
	})

	t.Run("does not prompt when approval flow is allowed", func(t *testing.T) {
		var promptCalls int
		code, err := resolveGuardCode(secretinput.Resolver{
			Input:        fakeFD{fd: 0},
			Getenv:       func(string) string { return "" },
			IsTerminal:   func(int) bool { return true },
			PromptOutput: io.Discard,
			ReadPassword: func(int) ([]byte, error) {
				promptCalls++
				return []byte("should-not-be-used"), nil
			},
		}, typedWithApproval)
		if err != nil {
			t.Fatalf("resolveGuardCode returned error: %v", err)
		}
		if code != "" {
			t.Fatalf("expected empty guard code, got %q", code)
		}
		if promptCalls != 0 {
			t.Fatalf("expected no prompt, got %d", promptCalls)
		}
	})
}

func TestWaitForPollIntervalHonorsContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	err := waitForPollInterval(ctx, 5*time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("waitForPollInterval returned too slowly: %s", elapsed)
	}
}
