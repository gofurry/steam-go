package httpasset

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Result struct {
	URL           string
	StatusCode    int
	ContentType   string
	ContentLength int64
	BytesWritten  int64
}

func Verify(ctx context.Context, client *http.Client, rawURL string) (Result, bool, error) {
	parsed, err := parseURL(rawURL)
	if err != nil {
		return Result{}, false, err
	}
	client = defaultClient(client)

	result, err := request(ctx, client, http.MethodHead, parsed.String())
	if err != nil {
		return Result{}, false, err
	}
	if result.StatusCode == http.StatusMethodNotAllowed || result.StatusCode == http.StatusNotImplemented {
		result, err = request(ctx, client, http.MethodGet, parsed.String())
		if err != nil {
			return Result{}, false, err
		}
	}
	return result, result.StatusCode >= http.StatusOK && result.StatusCode < http.StatusMultipleChoices, nil
}

func Download(ctx context.Context, client *http.Client, rawURL, path string) (Result, error) {
	parsed, err := parseURL(rawURL)
	if err != nil {
		return Result{}, err
	}
	if strings.TrimSpace(path) == "" {
		return Result{}, fmt.Errorf("path must not be empty")
	}
	client = defaultClient(client)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return Result{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	result := resultFromResponse(parsed.String(), resp)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, resp.Body)
		return result, fmt.Errorf("download %s returned HTTP %d", parsed.String(), resp.StatusCode)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return result, err
	}
	file, err := os.Create(path)
	if err != nil {
		return result, err
	}
	defer file.Close()

	written, err := io.Copy(file, resp.Body)
	result.BytesWritten = written
	if err != nil {
		return result, err
	}
	return result, nil
}

func Read(ctx context.Context, client *http.Client, rawURL string, maxBytes int64) (Result, []byte, error) {
	parsed, err := parseURL(rawURL)
	if err != nil {
		return Result{}, nil, err
	}
	if maxBytes < 1 {
		return Result{}, nil, fmt.Errorf("max bytes must be greater than zero")
	}
	client = defaultClient(client)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return Result{}, nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return Result{}, nil, err
	}
	defer resp.Body.Close()

	result := resultFromResponse(parsed.String(), resp)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, resp.Body)
		return result, nil, fmt.Errorf("read %s returned HTTP %d", parsed.String(), resp.StatusCode)
	}
	if result.ContentLength > maxBytes {
		_, _ = io.Copy(io.Discard, resp.Body)
		return result, nil, fmt.Errorf("read %s content length %d exceeds max bytes %d", parsed.String(), result.ContentLength, maxBytes)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return result, nil, err
	}
	if int64(len(data)) > maxBytes {
		return result, nil, fmt.Errorf("read %s exceeded max bytes %d", parsed.String(), maxBytes)
	}
	result.BytesWritten = int64(len(data))
	return result, data, nil
}

func Filename(rawURL string) (string, error) {
	parsed, err := parseURL(rawURL)
	if err != nil {
		return "", err
	}
	name := filepath.Base(parsed.Path)
	if name == "." || name == "/" || strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("url path has no file name")
	}
	return name, nil
}

func parseURL(rawURL string) (*url.URL, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, fmt.Errorf("url must not be empty")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("url scheme must be http or https")
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("url host must not be empty")
	}
	return parsed, nil
}

func request(ctx context.Context, client *http.Client, method, rawURL string) (Result, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return Result{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	if method == http.MethodGet {
		_, _ = io.Copy(io.Discard, resp.Body)
	}
	return resultFromResponse(rawURL, resp), nil
}

func resultFromResponse(rawURL string, resp *http.Response) Result {
	return Result{
		URL:           rawURL,
		StatusCode:    resp.StatusCode,
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: contentLength(resp),
	}
}

func contentLength(resp *http.Response) int64 {
	if resp.ContentLength >= 0 {
		return resp.ContentLength
	}
	raw := strings.TrimSpace(resp.Header.Get("Content-Length"))
	if raw == "" {
		return 0
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func defaultClient(client *http.Client) *http.Client {
	if client != nil {
		return client
	}
	return http.DefaultClient
}
