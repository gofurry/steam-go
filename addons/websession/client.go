package websession

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofurry/steam-go/api/authenticationservice"
)

type Client struct {
	auth                 *authenticationservice.Service
	httpClient           *http.Client
	timeout              time.Duration
	maxResponseBodyBytes int64
	loginBaseURL         *url.URL
	storeBaseURL         *url.URL
	communityBaseURL     *url.URL
}

func NewClient(auth *authenticationservice.Service, opts ...Option) (*Client, error) {
	if auth == nil {
		return nil, configError("new_client", "authentication service must not be nil", nil)
	}
	options, err := defaultClientOptions()
	if err != nil {
		return nil, configError("new_client", "invalid default options", err)
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&options); err != nil {
			return nil, err
		}
	}
	return &Client{
		auth:                 auth,
		httpClient:           options.httpClient,
		timeout:              options.timeout,
		maxResponseBodyBytes: options.maxResponseBodyBytes,
		loginBaseURL:         cloneURL(options.loginBaseURL),
		storeBaseURL:         cloneURL(options.storeBaseURL),
		communityBaseURL:     cloneURL(options.communityBaseURL),
	}, nil
}

func (c *Client) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c.timeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, c.timeout)
}

func readBodyLimited(r io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return io.ReadAll(r)
	}
	reader := &io.LimitedReader{R: r, N: maxBytes + 1}
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("response body exceeds limit of %d bytes", maxBytes)
	}
	return body, nil
}

func parseAbsoluteURL(raw string) (*url.URL, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("url must not be empty")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("url must be absolute")
	}
	return parsed, nil
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	cloned := *u
	return &cloned
}

func resolveURL(base *url.URL, path string) *url.URL {
	if base == nil {
		return &url.URL{Path: path}
	}
	ref := &url.URL{Path: path}
	return cloneURL(base).ResolveReference(ref)
}
