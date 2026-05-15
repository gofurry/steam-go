package websession

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/gofurry/steam-go/api/authenticationservice"
	"github.com/gofurry/steam-go/internal/request"
)

const (
	testProtoWireVarint = 0
	testProtoWireBytes  = 2
)

type testProtoField struct {
	Number uint64
	Wire   uint64
	Value  []byte
}

type authResponseSpec struct {
	status int
	body   string
	err    error
}

type capturedAuthRequest struct {
	method      string
	path        string
	query       url.Values
	body        []byte
	contentType string
}

type sequenceAuthTransport struct {
	mu        sync.Mutex
	requests  []capturedAuthRequest
	responses []authResponseSpec
}

func (t *sequenceAuthTransport) Do(_ context.Context, req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	clonedQuery := make(url.Values, len(req.URL.Query()))
	for key, values := range req.URL.Query() {
		copied := make([]string, len(values))
		copy(copied, values)
		clonedQuery[key] = copied
	}

	var body []byte
	if req.Body != nil {
		readBody, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		body = readBody
	}

	t.requests = append(t.requests, capturedAuthRequest{
		method:      req.Method,
		path:        req.URL.Path,
		query:       clonedQuery,
		body:        body,
		contentType: req.Header.Get("Content-Type"),
	})

	response := authResponseSpec{
		status: http.StatusOK,
		body:   `{"response":{}}`,
	}
	if len(t.responses) > 0 {
		response = t.responses[0]
		t.responses = t.responses[1:]
	}
	if response.status == 0 {
		response.status = http.StatusOK
	}
	if response.err != nil {
		return nil, response.err
	}
	if strings.TrimSpace(response.body) == "" {
		response.body = `{"response":{}}`
	}

	return &http.Response{
		StatusCode: response.status,
		Body:       io.NopCloser(strings.NewReader(response.body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func (t *sequenceAuthTransport) snapshotRequests() []capturedAuthRequest {
	t.mu.Lock()
	defer t.mu.Unlock()

	out := make([]capturedAuthRequest, len(t.requests))
	copy(out, t.requests)
	return out
}

func newTestAuthService(t *testing.T, transport *sequenceAuthTransport) *authenticationservice.Service {
	t.Helper()

	executor, err := request.NewExecutor(
		"https://api.steampowered.com",
		nil,
		nil,
		1<<20,
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
	return authenticationservice.NewService(executor)
}

func newTestClient(t *testing.T, auth *authenticationservice.Service, opts ...Option) *Client {
	t.Helper()

	client, err := NewClient(auth, opts...)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	return client
}

func decodeInputProtoFieldsFromBody(t *testing.T, body []byte) []testProtoField {
	t.Helper()

	values, err := url.ParseQuery(string(body))
	if err != nil {
		t.Fatalf("ParseQuery returned error: %v", err)
	}
	encoded := values.Get("input_protobuf_encoded")
	if encoded == "" {
		t.Fatal("expected input_protobuf_encoded field")
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("DecodeString returned error: %v", err)
	}
	fields, err := readTestProtoFields(decoded)
	if err != nil {
		t.Fatalf("readTestProtoFields returned error: %v", err)
	}
	return fields
}

func readTestProtoFields(message []byte) ([]testProtoField, error) {
	fields := make([]testProtoField, 0)
	for offset := 0; offset < len(message); {
		key, next, err := readTestProtoVarint(message, offset)
		if err != nil {
			return nil, err
		}
		offset = next
		field := testProtoField{
			Number: key >> 3,
			Wire:   key & 0x7,
		}

		switch field.Wire {
		case testProtoWireVarint:
			value, next, err := readTestProtoVarint(message, offset)
			if err != nil {
				return nil, err
			}
			field.Value = appendTestProtoVarint(nil, value)
			offset = next
		case testProtoWireBytes:
			size, next, err := readTestProtoVarint(message, offset)
			if err != nil {
				return nil, err
			}
			offset = next
			if size > uint64(len(message)-offset) {
				return nil, fmt.Errorf("truncated bytes field %d", field.Number)
			}
			field.Value = append([]byte(nil), message[offset:offset+int(size)]...)
			offset += int(size)
		default:
			return nil, fmt.Errorf("unsupported wire type %d", field.Wire)
		}

		fields = append(fields, field)
	}
	return fields, nil
}

func readTestProtoVarint(data []byte, offset int) (uint64, int, error) {
	var value uint64
	for shift := uint(0); shift < 64; shift += 7 {
		if offset >= len(data) {
			return 0, offset, fmt.Errorf("truncated varint")
		}
		b := data[offset]
		offset++
		value |= uint64(b&0x7f) << shift
		if b < 0x80 {
			return value, offset, nil
		}
	}
	return 0, offset, fmt.Errorf("varint overflows uint64")
}

func appendTestProtoVarint(dst []byte, value uint64) []byte {
	for value >= 0x80 {
		dst = append(dst, byte(value)|0x80)
		value >>= 7
	}
	return append(dst, byte(value))
}

func testProtoFieldByNumber(t *testing.T, fields []testProtoField, number uint64) testProtoField {
	t.Helper()

	for _, field := range fields {
		if field.Number == number {
			return field
		}
	}
	t.Fatalf("expected proto field %d in %#v", number, fields)
	return testProtoField{}
}

func testJWTWithSubject(t *testing.T, subject string) string {
	t.Helper()

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"` + subject + `"}`))
	return header + "." + payload + ".sig"
}
