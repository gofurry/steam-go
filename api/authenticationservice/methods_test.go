package authenticationservice

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	sdkerrors "github.com/gofurry/steam-go/internal/errors"
	"github.com/gofurry/steam-go/internal/request"
)

func TestGetPasswordRSAPublicKeyBuildsProtoQueryAndDecodesResponse(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{
		statuses: []int{http.StatusOK},
		responseBody: `{
			"response": {
				"publickey_mod": "c0ffee",
				"publickey_exp": "010001",
				"timestamp": "1710000000"
			}
		}`,
	}
	service := newTestService(t, transport)

	resp, err := service.GetPasswordRSAPublicKey(context.Background(), " account_name ")
	if err != nil {
		t.Fatalf("GetPasswordRSAPublicKey returned error: %v", err)
	}
	if resp.PublicKeyMod != "c0ffee" || resp.PublicKeyExp != "010001" || resp.Timestamp != 1710000000 {
		t.Fatalf("unexpected response: %#v", resp)
	}

	req := transport.onlyRequest(t)
	if req.method != http.MethodGet {
		t.Fatalf("unexpected method: %s", req.method)
	}
	if req.path != "/IAuthenticationService/GetPasswordRSAPublicKey/v1/" {
		t.Fatalf("unexpected path: %s", req.path)
	}

	fields := decodeInputProtoFields(t, req.query)
	if len(fields) != 1 {
		t.Fatalf("unexpected field count: %d", len(fields))
	}
	if fields[0].Number != 1 || fields[0].Wire != protoWireBytes || string(fields[0].Value) != "account_name" {
		t.Fatalf("unexpected account field: %#v", fields[0])
	}
}

func TestGetPasswordRSAPublicKeyValidationAndEResult(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}})
	if _, err := service.GetPasswordRSAPublicKey(context.Background(), " "); err == nil {
		t.Fatal("expected account name validation error")
	}

	erTransport := &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: `{"response":{"eresult":84,"message":"too many attempts"}}`,
	}
	service = newTestService(t, erTransport)
	_, err := service.GetPasswordRSAPublicKey(context.Background(), "test")
	var apiErr *sdkerrors.APIError
	if err == nil || !errors.As(err, &apiErr) || apiErr.Kind != sdkerrors.KindAPIResponse {
		t.Fatalf("expected api response error, got %v", err)
	}
	var resultErr *EResultError
	if !errors.As(err, &resultErr) {
		t.Fatalf("expected EResultError, got %T", err)
	}
	if resultErr.Code != 84 || resultErr.Name != "RateLimitExceeded" || resultErr.Message != "too many attempts" {
		t.Fatalf("unexpected EResultError: %#v", resultErr)
	}
}

func TestBeginAuthSessionViaCredentialsBuildsRequestAndDecodesResponse(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{
		statuses: []int{http.StatusOK},
		responseBody: `{
			"response": {
				"client_id": "12345",
				"request_id": "AQID",
				"interval": 2,
				"allowed_confirmations": [{"confirmation_type": 3, "associated_message": "mobile"}],
				"steamid": 76561198370695025,
				"weak_token": "weak"
			}
		}`,
	}
	service := newTestService(t, transport)

	resp, err := service.BeginAuthSessionViaCredentials(context.Background(), BeginAuthSessionViaCredentialsRequest{
		DeviceFriendlyName:  " test-device ",
		AccountName:         " account ",
		EncryptedPassword:   " encrypted ",
		EncryptionTimestamp: 1710000000,
		RememberLogin:       true,
		Language:            6,
	})
	if err != nil {
		t.Fatalf("BeginAuthSessionViaCredentials returned error: %v", err)
	}
	if resp.ClientID != 12345 || string(resp.RequestID) != "\x01\x02\x03" || resp.Interval != 2 {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if resp.SteamID != "76561198370695025" || resp.WeakToken != "weak" {
		t.Fatalf("unexpected identity fields: %#v", resp)
	}
	if len(resp.AllowedConfirmations) != 1 || resp.AllowedConfirmations[0].ConfirmationType != 3 {
		t.Fatalf("unexpected confirmations: %#v", resp.AllowedConfirmations)
	}

	req := transport.onlyRequest(t)
	if req.method != http.MethodPost {
		t.Fatalf("unexpected method: %s", req.method)
	}
	if req.path != "/IAuthenticationService/BeginAuthSessionViaCredentials/v1/" {
		t.Fatalf("unexpected path: %s", req.path)
	}
	if req.contentType != "application/x-www-form-urlencoded" {
		t.Fatalf("unexpected content type: %s", req.contentType)
	}

	fields := decodeInputProtoFieldsFromBody(t, req.body)
	assertProtoString(t, fields, 1, "test-device")
	assertProtoString(t, fields, 2, "account")
	assertProtoString(t, fields, 3, "encrypted")
	assertProtoVarint(t, fields, 4, 1710000000)
	assertProtoVarint(t, fields, 5, 1)
	assertProtoVarint(t, fields, 6, 2)
	assertProtoVarint(t, fields, 7, 1)
	assertProtoString(t, fields, 8, "Store")
	assertProtoVarint(t, fields, 11, 6)

	deviceFields := decodeNestedField(t, fields, 9)
	assertProtoString(t, deviceFields, 1, "test-device")
	assertProtoVarint(t, deviceFields, 2, 2)
}

func TestBeginAuthSessionViaQRBuildsRequestAndDecodesResponse(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{
		statuses: []int{http.StatusOK},
		responseBody: `{
			"response": {
				"client_id": 99,
				"challenge_url": "https://s.team/q/abc",
				"request_id": "BAU=",
				"interval": 5,
				"allowed_confirmations": [{"confirmation_type": 4, "associated_message": "approve"}],
				"version": 1
			}
		}`,
	}
	service := newTestService(t, transport)

	resp, err := service.BeginAuthSessionViaQR(context.Background(), "")
	if err != nil {
		t.Fatalf("BeginAuthSessionViaQR returned error: %v", err)
	}
	if resp.ClientID != 99 || resp.ChallengeURL != "https://s.team/q/abc" || string(resp.RequestID) != "\x04\x05" {
		t.Fatalf("unexpected response: %#v", resp)
	}

	req := transport.onlyRequest(t)
	if req.path != "/IAuthenticationService/BeginAuthSessionViaQR/v1/" {
		t.Fatalf("unexpected path: %s", req.path)
	}
	fields := decodeInputProtoFieldsFromBody(t, req.body)
	deviceFields := decodeNestedField(t, fields, 3)
	assertProtoString(t, deviceFields, 1, defaultDeviceFriendlyName)
	assertProtoVarint(t, deviceFields, 2, 2)
}

func TestUpdateAuthSessionWithSteamGuardCodeBuildsRequest(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{
		statuses:     []int{http.StatusOK},
		responseBody: `{"response":{}}`,
	}
	service := newTestService(t, transport)

	_, err := service.UpdateAuthSessionWithSteamGuardCode(context.Background(), UpdateAuthSessionWithSteamGuardCodeRequest{
		ClientID: 42,
		SteamID:  "76561198370695025",
		Code:     " ABC12 ",
		CodeType: GuardCodeTypeDeviceCode,
	})
	if err != nil {
		t.Fatalf("UpdateAuthSessionWithSteamGuardCode returned error: %v", err)
	}

	req := transport.onlyRequest(t)
	if req.path != "/IAuthenticationService/UpdateAuthSessionWithSteamGuardCode/v1/" {
		t.Fatalf("unexpected path: %s", req.path)
	}
	fields := decodeInputProtoFieldsFromBody(t, req.body)
	assertProtoVarint(t, fields, 1, 42)
	assertProtoFixed64(t, fields, 2, 76561198370695025)
	assertProtoString(t, fields, 3, "ABC12")
	assertProtoVarint(t, fields, 4, 3)
}

func TestPollAuthSessionStatusBuildsRequestAndDecodesResponse(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{
		statuses: []int{http.StatusOK},
		responseBody: `{
			"response": {
				"new_client_id": "43",
				"new_challenge_url": "https://s.team/new",
				"refresh_token": "refresh",
				"access_token": "access",
				"had_remote_interaction": true,
				"account_name": "account",
				"new_guard_data": "guard"
			}
		}`,
	}
	service := newTestService(t, transport)

	resp, err := service.PollAuthSessionStatus(context.Background(), PollAuthSessionStatusRequest{
		ClientID:  42,
		RequestID: []byte{1, 2, 3},
	})
	if err != nil {
		t.Fatalf("PollAuthSessionStatus returned error: %v", err)
	}
	if resp.NewClientID != 43 || resp.RefreshToken != "refresh" || resp.AccessToken != "access" || !resp.HadRemoteInteraction {
		t.Fatalf("unexpected response: %#v", resp)
	}

	req := transport.onlyRequest(t)
	if req.path != "/IAuthenticationService/PollAuthSessionStatus/v1/" {
		t.Fatalf("unexpected path: %s", req.path)
	}
	fields := decodeInputProtoFieldsFromBody(t, req.body)
	assertProtoVarint(t, fields, 1, 42)
	assertProtoBytes(t, fields, 2, []byte{1, 2, 3})
}

func TestAuthSessionRequestValidation(t *testing.T) {
	t.Parallel()

	service := newTestService(t, &recordingTransport{statuses: []int{http.StatusOK}})
	if _, err := service.BeginAuthSessionViaCredentials(context.Background(), BeginAuthSessionViaCredentialsRequest{}); err == nil {
		t.Fatal("expected credentials validation error")
	}
	if _, err := service.UpdateAuthSessionWithSteamGuardCode(context.Background(), UpdateAuthSessionWithSteamGuardCodeRequest{}); err == nil {
		t.Fatal("expected guard update validation error")
	}
	if _, err := service.PollAuthSessionStatus(context.Background(), PollAuthSessionStatusRequest{}); err == nil {
		t.Fatal("expected poll validation error")
	}
}

func TestProtoHelpersBuildFormAndDecodeFields(t *testing.T) {
	t.Parallel()

	message := appendProtoString(nil, 1, "alice")
	message = appendProtoUint64(message, 2, 42)
	encoded := encodeProtoBase64(message)

	form := string(buildProtoForm(encoded))
	values, err := url.ParseQuery(form)
	if err != nil {
		t.Fatalf("ParseQuery returned error: %v", err)
	}
	if got := values.Get("input_protobuf_encoded"); got != encoded {
		t.Fatalf("unexpected form value: %q", got)
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("DecodeString returned error: %v", err)
	}
	fields, err := readProtoFields(decoded)
	if err != nil {
		t.Fatalf("readProtoFields returned error: %v", err)
	}
	if len(fields) != 2 {
		t.Fatalf("unexpected field count: %d", len(fields))
	}
	if fields[0].Number != 1 || string(fields[0].Value) != "alice" {
		t.Fatalf("unexpected string field: %#v", fields[0])
	}
	value, _, err := readProtoVarint(fields[1].Value, 0)
	if err != nil {
		t.Fatalf("readProtoVarint returned error: %v", err)
	}
	if fields[1].Number != 2 || value != 42 {
		t.Fatalf("unexpected uint64 field: %#v value=%d", fields[1], value)
	}
}

func TestEncryptPasswordPKCS1v15(t *testing.T) {
	t.Parallel()

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("GenerateKey returned error: %v", err)
	}
	encrypted, err := EncryptPasswordPKCS1v15("secret-password", key.N.Text(16), big.NewInt(int64(key.E)).Text(16))
	if err != nil {
		t.Fatalf("EncryptPasswordPKCS1v15 returned error: %v", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatalf("DecodeString returned error: %v", err)
	}
	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, key, ciphertext)
	if err != nil {
		t.Fatalf("DecryptPKCS1v15 returned error: %v", err)
	}
	if got := string(plaintext); got != "secret-password" {
		t.Fatalf("unexpected plaintext: %q", got)
	}
}

func newTestService(t *testing.T, transport *recordingTransport) *Service {
	t.Helper()

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		1024,
		request.ExecutionPolicy{
			Retry:        0,
			RetryBackoff: request.DefaultRetryBackoffConfig(),
			Transport:    transport,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewExecutor returned error: %v", err)
	}
	return NewService(executor)
}

func decodeInputProtoFields(t *testing.T, query url.Values) []protoField {
	t.Helper()

	encoded := query.Get("input_protobuf_encoded")
	if encoded == "" {
		t.Fatal("expected input_protobuf_encoded")
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("DecodeString returned error: %v", err)
	}
	fields, err := readProtoFields(decoded)
	if err != nil {
		t.Fatalf("readProtoFields returned error: %v", err)
	}
	return fields
}

func decodeInputProtoFieldsFromBody(t *testing.T, body []byte) []protoField {
	t.Helper()

	values, err := url.ParseQuery(string(body))
	if err != nil {
		t.Fatalf("ParseQuery returned error: %v", err)
	}
	return decodeInputProtoFields(t, values)
}

func decodeNestedField(t *testing.T, fields []protoField, number uint64) []protoField {
	t.Helper()

	field := protoFieldByNumber(t, fields, number)
	nested, err := readProtoFields(field.Value)
	if err != nil {
		t.Fatalf("readProtoFields nested returned error: %v", err)
	}
	return nested
}

func assertProtoString(t *testing.T, fields []protoField, number uint64, want string) {
	t.Helper()

	field := protoFieldByNumber(t, fields, number)
	if field.Wire != protoWireBytes || string(field.Value) != want {
		t.Fatalf("unexpected string field %d: %#v want %q", number, field, want)
	}
}

func assertProtoBytes(t *testing.T, fields []protoField, number uint64, want []byte) {
	t.Helper()

	field := protoFieldByNumber(t, fields, number)
	if field.Wire != protoWireBytes || string(field.Value) != string(want) {
		t.Fatalf("unexpected bytes field %d: %#v want %v", number, field, want)
	}
}

func assertProtoVarint(t *testing.T, fields []protoField, number uint64, want uint64) {
	t.Helper()

	field := protoFieldByNumber(t, fields, number)
	got, _, err := readProtoVarint(field.Value, 0)
	if err != nil {
		t.Fatalf("readProtoVarint returned error: %v", err)
	}
	if field.Wire != protoWireVarint || got != want {
		t.Fatalf("unexpected varint field %d: %#v got %d want %d", number, field, got, want)
	}
}

func assertProtoFixed64(t *testing.T, fields []protoField, number uint64, want uint64) {
	t.Helper()

	field := protoFieldByNumber(t, fields, number)
	if field.Wire != protoWireFixed64 || len(field.Value) != 8 {
		t.Fatalf("unexpected fixed64 field %d: %#v", number, field)
	}
	got := binary.LittleEndian.Uint64(field.Value)
	if got != want {
		t.Fatalf("unexpected fixed64 field %d: got %d want %d", number, got, want)
	}
}

func protoFieldByNumber(t *testing.T, fields []protoField, number uint64) protoField {
	t.Helper()

	for _, field := range fields {
		if field.Number == number {
			return field
		}
	}
	t.Fatalf("expected proto field %d in %#v", number, fields)
	return protoField{}
}

type recordingTransport struct {
	mu           sync.Mutex
	requests     []capturedRequest
	statuses     []int
	responseBody string
}

type capturedRequest struct {
	method      string
	path        string
	query       url.Values
	body        []byte
	contentType string
}

func (t *recordingTransport) Do(_ context.Context, req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	clonedQuery := make(url.Values, len(req.URL.Query()))
	for key, values := range req.URL.Query() {
		copied := make([]string, len(values))
		copy(copied, values)
		clonedQuery[key] = copied
	}
	var requestBody []byte
	if req.Body != nil {
		var err error
		requestBody, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}
	t.requests = append(t.requests, capturedRequest{
		method:      req.Method,
		path:        req.URL.Path,
		query:       clonedQuery,
		body:        requestBody,
		contentType: req.Header.Get("Content-Type"),
	})

	status := http.StatusOK
	if len(t.statuses) > 0 {
		status = t.statuses[0]
		t.statuses = t.statuses[1:]
	}
	responseBody := t.responseBody
	if strings.TrimSpace(responseBody) == "" {
		responseBody = `{"response":{}}`
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func (t *recordingTransport) onlyRequest(tb testing.TB) capturedRequest {
	tb.Helper()
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.requests) != 1 {
		tb.Fatalf("expected one request, got %d", len(t.requests))
	}
	return t.requests[0]
}
