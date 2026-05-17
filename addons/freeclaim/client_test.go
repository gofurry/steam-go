package freeclaim

import (
	"errors"
	"testing"
)

func TestNewClientValidatesStorefrontService(t *testing.T) {
	t.Parallel()

	_, err := NewClient(nil)
	if err == nil {
		t.Fatal("expected nil storefront service error")
	}
	var clientErr *Error
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected freeclaim error, got %T", err)
	}
	if clientErr.Code != ErrorCodeConfig {
		t.Fatalf("unexpected error code: %s", clientErr.Code)
	}
}
