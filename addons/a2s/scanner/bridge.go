package scanner

import (
	masterupstream "github.com/GoFurry/a2s-go/master"
	upstream "github.com/GoFurry/a2s-go/scanner"
)

type Client = upstream.Client
type Option = upstream.Option
type Request = upstream.Request
type Result = upstream.Result
type PlayersResult = upstream.PlayersResult
type RulesResult = upstream.RulesResult
type Error = upstream.Error
type ErrorCode = upstream.ErrorCode

const (
	ErrorCodeInput       = upstream.ErrorCodeInput
	ErrorCodeConcurrency = upstream.ErrorCodeConcurrency
	ErrorCodeTimeout     = upstream.ErrorCodeTimeout
	ErrorCodePacketSize  = upstream.ErrorCodePacketSize
	ErrorCodeDiscovery   = upstream.ErrorCodeDiscovery
	ErrorCodeProbe       = upstream.ErrorCodeProbe
)

var (
	WithConcurrency   = upstream.WithConcurrency
	WithTimeout       = upstream.WithTimeout
	WithMaxPacketSize = upstream.WithMaxPacketSize
)

func NewClient(opts ...Option) (*Client, error) {
	return upstream.NewClient(opts...)
}

func ParseAddress(addr string) (masterupstream.ServerAddr, error) {
	return upstream.ParseAddress(addr)
}

func ParseAddresses(addrs []string) ([]masterupstream.ServerAddr, error) {
	return upstream.ParseAddresses(addrs)
}
