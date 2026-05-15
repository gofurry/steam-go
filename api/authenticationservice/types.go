package authenticationservice

import (
	"encoding/json"
	"strconv"
	"strings"
)

// GetPasswordRSAPublicKeyResponse contains the RSA public key metadata used by Steam credential auth.
type GetPasswordRSAPublicKeyResponse struct {
	PublicKeyMod string `json:"publickey_mod"`
	PublicKeyExp string `json:"publickey_exp"`
	Timestamp    uint64 `json:"timestamp"`
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
