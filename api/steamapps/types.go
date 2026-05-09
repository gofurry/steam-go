package steamapps

import "encoding/json"

// GetSDRConfigResponse matches ISteamApps/GetSDRConfig/v1.
type GetSDRConfigResponse struct {
	Revision       int64             `json:"revision"`
	Pops           map[string]SDRPop `json:"pops"`
	Certs          []string          `json:"certs"`
	P2PShareIP     map[string]int    `json:"p2p_share_ip"`
	RelayPublicKey string            `json:"relay_public_key"`
	RevokedKeys    []string          `json:"revoked_keys"`
	TypicalPings   []SDRTypicalPing  `json:"typical_pings"`
	Success        bool              `json:"success"`
}

// SDRPop matches one Steam Datagram Relay POP entry.
type SDRPop struct {
	Aliases  []string   `json:"aliases"`
	Desc     string     `json:"desc"`
	Geo      [2]float64 `json:"geo"`
	Partners int        `json:"partners"`
	Tier     int        `json:"tier"`
	Relays   []SDRRelay `json:"relays"`
}

// SDRRelay matches one relay entry in a POP.
type SDRRelay struct {
	IPv4      string    `json:"ipv4"`
	PortRange [2]uint32 `json:"port_range"`
}

// SDRTypicalPing matches one pairwise ping tuple in GetSDRConfig.
type SDRTypicalPing struct {
	FromPOP string
	ToPOP   string
	PingMS  int
}

// UnmarshalJSON decodes a compact tuple form like ["ams","fra",5].
func (p *SDRTypicalPing) UnmarshalJSON(data []byte) error {
	var raw [3]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[0], &p.FromPOP); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[1], &p.ToPOP); err != nil {
		return err
	}
	return json.Unmarshal(raw[2], &p.PingMS)
}

// GetServersAtAddressResponse matches ISteamApps/GetServersAtAddress/v1.
type GetServersAtAddressResponse struct {
	Response ServersAtAddressPayload `json:"response"`
}

// ServersAtAddressPayload is the top-level server listing payload.
type ServersAtAddressPayload struct {
	Success bool                   `json:"success"`
	Servers []ServersAtAddressItem `json:"servers"`
}

// ServersAtAddressItem matches one server returned by GetServersAtAddress.
type ServersAtAddressItem struct {
	Addr     string `json:"addr"`
	GMSIndex int    `json:"gmsindex"`
	SteamID  string `json:"steamid"`
	AppID    uint32 `json:"appid"`
	GameDir  string `json:"gamedir"`
	Region   int    `json:"region"`
	Secure   bool   `json:"secure"`
	LAN      bool   `json:"lan"`
	GamePort uint32 `json:"gameport"`
	SpecPort uint32 `json:"specport"`
}

// UpToDateCheckResponse matches ISteamApps/UpToDateCheck/v1.
type UpToDateCheckResponse struct {
	Response UpToDateCheckPayload `json:"response"`
}

// UpToDateCheckPayload is the top-level update-check payload.
type UpToDateCheckPayload struct {
	Success           bool   `json:"success"`
	UpToDate          bool   `json:"up_to_date"`
	VersionIsListable bool   `json:"version_is_listable"`
	RequiredVersion   uint32 `json:"required_version"`
	Message           string `json:"message"`
}
