package steamid_test

import (
	"errors"
	"fmt"

	"github.com/gofurry/steam-go/steamid"
)

func ExampleParse() {
	id, err := steamid.Parse("12345")
	if err != nil {
		panic(err)
	}
	steam2, err := id.Steam2()
	if err != nil {
		panic(err)
	}
	steam3, err := id.Steam3()
	if err != nil {
		panic(err)
	}
	fmt.Println(id, id.AccountID(), id.Universe(), id.AccountType(), id.Instance())
	fmt.Println(steam2)
	fmt.Println(steam3)
	// Output:
	// 76561197960278073 12345 1 1 1
	// STEAM_1:1:6172
	// [U:1:12345]
}

func ExampleParseCommunityURL() {
	id, err := steamid.ParseCommunityURL("https://steamcommunity.com/profiles/76561197960278073/")
	if err != nil {
		panic(err)
	}
	fmt.Println(id)
	_, err = steamid.ParseCommunityURL("https://steamcommunity.com/id/example-user")
	fmt.Println(errors.Is(err, steamid.ErrVanityReference))
	// Resolve a vanity token separately with client.API.SteamUser.ResolveVanityURL.
	// Output:
	// 76561197960278073
	// true
}
