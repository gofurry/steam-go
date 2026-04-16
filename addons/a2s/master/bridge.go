package master

import (
	upstream "github.com/GoFurry/a2s-go/master"
)

type Client = upstream.Client
type Option = upstream.Option
type Error = upstream.Error
type ErrorCode = upstream.ErrorCode
type Request = upstream.Request
type Page = upstream.Page
type Result = upstream.Result
type ServerAddr = upstream.ServerAddr
type Cursor = upstream.Cursor
type Region = upstream.Region

const (
	ErrorCodeAddress      = upstream.ErrorCodeAddress
	ErrorCodeDial         = upstream.ErrorCodeDial
	ErrorCodeWrite        = upstream.ErrorCodeWrite
	ErrorCodeRead         = upstream.ErrorCodeRead
	ErrorCodeTimeout      = upstream.ErrorCodeTimeout
	ErrorCodePacketHeader = upstream.ErrorCodePacketHeader
	ErrorCodeDecode       = upstream.ErrorCodeDecode
	ErrorCodeFilter       = upstream.ErrorCodeFilter
	ErrorCodeCursor       = upstream.ErrorCodeCursor
	ErrorCodeRegion       = upstream.ErrorCodeRegion

	RegionUSEast       = upstream.RegionUSEast
	RegionUSWest       = upstream.RegionUSWest
	RegionSouthAmerica = upstream.RegionSouthAmerica
	RegionEurope       = upstream.RegionEurope
	RegionAsia         = upstream.RegionAsia
	RegionAustralia    = upstream.RegionAustralia
	RegionMiddleEast   = upstream.RegionMiddleEast
	RegionAfrica       = upstream.RegionAfrica
	RegionRestOfWorld  = upstream.RegionRestOfWorld
)

var (
	StartCursor       = upstream.StartCursor
	RegionCustom      = upstream.RegionCustom
	WithTimeout       = upstream.WithTimeout
	WithBaseAddress   = upstream.WithBaseAddress
	WithMaxPacketSize = upstream.WithMaxPacketSize
)

func NewClient(opts ...Option) (*Client, error) {
	return upstream.NewClient(opts...)
}
