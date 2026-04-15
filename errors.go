package steam

import sdkerrors "github.com/GoFurry/steam-go/internal/errors"

// ErrorKind classifies SDK errors by lifecycle stage.
type ErrorKind = sdkerrors.Kind

const (
	ErrorKindRequestBuild = sdkerrors.KindRequestBuild
	ErrorKindTransport    = sdkerrors.KindTransport
	ErrorKindHTTPStatus   = sdkerrors.KindHTTPStatus
	ErrorKindDecode       = sdkerrors.KindDecode
	ErrorKindAPIResponse  = sdkerrors.KindAPIResponse
)

// APIError is the exported SDK error model.
type APIError = sdkerrors.APIError
