package openid

import "testing"

func TestVerifyReturnToRejectsDuplicateStateParameter(t *testing.T) {
	t.Parallel()

	verifier, err := NewVerifier(Config{
		Realm:    "https://example.com",
		ReturnTo: "https://example.com/auth/steam/callback",
	})
	if err != nil {
		t.Fatalf("NewVerifier returned error: %v", err)
	}

	_, err = verifier.verifyReturnTo("https://example.com/auth/steam/callback?state=demo&state=extra")
	if err == nil {
		t.Fatal("expected duplicate state parameter to be rejected")
	}
}

func TestVerifyReturnToRejectsUnexpectedExtraQuery(t *testing.T) {
	t.Parallel()

	verifier, err := NewVerifier(Config{
		Realm:    "https://example.com",
		ReturnTo: "https://example.com/auth/steam/callback",
	})
	if err != nil {
		t.Fatalf("NewVerifier returned error: %v", err)
	}

	_, err = verifier.verifyReturnTo("https://example.com/auth/steam/callback?state=demo&other=1")
	if err == nil {
		t.Fatal("expected unexpected query parameter to be rejected")
	}
}
