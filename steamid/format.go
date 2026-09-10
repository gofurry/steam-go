package steamid

import "fmt"

// Steam2 returns canonical STEAM_X:Y:Z text. Public uses STEAM_1, including
// IDs originally parsed from STEAM_0. Only Individual Desktop IDs are losslessly
// representable. Invalid IDs return ErrInvalidID; other valid IDs return
// ErrUnsupportedConversion.
func (id ID) Steam2() (string, error) {
	if !id.Valid() {
		return "", ErrInvalidID
	}
	if id.AccountType() != AccountTypeIndividual || id.Instance() != InstanceDesktop {
		return "", ErrUnsupportedConversion
	}
	return fmt.Sprintf("STEAM_%d:%d:%d", id.Universe(), id.AccountID()&1, id.AccountID()>>1), nil
}

// Steam3 returns canonical colon-separated Steam3 text without losing instance
// bits. Multiseat and anonymous game server IDs always include the instance;
// other types omit it only when it matches the type character's default.
// ConsoleUser has no supported type character and returns
// ErrUnsupportedConversion. Invalid IDs return ErrInvalidID.
func (id ID) Steam3() (string, error) {
	if !id.Valid() {
		return "", ErrInvalidID
	}
	var char byte
	switch id.AccountType() {
	case AccountTypeIndividual:
		char = 'U'
	case AccountTypeMultiseat:
		char = 'M'
	case AccountTypeGameServer:
		char = 'G'
	case AccountTypeAnonGameServer:
		char = 'A'
	case AccountTypePending:
		char = 'P'
	case AccountTypeContentServer:
		char = 'C'
	case AccountTypeClan:
		char = 'g'
	case AccountTypeChat:
		char = 'T'
		if id.Instance()&chatInstanceClan != 0 {
			char = 'c'
		} else if id.Instance()&chatInstanceLobby != 0 {
			char = 'L'
		}
	case AccountTypeAnonUser:
		char = 'a'
	default:
		return "", ErrUnsupportedConversion
	}
	_, base, flag, _ := steam3Type(char)
	if id.Instance() != base|flag || id.AccountType() == AccountTypeMultiseat || id.AccountType() == AccountTypeAnonGameServer {
		return fmt.Sprintf("[%c:%d:%d:%d]", char, id.Universe(), id.AccountID(), id.Instance()), nil
	}
	return fmt.Sprintf("[%c:%d:%d]", char, id.Universe(), id.AccountID()), nil
}
