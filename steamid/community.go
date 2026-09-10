package steamid

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseCommunityURL parses an http(s) numeric /profiles/<SteamID64> URL on
// steamcommunity.com or www.steamcommunity.com. One trailing slash, query, and
// fragment are allowed. The ID must be a valid Individual. Userinfo and other
// paths are rejected. Vanity /id/<name> URLs return ErrVanityReference; resolving
// those requires an explicit call to client.API.SteamUser.ResolveVanityURL.
func ParseCommunityURL(value string) (ID, error) {
	u, err := url.Parse(strings.TrimSpace(value))
	if err != nil || u.User != nil || (!strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https")) {
		return 0, ErrInvalidFormat
	}
	host := u.Hostname()
	if !strings.EqualFold(host, "steamcommunity.com") && !strings.EqualFold(host, "www.steamcommunity.com") {
		return 0, ErrInvalidFormat
	}
	parts := strings.SplitN(strings.TrimSuffix(u.Path, "/"), "/", 4)
	if len(parts) != 3 || parts[0] != "" || parts[2] == "" {
		return 0, ErrInvalidFormat
	}
	switch parts[1] {
	case "id":
		return 0, ErrVanityReference
	case "profiles":
		// URL path components must be decimal as written, without the outer
		// whitespace trimming accepted by the standalone SteamID64 parser.
		n, err := decimal(parts[2], 64)
		if err != nil {
			return 0, err
		}
		id, err := validPacked(n)
		if err != nil {
			return 0, err
		}
		if id.AccountType() != AccountTypeIndividual {
			return 0, fmt.Errorf("%w: profile URL requires an Individual", ErrInvalidID)
		}
		return id, nil
	default:
		return 0, ErrInvalidFormat
	}
}
