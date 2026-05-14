package request

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/traffic"
)

const defaultBlockHTMLSniffBytes = 8 * 1024

var blockChallengeMarkers = []string{
	"cf-browser-verification",
	"g-recaptcha",
	"hcaptcha",
	"cf-chl-",
	"/cdn-cgi/challenge-platform/",
	"turnstile",
}

var blockChallengeWeakMarkers = []string{
	"captcha",
	"verify you are human",
	"access denied",
	"attention required",
}

type BlockRuntime interface {
	detect(req *http.Request, resp *http.Response, body []byte) *BlockResult
}

type BlockResult struct {
	ErrorKind sdkerrors.Kind
	Message   string
	Retryable bool
}

type BlockConfig struct {
	HTMLSniffBytes int
}

type blockRuntime struct {
	class          traffic.Class
	htmlSniffBytes int
}

func NewBlockRuntime(class traffic.Class, cfg BlockConfig) BlockRuntime {
	class = traffic.NormalizeClass(class)
	if !supportsBlockDetection(class) {
		return nil
	}

	htmlSniffBytes := cfg.HTMLSniffBytes
	if htmlSniffBytes <= 0 {
		htmlSniffBytes = defaultBlockHTMLSniffBytes
	}

	return blockRuntime{
		class:          class,
		htmlSniffBytes: htmlSniffBytes,
	}
}

func (r blockRuntime) detect(_ *http.Request, resp *http.Response, body []byte) *BlockResult {
	if resp == nil {
		return nil
	}

	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		return &BlockResult{
			ErrorKind: sdkerrors.KindHTTPStatus,
			Message:   fmt.Sprintf("%s block detected: rate_limited (status=%d)", r.surfaceLabel(), resp.StatusCode),
			Retryable: true,
		}
	case http.StatusForbidden:
		return &BlockResult{
			ErrorKind: sdkerrors.KindHTTPStatus,
			Message:   fmt.Sprintf("%s block detected: forbidden (status=%d)", r.surfaceLabel(), resp.StatusCode),
			Retryable: true,
		}
	}

	if !blockLikeHTMLResponse(resp, body, r.htmlSniffBytes) {
		return nil
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return &BlockResult{
			ErrorKind: sdkerrors.KindAPIResponse,
			Message:   fmt.Sprintf("%s block detected: challenge html response", r.surfaceLabel()),
			Retryable: false,
		}
	}
	if resp.StatusCode >= http.StatusInternalServerError {
		return &BlockResult{
			ErrorKind: sdkerrors.KindHTTPStatus,
			Message:   fmt.Sprintf("%s block detected: challenge html response (status=%d)", r.surfaceLabel(), resp.StatusCode),
			Retryable: true,
		}
	}

	return nil
}

func supportsBlockDetection(class traffic.Class) bool {
	switch class {
	case traffic.ClassPublicStorePage, traffic.ClassCommunityWeb, traffic.ClassMarketWeb:
		return true
	default:
		return false
	}
}

func (r blockRuntime) surfaceLabel() string {
	switch r.class {
	case traffic.ClassCommunityWeb:
		return "community web"
	case traffic.ClassMarketWeb:
		return "market web"
	default:
		return "public store page"
	}
}

func blockLikeHTMLResponse(resp *http.Response, body []byte, sniffBytes int) bool {
	if resp == nil {
		return false
	}

	sniff := body
	if len(sniff) > sniffBytes {
		sniff = sniff[:sniffBytes]
	}
	sniffLower := strings.ToLower(string(sniff))
	if sniffLower == "" {
		return false
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	looksHTML := strings.Contains(contentType, "text/html") ||
		bytes.Contains(bytes.ToLower(bytes.TrimSpace(sniff)), []byte("<!doctype html")) ||
		bytes.Contains(bytes.ToLower(bytes.TrimSpace(sniff)), []byte("<html"))
	if !looksHTML {
		return false
	}

	for _, marker := range blockChallengeMarkers {
		if strings.Contains(sniffLower, marker) {
			return true
		}
	}

	weakMatches := 0
	for _, marker := range blockChallengeWeakMarkers {
		if strings.Contains(sniffLower, marker) {
			weakMatches++
		}
	}
	return weakMatches >= 2
}
