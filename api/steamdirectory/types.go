package steamdirectory

// GetCMListForConnectResponse matches ISteamDirectory/GetCMListForConnect/v1.
type GetCMListForConnectResponse struct {
	Response CMListForConnectPayload `json:"response"`
}

// CMListForConnectPayload is the top-level CM directory payload.
type CMListForConnectPayload struct {
	ServerList []CMConnectServer `json:"serverlist"`
	Success    bool              `json:"success"`
	Message    string            `json:"message"`
}

// CMConnectServer matches one CM candidate entry.
type CMConnectServer struct {
	Endpoint       string  `json:"endpoint"`
	LegacyEndpoint string  `json:"legacy_endpoint"`
	Type           string  `json:"type"`
	DC             string  `json:"dc"`
	Realm          string  `json:"realm"`
	Load           int     `json:"load"`
	WTDLoad        float64 `json:"wtd_load"`
}

// GetSteamPipeDomainsResponse matches ISteamDirectory/GetSteamPipeDomains/v1.
type GetSteamPipeDomainsResponse struct {
	Response SteamPipeDomainsPayload `json:"response"`
}

// SteamPipeDomainsPayload is the top-level SteamPipe domains payload.
type SteamPipeDomainsPayload struct {
	DomainList []string `json:"domainlist"`
	Result     int      `json:"result"`
	Message    string   `json:"message"`
}
