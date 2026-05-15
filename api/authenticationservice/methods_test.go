package authenticationservice

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
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

type recordingTransport struct {
	mu           sync.Mutex
	requests     []capturedRequest
	statuses     []int
	responseBody string
}

type capturedRequest struct {
	method string
	path   string
	query  url.Values
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
	t.requests = append(t.requests, capturedRequest{
		method: req.Method,
		path:   req.URL.Path,
		query:  clonedQuery,
	})

	status := http.StatusOK
	if len(t.statuses) > 0 {
		status = t.statuses[0]
		t.statuses = t.statuses[1:]
	}
	body := t.responseBody
	if strings.TrimSpace(body) == "" {
		body = `{"response":{}}`
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
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
