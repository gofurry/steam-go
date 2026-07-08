package gameserversservice

import (
	"strconv"
	"strings"
)

const (
	// ServerTypeDedicated matches dedicated servers.
	ServerTypeDedicated = "d"
	// ServerTypeListen matches listen or non-dedicated servers.
	ServerTypeListen = "l"
	// ServerTypeSourceTV matches SourceTV or proxy servers.
	ServerTypeSourceTV = "p"
)

// Filter builds Master Server style GetServerList filter strings.
//
// Filter values are immutable. Each method returns a new Filter that can be
// safely derived from an existing one without changing the original.
type Filter struct {
	terms []filterTerm
}

type filterTerm struct {
	key   string
	value string
}

// NewFilter returns an empty server-list filter.
func NewFilter() Filter {
	return Filter{}
}

// String renders the filter as a continuous \key\value string.
func (f Filter) String() string {
	if len(f.terms) == 0 {
		return ""
	}

	var b strings.Builder
	for _, term := range f.terms {
		b.WriteByte('\\')
		b.WriteString(term.key)
		b.WriteByte('\\')
		b.WriteString(term.value)
	}
	return b.String()
}

// AppID restricts results to one Steam app ID.
func (f Filter) AppID(appID uint32) Filter {
	return f.addUint("appid", appID)
}

// NotAppID excludes one Steam app ID.
func (f Filter) NotAppID(appID uint32) Filter {
	return f.addUint("napp", appID)
}

// Secure restricts results to secure or VAC-enabled servers.
func (f Filter) Secure() Filter {
	return f.addFlag("secure")
}

// GameDir restricts results by game directory or mod directory.
func (f Filter) GameDir(value string) Filter {
	return f.add("gamedir", value)
}

// Map restricts results by the current map name.
func (f Filter) Map(value string) Filter {
	return f.add("map", value)
}

// Linux restricts results to Linux servers.
func (f Filter) Linux() Filter {
	return f.addFlag("linux")
}

// NotEmpty excludes empty servers.
func (f Filter) NotEmpty() Filter {
	return f.addFlag("empty")
}

// NotFull excludes full servers.
func (f Filter) NotFull() Filter {
	return f.addFlag("full")
}

// NoPlayers restricts results to servers with no players.
func (f Filter) NoPlayers() Filter {
	return f.addFlag("noplayers")
}

// Proxy restricts results to SourceTV or proxy servers.
func (f Filter) Proxy() Filter {
	return f.addFlag("proxy")
}

// White restricts results to whitelisted servers.
func (f Filter) White() Filter {
	return f.addFlag("white")
}

// Dedicated restricts results to dedicated servers.
func (f Filter) Dedicated() Filter {
	return f.addFlag("dedicated")
}

// NoPassword excludes password-protected servers.
func (f Filter) NoPassword() Filter {
	return f.add("password", "0")
}

// Password restricts results to password-protected servers.
func (f Filter) Password() Filter {
	return f.add("password", "1")
}

// GameType matches public server tags such as sv_tags values.
func (f Filter) GameType(tags ...string) Filter {
	return f.addList("gametype", tags)
}

// GameData matches all hidden game tags where supported by the game.
func (f Filter) GameData(tags ...string) Filter {
	return f.addList("gamedata", tags)
}

// GameDataOr matches any hidden game tag where supported by the game.
func (f Filter) GameDataOr(tags ...string) Filter {
	return f.addList("gamedataor", tags)
}

// NameMatch matches server names using the upstream wildcard syntax.
func (f Filter) NameMatch(pattern string) Filter {
	return f.add("name_match", pattern)
}

// VersionMatch matches server versions using the upstream wildcard syntax.
func (f Filter) VersionMatch(pattern string) Filter {
	return f.add("version_match", pattern)
}

// CollapseAddrHash asks upstream to return at most one server per unique IP.
func (f Filter) CollapseAddrHash() Filter {
	return f.addFlag("collapse_addr_hash")
}

// GameAddr restricts results to one IP address or IP:port pair.
func (f Filter) GameAddr(addr string) Filter {
	return f.add("gameaddr", addr)
}

// Type restricts results by server type. Use ServerTypeDedicated,
// ServerTypeListen, or ServerTypeSourceTV when possible.
func (f Filter) Type(value string) Filter {
	return f.add("type", value)
}

// Raw appends one filter key/value pair for upstream filters not yet modeled by
// this package.
func (f Filter) Raw(key, value string) Filter {
	return f.add(key, value)
}

func (f Filter) addFlag(key string) Filter {
	return f.add(key, "1")
}

func (f Filter) addUint(key string, value uint32) Filter {
	return f.add(key, strconv.FormatUint(uint64(value), 10))
}

func (f Filter) addList(key string, values []string) Filter {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strings.TrimSpace(value))
	}
	return f.add(key, strings.Join(parts, ","))
}

func (f Filter) add(key, value string) Filter {
	terms := make([]filterTerm, len(f.terms), len(f.terms)+1)
	copy(terms, f.terms)
	terms = append(terms, filterTerm{
		key:   strings.TrimSpace(key),
		value: strings.TrimSpace(value),
	})
	return Filter{terms: terms}
}
