package websession

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofurry/steam-go/api/authenticationservice"
)

const defaultWebsiteID = "Store"

type StartWithCredentialsRequest struct {
	AccountName string
	Password    string
	DeviceName  string
	WebsiteID   string
	Language    uint64
}

type LoginChallenge struct {
	SteamID              string
	ClientID             uint64
	RequestID            []byte
	PollInterval         time.Duration
	AllowedConfirmations []authenticationservice.AuthConfirmation
}

type LoginResult struct {
	AccountName  string
	SteamID      string
	RefreshToken string
	AccessToken  string
}

func (c *Client) StartWithCredentials(ctx context.Context, req StartWithCredentialsRequest) (*LoginChallenge, error) {
	req.AccountName = strings.TrimSpace(req.AccountName)
	req.DeviceName = strings.TrimSpace(req.DeviceName)
	req.WebsiteID = strings.TrimSpace(req.WebsiteID)
	if req.AccountName == "" {
		return nil, &Error{Code: ErrorCodeRequestBuild, Op: "start_with_credentials", Message: "account name must not be empty"}
	}
	if req.Password == "" {
		return nil, &Error{Code: ErrorCodeRequestBuild, Op: "start_with_credentials", Message: "password must not be empty"}
	}
	if req.WebsiteID == "" {
		req.WebsiteID = defaultWebsiteID
	}

	key, err := c.auth.GetPasswordRSAPublicKey(ctx, req.AccountName)
	if err != nil {
		return nil, err
	}
	encrypted, err := authenticationservice.EncryptPasswordPKCS1v15(req.Password, key.PublicKeyMod, key.PublicKeyExp)
	if err != nil {
		return nil, &Error{Code: ErrorCodeRequestBuild, Op: "start_with_credentials", Message: "encrypt password failed", Err: err}
	}

	session, err := c.auth.BeginAuthSessionViaCredentials(ctx, authenticationservice.BeginAuthSessionViaCredentialsRequest{
		DeviceFriendlyName:  req.DeviceName,
		AccountName:         req.AccountName,
		EncryptedPassword:   encrypted,
		EncryptionTimestamp: key.Timestamp,
		RememberLogin:       true,
		WebsiteID:           req.WebsiteID,
		Language:            req.Language,
	})
	if err != nil {
		return nil, err
	}
	return &LoginChallenge{
		SteamID:              session.SteamID,
		ClientID:             session.ClientID,
		RequestID:            append([]byte(nil), session.RequestID...),
		PollInterval:         time.Duration(session.Interval) * time.Second,
		AllowedConfirmations: append([]authenticationservice.AuthConfirmation(nil), session.AllowedConfirmations...),
	}, nil
}

func (c *Client) SubmitSteamGuardCode(ctx context.Context, challenge *LoginChallenge, code string, typ authenticationservice.GuardCodeType) error {
	if challenge == nil {
		return &Error{Code: ErrorCodeRequestBuild, Op: "submit_steam_guard_code", Message: "challenge must not be nil"}
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return &Error{Code: ErrorCodeRequestBuild, Op: "submit_steam_guard_code", Message: "guard code must not be empty"}
	}
	_, err := c.auth.UpdateAuthSessionWithSteamGuardCode(ctx, authenticationservice.UpdateAuthSessionWithSteamGuardCodeRequest{
		ClientID: challenge.ClientID,
		SteamID:  challenge.SteamID,
		Code:     code,
		CodeType: typ,
	})
	var resultErr *authenticationservice.EResultError
	if errors.As(err, &resultErr) && resultErr.Code == 29 {
		return nil
	}
	return err
}

func (c *Client) Poll(ctx context.Context, challenge *LoginChallenge) (*LoginResult, error) {
	if challenge == nil {
		return nil, &Error{Code: ErrorCodeRequestBuild, Op: "poll", Message: "challenge must not be nil"}
	}
	resp, err := c.auth.PollAuthSessionStatus(ctx, authenticationservice.PollAuthSessionStatusRequest{
		ClientID:  challenge.ClientID,
		RequestID: challenge.RequestID,
	})
	if err != nil {
		return nil, err
	}
	if resp.NewClientID != 0 {
		challenge.ClientID = resp.NewClientID
	}
	return &LoginResult{
		AccountName:  resp.AccountName,
		SteamID:      challenge.SteamID,
		RefreshToken: resp.RefreshToken,
		AccessToken:  resp.AccessToken,
	}, nil
}
