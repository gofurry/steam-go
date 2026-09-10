// This example is entirely offline; it needs no Steam client or credentials.
package main

import (
	"errors"
	"fmt"

	"github.com/gofurry/steam-go/steamid"
)

func main() {
	// Small decimal values are AccountIDs in the Public Individual Desktop context.
	id, err := steamid.Parse("12345")
	if err != nil {
		panic(err)
	}
	fmt.Println("SteamID64:", id)
	fmt.Printf("AccountID=%d Universe=%d AccountType=%d Instance=%d Valid=%t\n",
		id.AccountID(), id.Universe(), id.AccountType(), id.Instance(), id.Valid())

	steam2, err := id.Steam2()
	if err != nil {
		panic(err)
	}
	steam3, err := id.Steam3()
	if err != nil {
		panic(err)
	}
	fmt.Println("Steam2:", steam2)
	fmt.Println("Steam3:", steam3)

	profile, err := steamid.ParseCommunityURL("https://steamcommunity.com/profiles/76561197960278073/")
	if err != nil {
		panic(err)
	}
	fmt.Println("Same numeric profile:", profile == id)

	_, err = steamid.ParseCommunityURL("https://steamcommunity.com/id/example-user")
	if errors.Is(err, steamid.ErrVanityReference) {
		// Network resolution is separate and is intentionally not performed here.
		fmt.Println("Vanity: resolve the token with client.API.SteamUser.ResolveVanityURL.")
	} else if err != nil {
		panic(err)
	}
}
