package authenticationservice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofurry/steam-go/internal/endpoint"
	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/request"
)

const (
	defaultDeviceFriendlyName = "steam-go"
	defaultWebsiteID          = "Store"
	formContentType           = "application/x-www-form-urlencoded"
)

// GetPasswordRSAPublicKey returns the RSA public key metadata for one account name.
func (s *Service) GetPasswordRSAPublicKey(ctx context.Context, accountName string) (GetPasswordRSAPublicKeyResponse, error) {
	body, err := s.GetPasswordRSAPublicKeyRaw(ctx, accountName)
	if err != nil {
		return GetPasswordRSAPublicKeyResponse{}, err
	}
	return decodeAuthenticationResponse[GetPasswordRSAPublicKeyResponse](body)
}

// GetPasswordRSAPublicKeyRaw returns the raw JSON response body for the RSA public key lookup.
func (s *Service) GetPasswordRSAPublicKeyRaw(ctx context.Context, accountName string) ([]byte, error) {
	accountName = strings.TrimSpace(accountName)
	if accountName == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "account name must not be empty", nil, nil)
	}

	query := url.Values{}
	query.Set("input_protobuf_encoded", encodeProtoBase64(appendProtoString(nil, 1, accountName)))

	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method: http.MethodGet,
		Path:   endpoint.AuthenticationServiceGetPasswordRSAPublicKey,
		Query:  query,
	})
}

// BeginAuthSessionViaCredentials starts an auth session using an already-encrypted password.
func (s *Service) BeginAuthSessionViaCredentials(ctx context.Context, req BeginAuthSessionViaCredentialsRequest) (BeginAuthSessionResponse, error) {
	body, err := s.BeginAuthSessionViaCredentialsRaw(ctx, req)
	if err != nil {
		return BeginAuthSessionResponse{}, err
	}
	return decodeAuthenticationResponse[BeginAuthSessionResponse](body)
}

// BeginAuthSessionViaCredentialsRaw returns the raw JSON response body for credential auth.
func (s *Service) BeginAuthSessionViaCredentialsRaw(ctx context.Context, req BeginAuthSessionViaCredentialsRequest) ([]byte, error) {
	req = normalizeCredentialsRequest(req)
	if err := validateCredentialsRequest(req); err != nil {
		return nil, err
	}

	deviceDetails := appendProtoString(nil, 1, req.DeviceFriendlyName)
	deviceDetails = appendProtoUint64(deviceDetails, 2, uint64(req.PlatformType))

	message := appendProtoString(nil, 1, req.DeviceFriendlyName)
	message = appendProtoString(message, 2, req.AccountName)
	message = appendProtoString(message, 3, req.EncryptedPassword)
	message = appendProtoUint64(message, 4, req.EncryptionTimestamp)
	message = appendProtoBool(message, 5, req.RememberLogin)
	message = appendProtoUint64(message, 6, uint64(req.PlatformType))
	message = appendProtoUint64(message, 7, uint64(req.Persistence))
	message = appendProtoString(message, 8, req.WebsiteID)
	message = appendProtoMessage(message, 9, deviceDetails)
	if req.Language != 0 {
		message = appendProtoUint64(message, 11, req.Language)
	}
	return s.doProtoPost(ctx, endpoint.AuthenticationServiceBeginAuthSessionViaCredentials, message)
}

// BeginAuthSessionViaQR starts an auth session using Steam's QR login challenge.
func (s *Service) BeginAuthSessionViaQR(ctx context.Context, deviceName string) (BeginQRAuthSessionResponse, error) {
	body, err := s.BeginAuthSessionViaQRRaw(ctx, deviceName)
	if err != nil {
		return BeginQRAuthSessionResponse{}, err
	}
	return decodeAuthenticationResponse[BeginQRAuthSessionResponse](body)
}

// BeginAuthSessionViaQRRaw returns the raw JSON response body for QR auth.
func (s *Service) BeginAuthSessionViaQRRaw(ctx context.Context, deviceName string) ([]byte, error) {
	deviceName = strings.TrimSpace(deviceName)
	if deviceName == "" {
		deviceName = defaultDeviceFriendlyName
	}

	deviceDetails := appendProtoString(nil, 1, deviceName)
	deviceDetails = appendProtoUint64(deviceDetails, 2, uint64(AuthSessionPlatformTypeWebBrowser))
	message := appendProtoMessage(nil, 3, deviceDetails)
	return s.doProtoPost(ctx, endpoint.AuthenticationServiceBeginAuthSessionViaQR, message)
}

// UpdateAuthSessionWithSteamGuardCode submits one Steam Guard code or confirmation.
func (s *Service) UpdateAuthSessionWithSteamGuardCode(ctx context.Context, req UpdateAuthSessionWithSteamGuardCodeRequest) (UpdateAuthSessionWithSteamGuardCodeResponse, error) {
	body, err := s.UpdateAuthSessionWithSteamGuardCodeRaw(ctx, req)
	if err != nil {
		return UpdateAuthSessionWithSteamGuardCodeResponse{}, err
	}
	return decodeAuthenticationResponse[UpdateAuthSessionWithSteamGuardCodeResponse](body)
}

// UpdateAuthSessionWithSteamGuardCodeRaw returns the raw JSON response body for Steam Guard updates.
func (s *Service) UpdateAuthSessionWithSteamGuardCodeRaw(ctx context.Context, req UpdateAuthSessionWithSteamGuardCodeRequest) ([]byte, error) {
	steamID, err := validateSteamID(req.SteamID)
	if err != nil {
		return nil, err
	}
	if req.ClientID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "client id must be greater than zero", nil, nil)
	}
	req.Code = strings.TrimSpace(req.Code)
	if req.Code == "" {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "guard code must not be empty", nil, nil)
	}
	if req.CodeType == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "guard code type must be greater than zero", nil, nil)
	}

	message := appendProtoUint64(nil, 1, req.ClientID)
	message = appendProtoFixed64(message, 2, steamID)
	message = appendProtoString(message, 3, req.Code)
	message = appendProtoUint64(message, 4, uint64(req.CodeType))
	return s.doProtoPost(ctx, endpoint.AuthenticationServiceUpdateAuthSessionWithSteamGuardCode, message)
}

// PollAuthSessionStatus polls one auth session for tokens and account state.
func (s *Service) PollAuthSessionStatus(ctx context.Context, req PollAuthSessionStatusRequest) (PollAuthSessionStatusResponse, error) {
	body, err := s.PollAuthSessionStatusRaw(ctx, req)
	if err != nil {
		return PollAuthSessionStatusResponse{}, err
	}
	return decodeAuthenticationResponse[PollAuthSessionStatusResponse](body)
}

// PollAuthSessionStatusRaw returns the raw JSON response body for auth session polling.
func (s *Service) PollAuthSessionStatusRaw(ctx context.Context, req PollAuthSessionStatusRequest) ([]byte, error) {
	if req.ClientID == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "client id must be greater than zero", nil, nil)
	}
	if len(req.RequestID) == 0 {
		return nil, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "request id must not be empty", nil, nil)
	}

	message := appendProtoUint64(nil, 1, req.ClientID)
	message = appendProtoBytes(message, 2, req.RequestID)
	return s.doProtoPost(ctx, endpoint.AuthenticationServicePollAuthSessionStatus, message)
}

func (s *Service) doProtoPost(ctx context.Context, path string, message []byte) ([]byte, error) {
	return s.executor.DoRaw(ctx, request.RequestSpec{
		Method:      http.MethodPost,
		Path:        path,
		Body:        buildProtoForm(encodeProtoBase64(message)),
		ContentType: formContentType,
	})
}

func normalizeCredentialsRequest(req BeginAuthSessionViaCredentialsRequest) BeginAuthSessionViaCredentialsRequest {
	req.DeviceFriendlyName = strings.TrimSpace(req.DeviceFriendlyName)
	if req.DeviceFriendlyName == "" {
		req.DeviceFriendlyName = defaultDeviceFriendlyName
	}
	req.AccountName = strings.TrimSpace(req.AccountName)
	req.EncryptedPassword = strings.TrimSpace(req.EncryptedPassword)
	if req.PlatformType == 0 {
		req.PlatformType = AuthSessionPlatformTypeWebBrowser
	}
	if req.Persistence == 0 {
		req.Persistence = AuthSessionPersistencePersistent
	}
	req.WebsiteID = strings.TrimSpace(req.WebsiteID)
	if req.WebsiteID == "" {
		req.WebsiteID = defaultWebsiteID
	}
	return req
}

func validateCredentialsRequest(req BeginAuthSessionViaCredentialsRequest) error {
	if req.AccountName == "" {
		return sdkerrors.New(sdkerrors.KindRequestBuild, 0, "account name must not be empty", nil, nil)
	}
	if req.EncryptedPassword == "" {
		return sdkerrors.New(sdkerrors.KindRequestBuild, 0, "encrypted password must not be empty", nil, nil)
	}
	if req.EncryptionTimestamp == 0 {
		return sdkerrors.New(sdkerrors.KindRequestBuild, 0, "encryption timestamp must be greater than zero", nil, nil)
	}
	return nil
}

func validateSteamID(steamID string) (uint64, error) {
	steamID = strings.TrimSpace(steamID)
	if steamID == "" {
		return 0, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam id must not be empty", nil, nil)
	}
	value, err := strconv.ParseUint(steamID, 10, 64)
	if err != nil {
		return 0, sdkerrors.New(sdkerrors.KindRequestBuild, 0, "steam id must be a uint64 string", nil, err)
	}
	return value, nil
}

type authenticationEnvelope[T any] struct {
	Response     T               `json:"response"`
	EResult      int             `json:"eresult"`
	Error        string          `json:"error"`
	ErrorMessage string          `json:"error_message"`
	Message      string          `json:"message"`
	RawResponse  json.RawMessage `json:"-"`
}

func decodeAuthenticationResponse[T any](body []byte) (T, error) {
	var out T
	var envelope authenticationEnvelope[T]
	if err := json.Unmarshal(body, &envelope); err != nil {
		return out, sdkerrors.New(sdkerrors.KindDecode, 0, "decode response body failed", body, err)
	}

	if err := authenticationEResult(body); err != nil {
		return out, err
	}
	return envelope.Response, nil
}

func authenticationEResult(body []byte) error {
	var raw struct {
		EResult      int             `json:"eresult"`
		Error        string          `json:"error"`
		ErrorMessage string          `json:"error_message"`
		Message      string          `json:"message"`
		Response     json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil
	}

	code := raw.EResult
	message := firstNonEmpty(raw.ErrorMessage, raw.Error, raw.Message)
	if len(raw.Response) > 0 {
		var response struct {
			EResult      int    `json:"eresult"`
			Error        string `json:"error"`
			ErrorMessage string `json:"error_message"`
			Message      string `json:"message"`
		}
		if err := json.Unmarshal(raw.Response, &response); err == nil {
			if response.EResult != 0 {
				code = response.EResult
			}
			message = firstNonEmpty(message, response.ErrorMessage, response.Error, response.Message)
		}
	}
	if code == 0 || code == 1 {
		return nil
	}

	resultErr := &EResultError{
		Code:    code,
		Name:    erResultName(code),
		Message: message,
	}
	return sdkerrors.New(sdkerrors.KindAPIResponse, 0, fmt.Sprintf("steam authentication failed with EResult %d", code), body, resultErr)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
