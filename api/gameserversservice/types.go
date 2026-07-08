package gameserversservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// GetServerListResponse matches IGameServersService/GetServerList/v1.
type GetServerListResponse struct {
	Response GetServerListPayload `json:"response"`
}

// GetServerListPayload contains server candidates returned by Steam.
type GetServerListPayload struct {
	Servers []GameServer `json:"servers"`
}

// GameServer describes one server candidate returned by GetServerList.
//
// Values are discovery-time metadata from Steam's Web API side. Query the
// address with A2S when callers need live server state.
type GameServer struct {
	Addr       string `json:"addr"`
	GamePort   uint32 `json:"gameport"`
	SpecPort   uint32 `json:"specport"`
	SteamID    string `json:"steamid"`
	Name       string `json:"name"`
	AppID      uint32 `json:"appid"`
	GameDir    string `json:"gamedir"`
	Version    string `json:"version"`
	Product    string `json:"product"`
	Region     int    `json:"region"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"max_players"`
	Bots       int    `json:"bots"`
	Map        string `json:"map"`
	Secure     bool   `json:"secure"`
	Dedicated  bool   `json:"dedicated"`
	OS         string `json:"os"`
	GameType   string `json:"gametype"`
}

// UnmarshalJSON accepts common scalar fields encoded as either JSON strings or
// native JSON numbers/bools. GetServerList is undocumented, so this keeps
// typed decoding tolerant without broadening the public model to map[string]any.
func (s *GameServer) UnmarshalJSON(data []byte) error {
	var raw struct {
		Addr       flexibleString `json:"addr"`
		GamePort   flexibleUint32 `json:"gameport"`
		SpecPort   flexibleUint32 `json:"specport"`
		SteamID    flexibleString `json:"steamid"`
		Name       flexibleString `json:"name"`
		AppID      flexibleUint32 `json:"appid"`
		GameDir    flexibleString `json:"gamedir"`
		Version    flexibleString `json:"version"`
		Product    flexibleString `json:"product"`
		Region     flexibleInt    `json:"region"`
		Players    flexibleInt    `json:"players"`
		MaxPlayers flexibleInt    `json:"max_players"`
		Bots       flexibleInt    `json:"bots"`
		Map        flexibleString `json:"map"`
		Secure     flexibleBool   `json:"secure"`
		Dedicated  flexibleBool   `json:"dedicated"`
		OS         flexibleString `json:"os"`
		GameType   flexibleString `json:"gametype"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*s = GameServer{
		Addr:       string(raw.Addr),
		GamePort:   uint32(raw.GamePort),
		SpecPort:   uint32(raw.SpecPort),
		SteamID:    string(raw.SteamID),
		Name:       string(raw.Name),
		AppID:      uint32(raw.AppID),
		GameDir:    string(raw.GameDir),
		Version:    string(raw.Version),
		Product:    string(raw.Product),
		Region:     int(raw.Region),
		Players:    int(raw.Players),
		MaxPlayers: int(raw.MaxPlayers),
		Bots:       int(raw.Bots),
		Map:        string(raw.Map),
		Secure:     bool(raw.Secure),
		Dedicated:  bool(raw.Dedicated),
		OS:         string(raw.OS),
		GameType:   string(raw.GameType),
	}
	return nil
}

type flexibleString string

func (v *flexibleString) UnmarshalJSON(data []byte) error {
	value, err := jsonScalarString(data)
	if err != nil {
		return err
	}
	*v = flexibleString(value)
	return nil
}

type flexibleUint32 uint32

func (v *flexibleUint32) UnmarshalJSON(data []byte) error {
	value, err := jsonTrimmedScalarString(data)
	if err != nil {
		return err
	}
	if value == "" {
		*v = 0
		return nil
	}
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid uint32 value %q: %w", value, err)
	}
	*v = flexibleUint32(parsed)
	return nil
}

type flexibleInt int

func (v *flexibleInt) UnmarshalJSON(data []byte) error {
	value, err := jsonTrimmedScalarString(data)
	if err != nil {
		return err
	}
	if value == "" {
		*v = 0
		return nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid int value %q: %w", value, err)
	}
	*v = flexibleInt(parsed)
	return nil
}

type flexibleBool bool

func (v *flexibleBool) UnmarshalJSON(data []byte) error {
	value, err := jsonTrimmedScalarString(data)
	if err != nil {
		return err
	}
	switch strings.ToLower(value) {
	case "", "0", "false":
		*v = false
	case "1", "true":
		*v = true
	default:
		return fmt.Errorf("invalid bool value %q", value)
	}
	return nil
}

func jsonScalarString(data []byte) (string, error) {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) {
		return "", nil
	}
	if len(data) > 0 && data[0] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return "", err
		}
		return value, nil
	}
	return strings.TrimSpace(string(data)), nil
}

func jsonTrimmedScalarString(data []byte) (string, error) {
	value, err := jsonScalarString(data)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(value), nil
}
