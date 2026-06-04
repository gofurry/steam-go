package steam_test

import (
	"fmt"
	"time"

	steam "github.com/gofurry/steam-go"
)

func ExampleNewClient() {
	client, err := steam.NewClient(
		steam.WithAPIKey("your-key"),
		steam.WithTimeout(10*time.Second),
		steam.WithRetry(2),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	_ = client.API.SteamUser
}

func ExampleWithSafeDefaults() {
	client, err := steam.NewClient(
		steam.WithAPIKey("your-key"),
		steam.WithSafeDefaults(),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	_ = client
}

func ExampleRedactSensitiveURL() {
	rawURL := "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/?key=SECRET&steamids=76561198370695025"

	fmt.Println(steam.RedactSensitiveURL(rawURL))

	// Output:
	// https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/?steamids=76561198370695025
}

func ExampleNewStaticProxySelector() {
	selector, err := steam.NewStaticProxySelector("http://127.0.0.1:7897")
	if err != nil {
		panic(err)
	}

	client, err := steam.NewClient(
		steam.WithAPIKey("your-key"),
		steam.WithProxySelector(selector),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	_ = client
}

func ExampleWithRequestObserver() {
	client, err := steam.NewClient(
		steam.WithRequestObserver(steam.RequestObserverFunc(func(event steam.RequestEvent) {
			_ = event.TrafficClass
			_ = event.StatusCode
			_ = event.Duration
		})),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	_ = client
}
