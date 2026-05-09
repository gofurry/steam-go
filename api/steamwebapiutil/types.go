package steamwebapiutil

// GetServerInfoResponse matches ISteamWebAPIUtil/GetServerInfo/v1.
type GetServerInfoResponse struct {
	ServerTime       int64  `json:"servertime"`
	ServerTimeString string `json:"servertimestring"`
}

// GetSupportedAPIListResponse matches ISteamWebAPIUtil/GetSupportedAPIList/v1.
type GetSupportedAPIListResponse struct {
	APIList SupportedAPIList `json:"apilist"`
}

// SupportedAPIList is the top-level API list payload.
type SupportedAPIList struct {
	Interfaces []SupportedAPIInterface `json:"interfaces"`
}

// SupportedAPIInterface describes a Steam Web API interface.
type SupportedAPIInterface struct {
	Name    string               `json:"name"`
	Methods []SupportedAPIMethod `json:"methods"`
}

// SupportedAPIMethod describes one Steam Web API method.
type SupportedAPIMethod struct {
	Name        string                  `json:"name"`
	Version     int                     `json:"version"`
	HTTPMethod  string                  `json:"httpmethod"`
	Description string                  `json:"description"`
	Parameters  []SupportedAPIParameter `json:"parameters"`
}

// SupportedAPIParameter describes one method parameter.
type SupportedAPIParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Optional    bool   `json:"optional"`
	Description string `json:"description"`
}
