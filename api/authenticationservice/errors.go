package authenticationservice

import "fmt"

// EResultError exposes Steam EResult failures while preserving the SDK's common APIError wrapper.
type EResultError struct {
	Code    int
	Name    string
	Message string
}

func (e *EResultError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Name != "" && e.Message != "" {
		return fmt.Sprintf("steam eresult %d %s: %s", e.Code, e.Name, e.Message)
	}
	if e.Name != "" {
		return fmt.Sprintf("steam eresult %d %s", e.Code, e.Name)
	}
	if e.Message != "" {
		return fmt.Sprintf("steam eresult %d: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("steam eresult %d", e.Code)
}

func erResultName(code int) string {
	switch code {
	case 29:
		return "DuplicateRequest"
	case 63:
		return "AccountLogonDenied"
	case 65:
		return "InvalidLoginAuthCode"
	case 71:
		return "ExpiredLoginAuthCode"
	case 84:
		return "RateLimitExceeded"
	default:
		return ""
	}
}
