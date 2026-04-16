package openid

import (
	"fmt"
)

type ErrorCode string

const (
	ErrorCodeConfig       ErrorCode = "config"
	ErrorCodeRequestBuild ErrorCode = "request_build"
	ErrorCodeTransport    ErrorCode = "transport"
	ErrorCodeHTTPStatus   ErrorCode = "http_status"
	ErrorCodeVerify       ErrorCode = "verify"
	ErrorCodeIdentity     ErrorCode = "identity"
)

type Error struct {
	Code    ErrorCode
	Op      string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}

	if e.Message == "" {
		if e.Err != nil {
			return fmt.Sprintf("%s: %s: %v", e.Code, e.Op, e.Err)
		}
		return fmt.Sprintf("%s: %s", e.Code, e.Op)
	}

	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %s: %v", e.Code, e.Op, e.Message, e.Err)
	}

	return fmt.Sprintf("%s: %s: %s", e.Code, e.Op, e.Message)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
