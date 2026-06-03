# Cookbook: Steam OpenID

Use `addons/openid` when you need browser-based Steam sign-in.

## Build a Login URL

```go
verifier, err := openid.NewVerifier(openid.Config{
	Realm:    "https://example.com/",
	ReturnTo: "https://example.com/auth/steam/callback",
})
if err != nil {
	panic(err)
}

loginURL, err := verifier.LoginURL("csrf-state")
if err != nil {
	panic(err)
}

fmt.Println(loginURL)
```

## Verify the Callback

```go
identity, err := verifier.VerifyRequest(r.Context(), r)
if err != nil {
	http.Error(w, "invalid Steam login", http.StatusUnauthorized)
	return
}

fmt.Fprintf(w, "SteamID64: %s", identity.SteamID)
```

## Notes

- Store `state` in a secure cookie or server-side session before redirecting.
- Compare the returned `identity.State` with the stored value.
- OpenID proves identity only. It does not replace Web API keys or access tokens.
- Full example: `go run ./examples/openid`.
