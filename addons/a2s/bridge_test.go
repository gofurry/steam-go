package a2s_test

import (
	"errors"
	"testing"
	"time"

	a2saddon "github.com/GoFurry/steam-go/addons/a2s"
)

func TestNewClientBridgesToA2SGo(t *testing.T) {
	t.Parallel()

	client, err := a2saddon.NewClient(
		"127.0.0.1:27015",
		a2saddon.WithTimeout(2*time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	if client == nil {
		t.Fatal("expected client")
	}
}

func TestNewClientValidationStillComesFromA2SGo(t *testing.T) {
	t.Parallel()

	_, err := a2saddon.NewClient("")
	if err == nil {
		t.Fatal("expected error")
	}

	var a2sErr *a2saddon.Error
	if !errors.As(err, &a2sErr) {
		t.Fatalf("expected *a2s.Error, got %T", err)
	}
	if a2sErr.Code != a2saddon.ErrorCodeAddress {
		t.Fatalf("unexpected error code: %s", a2sErr.Code)
	}
}
