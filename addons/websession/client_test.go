package websession

import (
	"errors"
	"testing"
)

func TestNewClientValidatesInputs(t *testing.T) {
	t.Parallel()

	_, err := NewClient(nil)
	if err == nil {
		t.Fatal("expected nil authentication service error")
	}
	var clientErr *Error
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected websession error, got %T", err)
	}
	if clientErr.Code != ErrorCodeConfig {
		t.Fatalf("unexpected error code: %s", clientErr.Code)
	}

	auth := newTestAuthService(t, &sequenceAuthTransport{})
	_, err = NewClient(auth, WithLoginBaseURL("://bad"))
	if err == nil {
		t.Fatal("expected invalid login base url error")
	}
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected websession error, got %T", err)
	}
	if clientErr.Code != ErrorCodeConfig {
		t.Fatalf("unexpected error code: %s", clientErr.Code)
	}
}
