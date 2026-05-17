package main

import (
	"io"
	"strings"
	"testing"

	"github.com/gofurry/steam-go/examples/internal/secretinput"
)

type fakeFD struct {
	fd uintptr
}

func (f fakeFD) Fd() uintptr {
	return f.fd
}

func TestResolveRefreshTokenPrefersEnv(t *testing.T) {
	t.Parallel()

	var promptCalls int
	token, err := resolveRefreshToken(secretinput.Resolver{
		Input: fakeFD{fd: 0},
		Getenv: func(key string) string {
			if key == "STEAM_REFRESH_TOKEN" {
				return "env-refresh-token"
			}
			return ""
		},
		IsTerminal: func(int) bool { return true },
		ReadPassword: func(int) ([]byte, error) {
			promptCalls++
			return []byte("prompt-refresh-token"), nil
		},
		PromptOutput: io.Discard,
	})
	if err != nil {
		t.Fatalf("resolveRefreshToken returned error: %v", err)
	}
	if token != "env-refresh-token" {
		t.Fatalf("unexpected refresh token: %q", token)
	}
	if promptCalls != 0 {
		t.Fatalf("expected no terminal prompt, got %d calls", promptCalls)
	}
}

func TestResolveRefreshTokenRejectsMissingSecretOnNonInteractiveInput(t *testing.T) {
	t.Parallel()

	_, err := resolveRefreshToken(secretinput.Resolver{
		Input:        fakeFD{fd: 0},
		Getenv:       func(string) string { return "" },
		IsTerminal:   func(int) bool { return false },
		ReadPassword: func(int) ([]byte, error) { return []byte("unused"), nil },
		PromptOutput: io.Discard,
	})
	if err == nil {
		t.Fatal("expected missing refresh token error")
	}
	if !strings.Contains(err.Error(), "STEAM_REFRESH_TOKEN") {
		t.Fatalf("expected STEAM_REFRESH_TOKEN in error, got %q", err)
	}
}
