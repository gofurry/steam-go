package openid_test

import (
	"fmt"
	"net/url"

	"github.com/gofurry/steam-go/addons/openid"
)

func Example() {
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

	parsed, err := url.Parse(loginURL)
	if err != nil {
		panic(err)
	}
	returnTo, err := url.Parse(parsed.Query().Get("openid.return_to"))
	if err != nil {
		panic(err)
	}

	fmt.Println(parsed.Host)
	fmt.Println(parsed.Query().Get("openid.mode"))
	fmt.Println(returnTo.Query().Get("state"))

	// Output:
	// steamcommunity.com
	// checkid_setup
	// csrf-state
}
