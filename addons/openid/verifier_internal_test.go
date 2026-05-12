package openid

import "testing"

func TestParseSteamIDRejectsMalformedVariants(t *testing.T) {
	t.Parallel()

	tests := []string{
		"http://steamcommunity.com/openid/id/76561198000000000",
		"https://example.com/openid/id/76561198000000000",
		"https://steamcommunity.com/openid/id/76561198000000000/extra",
		"https://steamcommunity.com/openid/id/not-a-number",
		"https://steamcommunity.com/not-openid/id/76561198000000000",
	}

	for _, raw := range tests {
		raw := raw
		t.Run(raw, func(t *testing.T) {
			t.Parallel()
			if _, err := parseSteamID(raw); err == nil {
				t.Fatalf("expected parseSteamID to reject %q", raw)
			}
		})
	}
}

func FuzzParseSteamID(f *testing.F) {
	seeds := []string{
		"https://steamcommunity.com/openid/id/76561198000000000",
		"http://steamcommunity.com/openid/id/76561198000000000",
		"https://STEAMCOMMUNITY.COM/openid/id/76561198000000000",
		"https://steamcommunity.com/openid/id/not-a-number",
		"https://steamcommunity.com/openid/id/76561198000000000/extra",
		"://bad",
		"",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		_, _ = parseSteamID(raw)
	})
}

func FuzzVerifyReturnTo(f *testing.F) {
	verifier, err := NewVerifier(Config{
		Realm:    "https://example.com",
		ReturnTo: "https://example.com/auth/steam/callback",
	})
	if err != nil {
		f.Fatalf("NewVerifier returned error: %v", err)
	}

	seeds := []string{
		"https://example.com/auth/steam/callback",
		"https://example.com/auth/steam/callback?state=demo",
		"https://example.com/auth/steam/callback?state=demo&state=dup",
		"https://example.com/other",
		"://bad",
		"",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		_, _ = verifier.verifyReturnTo(raw)
	})
}
