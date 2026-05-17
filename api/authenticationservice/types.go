package authenticationservice

import (
	"encoding/json"
	"strconv"
	"strings"
)

// AuthSessionPlatformType identifies the Steam auth client platform.
type AuthSessionPlatformType uint64

const (
	// AuthSessionPlatformTypeWebBrowser matches Steam's web browser auth platform.
	AuthSessionPlatformTypeWebBrowser AuthSessionPlatformType = 2
)

// AuthSessionPersistence controls Steam auth session persistence.
type AuthSessionPersistence uint64

const (
	// AuthSessionPersistencePersistent asks Steam for a persistent auth session.
	AuthSessionPersistencePersistent AuthSessionPersistence = 1
)

// GuardCodeType identifies the Steam Guard confirmation code type.
type GuardCodeType uint64

const (
	GuardCodeTypeEmailCode          GuardCodeType = 2
	GuardCodeTypeDeviceCode         GuardCodeType = 3
	GuardCodeTypeDeviceConfirmation GuardCodeType = 4
	GuardCodeTypeEmailConfirmation  GuardCodeType = 5
)

// GetPasswordRSAPublicKeyResponse contains the RSA public key metadata used by Steam credential auth.
type GetPasswordRSAPublicKeyResponse struct {
	PublicKeyMod string `json:"publickey_mod"`
	PublicKeyExp string `json:"publickey_exp"`
	Timestamp    uint64 `json:"timestamp"`
}

// BeginAuthSessionViaCredentialsRequest contains the already-encrypted credential auth payload.
type BeginAuthSessionViaCredentialsRequest struct {
	DeviceFriendlyName  string
	AccountName         string
	EncryptedPassword   string
	EncryptionTimestamp uint64
	RememberLogin       bool
	PlatformType        AuthSessionPlatformType
	Persistence         AuthSessionPersistence
	WebsiteID           string
	Language            uint64
}

// AuthConfirmation describes one Steam Guard confirmation option.
type AuthConfirmation struct {
	ConfirmationType  uint32 `json:"confirmation_type"`
	AssociatedMessage string `json:"associated_message"`
}

// BeginAuthSessionResponse contains the shared response shape for credential auth sessions.
type BeginAuthSessionResponse struct {
	ClientID             uint64             `json:"client_id"`
	RequestID            []byte             `json:"request_id"`
	Interval             uint32             `json:"interval"`
	AllowedConfirmations []AuthConfirmation `json:"allowed_confirmations"`
	SteamID              string             `json:"steamid"`
	WeakToken            string             `json:"weak_token"`
}

// BeginQRAuthSessionResponse contains the QR auth challenge metadata.
type BeginQRAuthSessionResponse struct {
	ClientID             uint64             `json:"client_id"`
	ChallengeURL         string             `json:"challenge_url"`
	RequestID            []byte             `json:"request_id"`
	Interval             uint32             `json:"interval"`
	AllowedConfirmations []AuthConfirmation `json:"allowed_confirmations"`
	Version              uint32             `json:"version"`
}

// UpdateAuthSessionWithSteamGuardCodeRequest submits one Steam Guard code or confirmation.
type UpdateAuthSessionWithSteamGuardCodeRequest struct {
	ClientID uint64
	SteamID  string
	Code     string
	CodeType GuardCodeType
}

// UpdateAuthSessionWithSteamGuardCodeResponse is reserved for Steam's update response payload.
type UpdateAuthSessionWithSteamGuardCodeResponse struct{}

// PollAuthSessionStatusRequest polls one auth session for tokens and account state.
type PollAuthSessionStatusRequest struct {
	ClientID  uint64
	RequestID []byte
}

// PollAuthSessionStatusResponse contains the auth session polling result.
type PollAuthSessionStatusResponse struct {
	NewClientID          uint64 `json:"new_client_id"`
	NewChallengeURL      string `json:"new_challenge_url"`
	RefreshToken         string `json:"refresh_token"`
	AccessToken          string `json:"access_token"`
	HadRemoteInteraction bool   `json:"had_remote_interaction"`
	AccountName          string `json:"account_name"`
	NewGuardData         string `json:"new_guard_data"`
}

// UnmarshalJSON accepts Steam timestamp values encoded as either JSON strings or numbers.
func (r *GetPasswordRSAPublicKeyResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		PublicKeyMod string          `json:"publickey_mod"`
		PublicKeyExp string          `json:"publickey_exp"`
		Timestamp    json.RawMessage `json:"timestamp"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.PublicKeyMod = raw.PublicKeyMod
	r.PublicKeyExp = raw.PublicKeyExp
	if len(raw.Timestamp) == 0 || string(raw.Timestamp) == "null" {
		r.Timestamp = 0
		return nil
	}

	var number uint64
	if err := json.Unmarshal(raw.Timestamp, &number); err == nil {
		r.Timestamp = number
		return nil
	}

	var text string
	if err := json.Unmarshal(raw.Timestamp, &text); err != nil {
		return err
	}
	value, err := strconv.ParseUint(strings.TrimSpace(text), 10, 64)
	if err != nil {
		return err
	}
	r.Timestamp = value
	return nil
}

func (r *BeginAuthSessionResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		ClientID             json.RawMessage    `json:"client_id"`
		RequestID            []byte             `json:"request_id"`
		Interval             uint32             `json:"interval"`
		AllowedConfirmations []AuthConfirmation `json:"allowed_confirmations"`
		SteamID              json.RawMessage    `json:"steamid"`
		WeakToken            string             `json:"weak_token"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	clientID, err := decodeFlexibleUint64(raw.ClientID)
	if err != nil {
		return err
	}
	r.ClientID = clientID
	r.RequestID = raw.RequestID
	r.Interval = raw.Interval
	r.AllowedConfirmations = raw.AllowedConfirmations
	r.SteamID = decodeFlexibleString(raw.SteamID)
	r.WeakToken = raw.WeakToken
	return nil
}

func (r *BeginQRAuthSessionResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		ClientID             json.RawMessage    `json:"client_id"`
		ChallengeURL         string             `json:"challenge_url"`
		RequestID            []byte             `json:"request_id"`
		Interval             uint32             `json:"interval"`
		AllowedConfirmations []AuthConfirmation `json:"allowed_confirmations"`
		Version              uint32             `json:"version"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	clientID, err := decodeFlexibleUint64(raw.ClientID)
	if err != nil {
		return err
	}
	r.ClientID = clientID
	r.ChallengeURL = raw.ChallengeURL
	r.RequestID = raw.RequestID
	r.Interval = raw.Interval
	r.AllowedConfirmations = raw.AllowedConfirmations
	r.Version = raw.Version
	return nil
}

func (r *PollAuthSessionStatusResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		NewClientID          json.RawMessage `json:"new_client_id"`
		NewChallengeURL      string          `json:"new_challenge_url"`
		RefreshToken         string          `json:"refresh_token"`
		AccessToken          string          `json:"access_token"`
		HadRemoteInteraction bool            `json:"had_remote_interaction"`
		AccountName          string          `json:"account_name"`
		NewGuardData         string          `json:"new_guard_data"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	newClientID, err := decodeFlexibleUint64(raw.NewClientID)
	if err != nil {
		return err
	}
	r.NewClientID = newClientID
	r.NewChallengeURL = raw.NewChallengeURL
	r.RefreshToken = raw.RefreshToken
	r.AccessToken = raw.AccessToken
	r.HadRemoteInteraction = raw.HadRemoteInteraction
	r.AccountName = raw.AccountName
	r.NewGuardData = raw.NewGuardData
	return nil
}

func decodeFlexibleUint64(raw json.RawMessage) (uint64, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, nil
	}
	var number uint64
	if err := json.Unmarshal(raw, &number); err == nil {
		return number, nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(text), 10, 64)
}

func decodeFlexibleString(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}
	var number uint64
	if err := json.Unmarshal(raw, &number); err == nil {
		return strconv.FormatUint(number, 10)
	}
	return ""
}
