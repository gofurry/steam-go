package errors

import "fmt"

// Kind identifies where an SDK failure happened.
type Kind string

const (
	KindRequestBuild Kind = "request_build"
	KindTransport    Kind = "transport"
	KindHTTPStatus   Kind = "http_status"
	KindDecode       Kind = "decode"
	KindAPIResponse  Kind = "api_response"
)

// APIError is the common SDK error shape.
type APIError struct {
	Kind       Kind
	StatusCode int
	Message    string
	Body       []byte
	Err        error
}

func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.StatusCode > 0 {
		return fmt.Sprintf("%s: %s (status=%d)", e.Kind, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

// Unwrap exposes the wrapped error.
func (e *APIError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// New constructs a typed SDK error.
func New(kind Kind, statusCode int, message string, body []byte, err error) *APIError {
	return &APIError{
		Kind:       kind,
		StatusCode: statusCode,
		Message:    message,
		Body:       body,
		Err:        err,
	}
}
