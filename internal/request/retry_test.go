package request

import (
	"net/http"
	"testing"
	"time"
)

func TestRetryAfterDelayParsesSeconds(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		Header: http.Header{"Retry-After": []string{"3"}},
	}

	delay, ok := retryAfterDelay(resp, time.Unix(0, 0))
	if !ok {
		t.Fatal("expected Retry-After delay to be parsed")
	}
	if delay != 3*time.Second {
		t.Fatalf("unexpected delay: %s", delay)
	}
}

func TestRetryAfterDelayParsesHTTPDate(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	resp := &http.Response{
		Header: http.Header{
			"Retry-After": []string{now.Add(2 * time.Second).Format(http.TimeFormat)},
		},
	}

	delay, ok := retryAfterDelay(resp, now)
	if !ok {
		t.Fatal("expected Retry-After date to be parsed")
	}
	if delay != 2*time.Second {
		t.Fatalf("unexpected delay: %s", delay)
	}
}

func TestRetryAfterDelayReturnsZeroForPastDate(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	resp := &http.Response{
		Header: http.Header{
			"Retry-After": []string{now.Add(-1 * time.Second).Format(http.TimeFormat)},
		},
	}

	delay, ok := retryAfterDelay(resp, now)
	if !ok {
		t.Fatal("expected past Retry-After date to be accepted")
	}
	if delay != 0 {
		t.Fatalf("expected zero delay, got %s", delay)
	}
}

func TestRetryAfterDelayRejectsInvalidHeader(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		Header: http.Header{"Retry-After": []string{"not-a-date"}},
	}

	if delay, ok := retryAfterDelay(resp, time.Unix(0, 0)); ok {
		t.Fatalf("expected invalid header to be ignored, got %s", delay)
	}
}

func TestRetryDelayAddsBoundedJitterWithoutRetryAfter(t *testing.T) {
	t.Parallel()

	delay := retryDelay(0, nil, time.Unix(0, 0))
	if delay < 100*time.Millisecond || delay > 150*time.Millisecond {
		t.Fatalf("unexpected jittered delay: %s", delay)
	}
}
