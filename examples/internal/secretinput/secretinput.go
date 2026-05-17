package secretinput

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

type fdProvider interface {
	Fd() uintptr
}

// Resolver reads sensitive values from one environment variable first and
// falls back to one hidden terminal prompt when interactive stdin is available.
type Resolver struct {
	Input        fdProvider
	PromptOutput io.Writer
	Getenv       func(string) string
	IsTerminal   func(int) bool
	ReadPassword func(int) ([]byte, error)
}

func DefaultResolver() Resolver {
	return Resolver{
		Input:        os.Stdin,
		PromptOutput: os.Stderr,
		Getenv:       os.Getenv,
		IsTerminal:   term.IsTerminal,
		ReadPassword: term.ReadPassword,
	}
}

func (r Resolver) Env(key string) string {
	if r.Getenv == nil {
		return os.Getenv(key)
	}
	return r.Getenv(key)
}

func (r Resolver) ResolveSensitive(envKey, prompt string) (string, error) {
	if value := r.Env(envKey); strings.TrimSpace(value) != "" {
		return value, nil
	}
	if r.Input == nil || r.IsTerminal == nil || !r.IsTerminal(int(r.Input.Fd())) {
		return "", fmt.Errorf("%s is required; set the environment variable or run the example in an interactive terminal", envKey)
	}
	if r.ReadPassword == nil {
		return "", fmt.Errorf("interactive secret input is not configured")
	}
	if r.PromptOutput != nil && prompt != "" {
		_, _ = fmt.Fprint(r.PromptOutput, prompt)
	}
	valueBytes, err := r.ReadPassword(int(r.Input.Fd()))
	if r.PromptOutput != nil {
		_, _ = fmt.Fprintln(r.PromptOutput)
	}
	if err != nil {
		return "", fmt.Errorf("read %s from terminal: %w", envKey, err)
	}
	value := string(valueBytes)
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s must not be empty", envKey)
	}
	return value, nil
}
