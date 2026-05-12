package request

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	sdkerrors "github.com/GoFurry/steam-go/internal/errors"
	"github.com/GoFurry/steam-go/internal/traffic"
)

const defaultBlockHTMLSniffBytes = 8 * 1024

var blockChallengeMarkers = []string{
	"captcha",
	"verify you are human",
	"access denied",
	"cf-browser-verification",
	"cloudflare",
	"g-recaptcha",
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
	if class != traffic.ClassPublicStorePage {
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
			Message:   fmt.Sprintf("public store page block detected: rate_limited (status=%d)", resp.StatusCode),
			Retryable: true,
		}
	case http.StatusForbidden:
		return &BlockResult{
			ErrorKind: sdkerrors.KindHTTPStatus,
			Message:   fmt.Sprintf("public store page block detected: forbidden (status=%d)", resp.StatusCode),
			Retryable: true,
		}
	}

	if !blockLikeHTMLResponse(resp, body, r.htmlSniffBytes) {
		return nil
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return &BlockResult{
			ErrorKind: sdkerrors.KindAPIResponse,
			Message:   "public store page block detected: challenge html response",
			Retryable: false,
		}
	}
	if resp.StatusCode >= http.StatusInternalServerError {
		return &BlockResult{
			ErrorKind: sdkerrors.KindHTTPStatus,
			Message:   fmt.Sprintf("public store page block detected: challenge html response (status=%d)", resp.StatusCode),
			Retryable: true,
		}
	}

	return nil
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
	return false
}
