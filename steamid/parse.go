package steamid

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse trims whitespace and parses Steam2, Steam3, or unsigned decimal text.
// Decimal values <= MaxUint32 are Public Individual Desktop AccountIDs; larger
// values are packed SteamID64 values. Zero is invalid. URLs require the separate
// ParseCommunityURL function. Every successful result satisfies Valid.
func Parse(value string) (ID, error) {
	value = strings.TrimSpace(value)
	switch {
	case len(value) >= 6 && strings.EqualFold(value[:6], "STEAM_"):
		return ParseSteam2(value)
	case strings.HasPrefix(value, "["):
		return ParseSteam3(value)
	}
	n, err := decimal(value, 64)
	if err != nil {
		return 0, err
	}
	if n <= accountIDMask {
		return NewIndividual(uint32(n))
	}
	return validPacked(n)
}

// ParseSteamID64 parses unsigned decimal packed SteamID64 text after trimming
// whitespace. Signs, hexadecimal, and floating-point notation are not accepted.
// Small integers that are not valid packed identities return ErrInvalidID.
func ParseSteamID64(value string) (ID, error) {
	n, err := decimal(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, err
	}
	return validPacked(n)
}

func validPacked(n uint64) (ID, error) {
	id := ID(n)
	if !id.Valid() {
		return 0, ErrInvalidID
	}
	return id, nil
}

// ParseSteam2 parses STEAM_X:Y:Z as an Individual Desktop ID. The prefix is
// case-insensitive. Historical universe 0 is normalized to Public (1).
func ParseSteam2(value string) (ID, error) {
	value = strings.TrimSpace(value)
	if len(value) < 6 || !strings.EqualFold(value[:6], "STEAM_") {
		return 0, ErrInvalidFormat
	}
	parts := strings.SplitN(value[6:], ":", 4)
	if len(parts) != 3 || len(parts[1]) != 1 {
		return 0, ErrInvalidFormat
	}
	universe, err := decimal(parts[0], 8)
	if err != nil {
		return 0, err
	}
	parity, err := decimal(parts[1], 8)
	if err != nil {
		return 0, err
	}
	half, err := decimal(parts[2], 32)
	if err != nil {
		return 0, err
	}
	if parity > 1 || half > accountIDMask/2 {
		return 0, fmt.Errorf("%w: Steam2 parity or account out of range", ErrInvalidID)
	}
	if universe == 0 {
		universe = uint64(UniversePublic)
	}
	return New(uint32(half*2+parity), Universe(universe), AccountTypeIndividual, InstanceDesktop)
}

// ParseSteam3 parses [type:universe:account] with an optional :instance.
// Type characters are case-sensitive. The legacy account(instance) form is
// also accepted. Clan/lobby chat characters restore their instance flag; an
// explicit instance is ORed with that flag without discarding other bits.
func ParseSteam3(value string) (ID, error) {
	value = strings.TrimSpace(value)
	if len(value) < 2 || value[0] != '[' || value[len(value)-1] != ']' {
		return 0, ErrInvalidFormat
	}
	parts := strings.SplitN(value[1:len(value)-1], ":", 5)
	if (len(parts) != 3 && len(parts) != 4) || len(parts[0]) != 1 {
		return 0, ErrInvalidFormat
	}
	accountType, instance, flag, ok := steam3Type(parts[0][0])
	if !ok {
		return 0, fmt.Errorf("%w: unknown Steam3 type character", ErrInvalidFormat)
	}
	instanceText := ""
	explicitInstance := len(parts) == 4
	if explicitInstance {
		instanceText = parts[3]
	} else if account, suffix, found := strings.Cut(parts[2], "("); found {
		if !strings.HasSuffix(suffix, ")") {
			return 0, ErrInvalidFormat
		}
		parts[2] = account
		instanceText = strings.TrimSuffix(suffix, ")")
		explicitInstance = true
	}
	universe, err := decimal(parts[1], 8)
	if err != nil {
		return 0, err
	}
	account, err := decimal(parts[2], 32)
	if err != nil {
		return 0, err
	}
	if explicitInstance {
		n, err := decimal(instanceText, 20)
		if err != nil {
			return 0, err
		}
		instance = uint32(n)
	}
	return New(uint32(account), Universe(universe), accountType, instance|flag)
}

// decimal distinguishes malformed text from a syntactically decimal overflow.
// It deliberately does not include the original input in errors.
func decimal(value string, bits int) (uint64, error) {
	if value == "" {
		return 0, ErrInvalidFormat
	}
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return 0, ErrInvalidFormat
		}
	}
	n, err := strconv.ParseUint(value, 10, bits)
	if err != nil {
		return 0, fmt.Errorf("%w: decimal exceeds %d bits", ErrInvalidID, bits)
	}
	return n, nil
}

// steam3Type returns the account type, default base instance, and implied flag.
func steam3Type(char byte) (AccountType, uint32, uint32, bool) {
	switch char {
	case 'U':
		return AccountTypeIndividual, InstanceDesktop, 0, true
	case 'M':
		return AccountTypeMultiseat, InstanceDesktop, 0, true
	case 'G':
		return AccountTypeGameServer, InstanceDesktop, 0, true
	case 'A':
		return AccountTypeAnonGameServer, InstanceDesktop, 0, true
	case 'P':
		return AccountTypePending, InstanceDesktop, 0, true
	case 'C':
		return AccountTypeContentServer, InstanceDesktop, 0, true
	case 'g':
		return AccountTypeClan, InstanceAll, 0, true
	case 'T':
		return AccountTypeChat, InstanceAll, 0, true
	case 'c':
		return AccountTypeChat, InstanceAll, chatInstanceClan, true
	case 'L':
		return AccountTypeChat, InstanceAll, chatInstanceLobby, true
	case 'a':
		return AccountTypeAnonUser, InstanceDesktop, 0, true
	default:
		return AccountTypeInvalid, 0, 0, false
	}
}
