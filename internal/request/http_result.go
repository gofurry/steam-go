package request

import (
	"net/http"
	"net/url"
)

// HTTPResult captures one raw HTTP response after retries, cache, and block detection.
type HTTPResult struct {
	StatusCode int
	Header     http.Header
	FinalURL   *url.URL
	Body       []byte
	Block      *BlockResult
}

func cloneHTTPResult(src HTTPResult) HTTPResult {
	cloned := HTTPResult{
		StatusCode: src.StatusCode,
		Header:     cloneHeader(src.Header),
		FinalURL:   cloneURL(src.FinalURL),
		Body:       cloneBytes(src.Body),
	}
	if src.Block != nil {
		block := *src.Block
		cloned.Block = &block
	}
	return cloned
}

func cloneHeader(src http.Header) http.Header {
	if len(src) == 0 {
		return nil
	}
	cloned := make(http.Header, len(src))
	for key, values := range src {
		copied := make([]string, len(values))
		copy(copied, values)
		cloned[key] = copied
	}
	return cloned
}

func cloneURL(src *url.URL) *url.URL {
	if src == nil {
		return nil
	}
	cloned := *src
	return &cloned
}
