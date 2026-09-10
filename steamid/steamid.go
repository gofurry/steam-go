// Package steamid parses and converts Steam identities locally, without network
// requests or credentials. Validation checks structure, not account existence.
//
// Parse accepts decimal AccountIDs, packed SteamID64 values, Steam2, and Steam3.
// Use ParseCommunityURL explicitly for numeric Community profile URLs. Vanity
// URLs return ErrVanityReference; callers can resolve those separately through
// client.API.SteamUser.ResolveVanityURL.
//
// This package does not change the SDK's existing request validation semantics.
package steamid

import (
	"fmt"
	"strconv"
)

const (
	accountIDMask    = 0xFFFFFFFF
	instanceMask     = 0x000FFFFF
	accountTypeMask  = 0xF
	universeMask     = 0xFF
	instanceShift    = 32
	accountTypeShift = 52
	universeShift    = 56

	chatInstanceClan  uint32 = 1 << 19
	chatInstanceLobby uint32 = 1 << 18
)

// ID stores a packed SteamID64. Its zero value is invalid. A direct cast from
// uint64 may also produce an invalid ID; use Valid to check it independently.
type ID uint64

// New constructs an ID without truncating components. Invalid components,
// including an instance exceeding 20 bits, return ErrInvalidID.
func New(accountID uint32, universe Universe, accountType AccountType, instance uint32) (ID, error) {
	if universe <= UniverseInvalid || universe > UniverseDev {
		return 0, fmt.Errorf("%w: universe out of range", ErrInvalidID)
	}
	if accountType <= AccountTypeInvalid || accountType > AccountTypeAnonUser {
		return 0, fmt.Errorf("%w: account type out of range", ErrInvalidID)
	}
	if instance > instanceMask {
		return 0, fmt.Errorf("%w: instance exceeds 20 bits", ErrInvalidID)
	}
	id := ID(uint64(accountID) | uint64(instance)<<instanceShift |
		uint64(accountType)<<accountTypeShift | uint64(universe)<<universeShift)
	if !id.Valid() {
		return 0, fmt.Errorf("%w: account and instance combination", ErrInvalidID)
	}
	return id, nil
}

// NewIndividual interprets accountID as a Public Individual Desktop account.
// AccountID zero returns ErrInvalidID.
func NewIndividual(accountID uint32) (ID, error) {
	return New(accountID, UniversePublic, AccountTypeIndividual, InstanceDesktop)
}

// Uint64 returns the raw packed SteamID64 value.
func (id ID) Uint64() uint64 { return uint64(id) }

// String returns canonical decimal SteamID64 text, even for an invalid ID.
func (id ID) String() string { return strconv.FormatUint(uint64(id), 10) }

// AccountID returns the low 32 bits; it carries no universe, type, or instance.
func (id ID) AccountID() uint32 { return uint32(uint64(id) & accountIDMask) }

// Instance returns the 20-bit account instance, including any chat flags.
func (id ID) Instance() uint32 { return uint32(uint64(id) >> instanceShift & instanceMask) }

// Universe returns the 8-bit Steam universe.
func (id ID) Universe() Universe { return Universe(uint64(id) >> universeShift & universeMask) }

// AccountType returns the 4-bit account type.
func (id ID) AccountType() AccountType {
	return AccountType(uint64(id) >> accountTypeShift & accountTypeMask)
}

// Valid checks structural Steam identity rules, not whether an account exists.
// Universes Public through Dev and types Individual through AnonUser are known.
// Individuals require a nonzero AccountID and instance <= Web (including 3).
// Clans require a nonzero AccountID and instance 0; GameServers require a
// nonzero AccountID. Other known types have no additional component constraints.
func (id ID) Valid() bool {
	if id.Universe() <= UniverseInvalid || id.Universe() > UniverseDev ||
		id.AccountType() <= AccountTypeInvalid || id.AccountType() > AccountTypeAnonUser {
		return false
	}
	switch id.AccountType() {
	case AccountTypeIndividual:
		return id.AccountID() != 0 && id.Instance() <= InstanceWeb
	case AccountTypeClan:
		return id.AccountID() != 0 && id.Instance() == InstanceAll
	case AccountTypeGameServer:
		return id.AccountID() != 0
	default:
		return true
	}
}
