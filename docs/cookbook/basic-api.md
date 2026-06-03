# Cookbook: Basic Official API

Use `client.API.*` for official `api.steampowered.com` methods.

## Get Player Summaries

```go
package main

import (
	"context"
	"fmt"
	"time"

	steam "github.com/gofurry/steam-go"
)

func main() {
	client, err := steam.NewClient(
		steam.WithAPIKey("your-key"),
		steam.WithTimeout(10*time.Second),
		steam.WithRetry(2),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	resp, err := client.API.SteamUser.GetPlayerSummaries(
		context.Background(),
		[]string{"76561198370695025"},
	)
	if err != nil {
		panic(err)
	}

	for _, player := range resp.Response.Players {
		fmt.Printf("%s: %s\n", player.SteamID, player.PersonaName)
	}
}
```

## Notes

- `client.API.*` is the official Steam Web API surface.
- `WithAPIKey(...)` injects `key` only when the endpoint request does not already set one.
- `WithRetry(...)` retries transport failures and retryable HTTP statuses.
- Use [API reference](../api/reference.md) for available service groups.
