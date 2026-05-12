package steam_test

import (
	"errors"
	"testing"

	steam "github.com/GoFurry/steam-go"
)

func TestAPIErrorBodyPreviewReturnsEmptyForNonPositiveMax(t *testing.T) {
	t.Parallel()

	err := &steam.APIError{Body: []byte("hello")}
	if got := err.BodyPreview(0); got != "" {
		t.Fatalf("BodyPreview(0) = %q, want empty", got)
	}
	if got := err.BodyPreview(-1); got != "" {
		t.Fatalf("BodyPreview(-1) = %q, want empty", got)
	}
}

func TestAPIErrorBodyPreviewReturnsFullBodyWhenShortEnough(t *testing.T) {
	t.Parallel()

	err := &steam.APIError{Body: []byte("hello")}
	if got := err.BodyPreview(16); got != "hello" {
		t.Fatalf("BodyPreview(16) = %q, want %q", got, "hello")
	}
}

func TestAPIErrorBodyPreviewTruncatesLongBodies(t *testing.T) {
	t.Parallel()

	err := &steam.APIError{Body: []byte("hello world")}
	if got := err.BodyPreview(5); got != "hello" {
		t.Fatalf("BodyPreview(5) = %q, want %q", got, "hello")
	}
}

func TestAPIErrorBodyPreviewDoesNotBreakErrorsAs(t *testing.T) {
	t.Parallel()

	want := &steam.APIError{
		Kind:    steam.ErrorKindHTTPStatus,
		Message: "boom",
		Body:    []byte("payload"),
	}
	err := errors.Join(want)

	var apiErr *steam.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected errors.As to find APIError, got %T", err)
	}
	if got := apiErr.BodyPreview(4); got != "payl" {
		t.Fatalf("BodyPreview(4) = %q, want %q", got, "payl")
	}
}
