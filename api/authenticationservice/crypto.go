package authenticationservice

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
)

// EncryptPasswordPKCS1v15 encrypts a plaintext password with Steam's RSA public key.
//
// The helper only performs local encryption for IAuthenticationService requests. It does not
// store credentials or start an auth session by itself.
func EncryptPasswordPKCS1v15(password string, publicKeyModHex string, publicKeyExpHex string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password must not be empty")
	}
	modulus, ok := new(big.Int).SetString(strings.TrimSpace(publicKeyModHex), 16)
	if !ok || modulus.Sign() <= 0 {
		return "", fmt.Errorf("public key modulus must be a positive hex integer")
	}
	exponentValue, ok := new(big.Int).SetString(strings.TrimSpace(publicKeyExpHex), 16)
	if !ok || !exponentValue.IsInt64() || exponentValue.Sign() <= 0 {
		return "", fmt.Errorf("public key exponent must be a positive hex integer")
	}

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, &rsa.PublicKey{
		N: modulus,
		E: int(exponentValue.Int64()),
	}, []byte(password))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}
