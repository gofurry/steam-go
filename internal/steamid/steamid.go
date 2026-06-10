package steamid

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidateSteamID64 validates and normalizes one SteamID64 string.
func ValidateSteamID64(value string) (string, error) {
	return ValidateNumericID("steam id", value)
}

// ValidateNumericID validates and normalizes one unsigned decimal identifier.
func ValidateNumericID(name, value string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "id"
	}

	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s must not be empty", name)
	}
	for _, r := range trimmed {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("%s must contain digits only", name)
		}
	}
	if _, err := strconv.ParseUint(trimmed, 10, 64); err != nil {
		return "", fmt.Errorf("%s must be a uint64 string: %w", name, err)
	}
	return trimmed, nil
}
