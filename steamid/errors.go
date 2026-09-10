package steamid

import "errors"

var (
	// ErrInvalidFormat means the input does not match the requested text format.
	ErrInvalidFormat = errors.New("steamid: invalid format")
	// ErrInvalidID means a number is out of range or its components are invalid.
	ErrInvalidID = errors.New("steamid: invalid identity")
	// ErrUnsupportedConversion means a valid ID cannot be represented losslessly.
	ErrUnsupportedConversion = errors.New("steamid: unsupported conversion")
	// ErrVanityReference means a Community vanity URL requires network resolution.
	ErrVanityReference = errors.New("steamid: vanity reference requires resolution")
)
