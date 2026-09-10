package steamid

// Universe identifies a Steam environment.
type Universe uint8

// Steamworks universe values. UniverseInvalid is not a valid identity universe.
const (
	UniverseInvalid Universe = iota
	UniversePublic
	UniverseBeta
	UniverseInternal
	UniverseDev
)

// AccountType identifies the kind of Steam account encoded in an ID.
type AccountType uint8

// Steamworks account types. AccountTypeInvalid is not a valid identity type.
const (
	AccountTypeInvalid AccountType = iota
	AccountTypeIndividual
	AccountTypeMultiseat
	AccountTypeGameServer
	AccountTypeAnonGameServer
	AccountTypePending
	AccountTypeContentServer
	AccountTypeClan
	AccountTypeChat
	AccountTypeConsoleUser
	AccountTypeAnonUser
)

// Common individual account instances.
const (
	InstanceAll     uint32 = 0
	InstanceDesktop uint32 = 1
	InstanceConsole uint32 = 2
	InstanceWeb     uint32 = 4
)
