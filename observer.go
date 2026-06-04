package steam

import "github.com/gofurry/steam-go/internal/request"

// RequestObserver receives one sanitized event after an SDK request completes.
type RequestObserver = request.RequestObserver

// RequestObserverFunc adapts one function into a RequestObserver.
type RequestObserverFunc = request.RequestObserverFunc

// RequestEvent contains sanitized request execution metadata.
//
// Path never includes the raw query string. Headers, body, credentials, cookies,
// and proxy passwords are not included in this event.
type RequestEvent = request.RequestEvent
