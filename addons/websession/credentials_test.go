package websession

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/gofurry/steam-go/api/authenticationservice"
)

func TestStartWithCredentialsPreservesPasswordAndBuildsChallenge(t *testing.T) {
	t.Parallel()

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("GenerateKey returned error: %v", err)
	}

	transport := &sequenceAuthTransport{
		responses: []authResponseSpec{
			{
				body: `{"response":{"publickey_mod":"` + key.N.Text(16) + `","publickey_exp":"` + big.NewInt(int64(key.E)).Text(16) + `","timestamp":"1710000000"}}`,
			},
			{
				body: `{"response":{"client_id":"42","request_id":"AQID","interval":2,"allowed_confirmations":[{"confirmation_type":4,"associated_message":"approve in mobile app"}],"steamid":"76561198000000001","weak_token":"weak"}}`,
			},
		},
	}
	client := newTestClient(t, newTestAuthService(t, transport))

	challenge, err := client.StartWithCredentials(context.Background(), StartWithCredentialsRequest{
		AccountName: " account ",
		Password:    "  secret  ",
		DeviceName:  " steam-go test ",
	})
	if err != nil {
		t.Fatalf("StartWithCredentials returned error: %v", err)
	}
	if challenge.SteamID != "76561198000000001" {
		t.Fatalf("unexpected steam id: %s", challenge.SteamID)
	}
	if challenge.ClientID != 42 {
		t.Fatalf("unexpected client id: %d", challenge.ClientID)
	}
	if challenge.PollInterval != 2*time.Second {
		t.Fatalf("unexpected poll interval: %s", challenge.PollInterval)
	}
	if len(challenge.AllowedConfirmations) != 1 || challenge.AllowedConfirmations[0].ConfirmationType != 4 {
		t.Fatalf("unexpected confirmations: %#v", challenge.AllowedConfirmations)
	}

	requests := transport.snapshotRequests()
	if len(requests) != 2 {
		t.Fatalf("expected 2 auth requests, got %d", len(requests))
	}
	if requests[0].path != "/IAuthenticationService/GetPasswordRSAPublicKey/v1/" {
		t.Fatalf("unexpected rsa key path: %s", requests[0].path)
	}
	if requests[1].path != "/IAuthenticationService/BeginAuthSessionViaCredentials/v1/" {
		t.Fatalf("unexpected begin auth path: %s", requests[1].path)
	}

	fields := decodeInputProtoFieldsFromBody(t, requests[1].body)
	if got := string(testProtoFieldByNumber(t, fields, 1).Value); got != "steam-go test" {
		t.Fatalf("unexpected device name: %q", got)
	}
	if got := string(testProtoFieldByNumber(t, fields, 2).Value); got != "account" {
		t.Fatalf("unexpected account name: %q", got)
	}
	if got := string(testProtoFieldByNumber(t, fields, 8).Value); got != "Store" {
		t.Fatalf("unexpected website id: %q", got)
	}

	encrypted := string(testProtoFieldByNumber(t, fields, 3).Value)
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatalf("DecodeString returned error: %v", err)
	}
	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, key, ciphertext)
	if err != nil {
		t.Fatalf("DecryptPKCS1v15 returned error: %v", err)
	}
	if got := string(plaintext); got != "  secret  " {
		t.Fatalf("unexpected plaintext password: %q", got)
	}
}

func TestSubmitSteamGuardCodeIgnoresDuplicateRequest(t *testing.T) {
	t.Parallel()

	transport := &sequenceAuthTransport{
		responses: []authResponseSpec{
			{
				body: `{"response":{"eresult":29,"message":"duplicate"}}`,
			},
		},
	}
	client := newTestClient(t, newTestAuthService(t, transport))

	err := client.SubmitSteamGuardCode(context.Background(), &LoginChallenge{
		SteamID:  "76561198000000001",
		ClientID: 42,
	}, "123456", authenticationservice.GuardCodeTypeDeviceCode)
	if err != nil {
		t.Fatalf("SubmitSteamGuardCode returned error: %v", err)
	}
}

func TestPollReturnsTokensAndUpdatesClientID(t *testing.T) {
	t.Parallel()

	transport := &sequenceAuthTransport{
		responses: []authResponseSpec{
			{
				body: `{"response":{"new_client_id":"84","refresh_token":"refresh","access_token":"access","account_name":"demo"}}`,
			},
		},
	}
	client := newTestClient(t, newTestAuthService(t, transport))
	challenge := &LoginChallenge{
		SteamID:   "76561198000000001",
		ClientID:  42,
		RequestID: []byte{1, 2, 3},
	}

	result, err := client.Poll(context.Background(), challenge)
	if err != nil {
		t.Fatalf("Poll returned error: %v", err)
	}
	if result.AccountName != "demo" || result.SteamID != "76561198000000001" {
		t.Fatalf("unexpected login result: %#v", result)
	}
	if result.RefreshToken != "refresh" || result.AccessToken != "access" {
		t.Fatalf("unexpected tokens: %#v", result)
	}
	if challenge.ClientID != 84 {
		t.Fatalf("expected updated client id, got %d", challenge.ClientID)
	}
}

func TestSubmitSteamGuardCodeValidation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, newTestAuthService(t, &sequenceAuthTransport{}))

	err := client.SubmitSteamGuardCode(context.Background(), nil, "123456", authenticationservice.GuardCodeTypeDeviceCode)
	if err == nil {
		t.Fatal("expected validation error")
	}
	var clientErr *Error
	if !errors.As(err, &clientErr) {
		t.Fatalf("expected websession error, got %T", err)
	}
	if clientErr.Code != ErrorCodeRequestBuild {
		t.Fatalf("unexpected error code: %s", clientErr.Code)
	}
}
