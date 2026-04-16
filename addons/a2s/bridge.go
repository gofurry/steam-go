package a2s

import (
	"net"

	upstream "github.com/GoFurry/a2s-go"
)

type Client = upstream.Client
type Option = upstream.Option
type Error = upstream.Error
type ErrorCode = upstream.ErrorCode
type ServerType = upstream.ServerType
type Environment = upstream.Environment
type TheShipMode = upstream.TheShipMode
type TheShipInfo = upstream.TheShipInfo
type Info = upstream.Info
type Players = upstream.Players
type TheShipPlayer = upstream.TheShipPlayer
type Player = upstream.Player
type Rules = upstream.Rules

const (
	ErrorCodeAddress      = upstream.ErrorCodeAddress
	ErrorCodeDial         = upstream.ErrorCodeDial
	ErrorCodeWrite        = upstream.ErrorCodeWrite
	ErrorCodeRead         = upstream.ErrorCodeRead
	ErrorCodeTimeout      = upstream.ErrorCodeTimeout
	ErrorCodePacketHeader = upstream.ErrorCodePacketHeader
	ErrorCodeDecode       = upstream.ErrorCodeDecode
	ErrorCodeChallenge    = upstream.ErrorCodeChallenge
	ErrorCodeMultiPacket  = upstream.ErrorCodeMultiPacket
	ErrorCodeUnsupported  = upstream.ErrorCodeUnsupported

	ServerTypeDedicated    = upstream.ServerTypeDedicated
	ServerTypeNonDedicated = upstream.ServerTypeNonDedicated
	ServerTypeSourceTV     = upstream.ServerTypeSourceTV

	EnvironmentLinux   = upstream.EnvironmentLinux
	EnvironmentWindows = upstream.EnvironmentWindows
	EnvironmentMacOS   = upstream.EnvironmentMacOS
	EnvironmentMacOSX  = upstream.EnvironmentMacOSX

	TheShipModeHunt            = upstream.TheShipModeHunt
	TheShipModeElimination     = upstream.TheShipModeElimination
	TheShipModeDuel            = upstream.TheShipModeDuel
	TheShipModeDeathmatch      = upstream.TheShipModeDeathmatch
	TheShipModeTeamVIP         = upstream.TheShipModeTeamVIP
	TheShipModeTeamElimination = upstream.TheShipModeTeamElimination
	TheShipModeUnknown         = upstream.TheShipModeUnknown
)

var (
	WithTimeout       = upstream.WithTimeout
	WithMaxPacketSize = upstream.WithMaxPacketSize
)

func NewClient(addr string, opts ...Option) (*Client, error) {
	return upstream.NewClient(addr, opts...)
}

func NewClientWithConn(addr string, conn *net.UDPConn, opts ...Option) (*Client, error) {
	return upstream.NewClientWithConn(addr, conn, opts...)
}
