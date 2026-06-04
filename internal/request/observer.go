package request

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/traffic"
)

// RequestObserver receives one sanitized event after a request completes.
type RequestObserver interface {
	ObserveRequest(RequestEvent)
}

// RequestObserverFunc adapts one function into a RequestObserver.
type RequestObserverFunc func(RequestEvent)

// ObserveRequest implements RequestObserver.
func (f RequestObserverFunc) ObserveRequest(event RequestEvent) {
	f(event)
}

// RequestEvent contains sanitized request execution metadata.
type RequestEvent struct {
	TrafficClass  traffic.Class
	Method        string
	Host          string
	Path          string
	StatusCode    int
	ErrorKind     string
	Attempts      int
	CacheHit      bool
	BlockDetected bool
	Duration      time.Duration
}

func observeRequest(observer RequestObserver, req *http.Request, class traffic.Class, statusCode int, err error, attempts int, cacheHit, blockDetected bool, started time.Time) {
	if observer == nil {
		return
	}
	observer.ObserveRequest(RequestEvent{
		TrafficClass:  traffic.NormalizeClass(class),
		Method:        requestMethod(req),
		Host:          requestHost(req),
		Path:          requestPath(req),
		StatusCode:    statusCode,
		ErrorKind:     requestErrorKind(err),
		Attempts:      attempts,
		CacheHit:      cacheHit,
		BlockDetected: blockDetected,
		Duration:      time.Since(started),
	})
}

func requestMethod(req *http.Request) string {
	if req == nil {
		return ""
	}
	return req.Method
}

func requestHost(req *http.Request) string {
	if req == nil || req.URL == nil {
		return ""
	}
	return req.URL.Host
}

func requestPath(req *http.Request) string {
	if req == nil {
		return ""
	}
	return pathOnly(req.URL)
}

func pathOnly(u *url.URL) string {
	if u == nil {
		return ""
	}
	if u.Path == "" {
		return "/"
	}
	return u.Path
}

func requestErrorKind(err error) string {
	if err == nil {
		return ""
	}
	var apiErr *sdkerrors.APIError
	if errors.As(err, &apiErr) && apiErr != nil {
		return string(apiErr.Kind)
	}
	return string(sdkerrors.KindTransport)
}
