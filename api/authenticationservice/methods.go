package authenticationservice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofurry/steam-go/internal/endpoint"
	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/request"
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
