package contentserverdirectoryservice

import "encoding/json"

// GetCDNForVideoRequest contains the CDN video lookup query.
type GetCDNForVideoRequest struct {
	PropertyType int32
	ClientIP     string
	ClientRegion string
}

// GetCDNForVideoResponse keeps Steam's video CDN payload raw because the nested
// shape is not stable across public inventory snapshots.
type GetCDNForVideoResponse struct {
	Response json.RawMessage `json:"response"`
}

// GetClientUpdateHostsResponse matches IContentServerDirectoryService/GetClientUpdateHosts/v1.
type GetClientUpdateHostsResponse struct {
	Response ClientUpdateHostsPayload `json:"response"`
}

// ClientUpdateHostsPayload contains Steam client update host metadata.
type ClientUpdateHostsPayload struct {
	HostsKV        string `json:"hosts_kv"`
	ValidUntilTime uint64 `json:"valid_until_time"`
	IPCountry      string `json:"ip_country"`
}

// GetDepotPatchInfoRequest contains one depot patch lookup.
type GetDepotPatchInfoRequest struct {
	AppID            uint32
	DepotID          uint32
	SourceManifestID uint64
	TargetManifestID uint64
}

// GetDepotPatchInfoResponse matches IContentServerDirectoryService/GetDepotPatchInfo/v1.
type GetDepotPatchInfoResponse struct {
	Response DepotPatchInfoPayload `json:"response"`
}

// DepotPatchInfoPayload contains patch availability and size metadata.
type DepotPatchInfoPayload struct {
	IsAvailable       bool   `json:"is_available"`
	PatchSize         uint64 `json:"patch_size"`
	PatchedChunksSize uint64 `json:"patched_chunks_size"`
}

// GetServersForSteamPipeOptions controls optional SteamPipe directory query parameters.
type GetServersForSteamPipeOptions struct {
	MaxServers         *uint32
	IPOverride         string
	LauncherType       *int32
	IPv6Public         string
	CurrentConnections string
}

// GetServersForSteamPipeResponse matches IContentServerDirectoryService/GetServersForSteamPipe/v1.
type GetServersForSteamPipeResponse struct {
	Response ServersForSteamPipePayload `json:"response"`
}

// ServersForSteamPipePayload contains SteamPipe content server candidates.
type ServersForSteamPipePayload struct {
	Servers []ContentServerInfo `json:"servers"`
}

// ContentServerInfo describes one content server candidate returned by Steam.
type ContentServerInfo struct {
	Type                     string   `json:"type"`
	SourceID                 int32    `json:"source_id"`
	CellID                   int32    `json:"cell_id"`
	Load                     int32    `json:"load"`
	WeightedLoad             float64  `json:"weighted_load"`
	NumEntriesInClientList   int32    `json:"num_entries_in_client_list"`
	SteamChinaOnly           bool     `json:"steam_china_only"`
	Host                     string   `json:"host"`
	VHost                    string   `json:"vhost"`
	UseAsProxy               bool     `json:"use_as_proxy"`
	ProxyRequestPathTemplate string   `json:"proxy_request_path_template"`
	HTTPSSupport             string   `json:"https_support"`
	AllowedAppIDs            []uint32 `json:"allowed_app_ids"`
	PreferredServer          bool     `json:"preferred_server"`
}
