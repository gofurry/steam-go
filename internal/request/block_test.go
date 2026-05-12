package request

import (
	"net/http"
	"strings"
	"testing"

	"github.com/GoFurry/steam-go/internal/errors"
	"github.com/GoFurry/steam-go/internal/traffic"
)

func TestNewBlockRuntimeOnlySupportsPublicStorePage(t *testing.T) {
	t.Parallel()

	if runtime := NewBlockRuntime(traffic.ClassOfficialAPI, BlockConfig{}); runtime != nil {
		t.Fatal("expected nil block runtime for official api")
	}
	if runtime := NewBlockRuntime(traffic.ClassPublicStorePage, BlockConfig{}); runtime == nil {
		t.Fatal("expected block runtime for public store page")
	}
}

func TestBlockRuntimeDetectsForbidden(t *testing.T) {
	t.Parallel()

	runtime := NewBlockRuntime(traffic.ClassPublicStorePage, BlockConfig{})
	resp := &http.Response{StatusCode: http.StatusForbidden, Header: make(http.Header)}

	result := runtime.detect(nil, resp, []byte("access denied"))
	if result == nil {
		t.Fatal("expected block result")
	}
	if result.ErrorKind != errors.KindHTTPStatus {
		t.Fatalf("unexpected error kind: %s", result.ErrorKind)
	}
	if !result.Retryable {
		t.Fatal("expected forbidden block to be retryable")
	}
	if !strings.Contains(result.Message, "forbidden") {
		t.Fatalf("unexpected message: %q", result.Message)
	}
}

func TestBlockRuntimeDetectsHTMLChallengeOnSuccessfulStatus(t *testing.T) {
	t.Parallel()

	runtime := NewBlockRuntime(traffic.ClassPublicStorePage, BlockConfig{})
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
	}

	result := runtime.detect(nil, resp, []byte("<html><body>Verify you are human with g-recaptcha</body></html>"))
	if result == nil {
		t.Fatal("expected block result")
	}
	if result.ErrorKind != errors.KindAPIResponse {
		t.Fatalf("unexpected error kind: %s", result.ErrorKind)
	}
	if result.Retryable {
		t.Fatal("did not expect 200 challenge to be retryable")
	}
	if !strings.Contains(result.Message, "challenge") {
		t.Fatalf("unexpected message: %q", result.Message)
	}
}

func TestBlockRuntimeIgnoresNonChallengeJSONResponses(t *testing.T) {
	t.Parallel()

	runtime := NewBlockRuntime(traffic.ClassPublicStorePage, BlockConfig{})
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}

	result := runtime.detect(nil, resp, []byte(`{"response":{"ok":true}}`))
	if result != nil {
		t.Fatalf("did not expect block result, got %#v", result)
	}
}

func TestBlockRuntimeUsesDefaultSniffBytes(t *testing.T) {
	t.Parallel()

	runtime := NewBlockRuntime(traffic.ClassPublicStorePage, BlockConfig{HTMLSniffBytes: 0})
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/html"}},
	}
	body := strings.Repeat(" ", defaultBlockHTMLSniffBytes-32) + "<html>cloudflare captcha</html>"

	result := runtime.detect(nil, resp, []byte(body))
	if result == nil {
		t.Fatal("expected block result with default sniff size")
	}
}
