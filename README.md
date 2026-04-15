# steam-go

`steam-go` is a lightweight Go SDK focused on the Steam Web API.

## Features

- Root `Client` with service-oriented access to `SteamUser`, `PlayerService`, `SteamNews`, and `SteamUserStats`
- Functional options for API key, access token, timeout, retry, rate limit, proxy selection, and logger injection
- `key` and `access_token` are treated as different credentials and can be configured independently
- API key is optional and can be supplied through a rotating key provider
- Typed responses by default with matching raw response methods
- `401/429` can automatically retry with the next API key when `WithAPIKeys(...)` and `WithRetry(...)` are used together
- Shared request executor, centralized endpoint registry, and a single SDK error model
- No crawler, HTML parsing, A2S, heavy logging, or broad util package baggage

## Installation

```bash
go get github.com/GoFurry/steam-go@latest
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"time"

	steam "github.com/GoFurry/steam-go"
)

func main() {
	client, err := steam.NewClient(
		steam.WithTimeout(10*time.Second),
		steam.WithRetry(2),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	resp, err := client.SteamUser.GetPlayerSummaries(
		context.Background(),
		[]string{"76561197960435530"},
	)
	if err != nil {
		panic(err)
	}

	for _, player := range resp.Response.Players {
		fmt.Printf("%s: %s\n", player.SteamID, player.PersonaName)
	}
}
```

## API overview

- `client.SteamUser.GetPlayerSummaries(ctx, steamIDs)`
- `client.SteamUser.GetPlayerSummariesRaw(ctx, steamIDs)`
- `client.PlayerService.GetOwnedGames(ctx, steamID, opts)`
- `client.PlayerService.GetOwnedGamesRaw(ctx, steamID, opts)`
- `client.SteamNews.GetNewsForApp(ctx, appID, opts)`
- `client.SteamNews.GetNewsForAppRaw(ctx, appID, opts)`
- `client.SteamUserStats.GetPlayerAchievements(ctx, steamID, appID, opts)`
- `client.SteamUserStats.GetPlayerAchievementsRaw(ctx, steamID, appID, opts)`

## Options

- `WithAPIKey(key string)`
- `WithAPIKeys(keys ...string)`
- `WithAPIKeyProvider(provider APIKeyProvider)`
- `WithAccessToken(token string)`
- `WithAccessTokens(tokens ...string)`
- `WithAccessTokenProvider(provider AccessTokenProvider)`
- `WithBaseURL(url string)`
- `WithHTTPClient(client *http.Client)`
- `WithTimeout(timeout time.Duration)`
- `WithRetry(retry int)`
- `WithRateLimit(requestsPerSecond int)`
- `WithProxySelector(selector ProxySelector)`
- `WithLogger(logger Logger)`

## Error handling

SDK errors use `*steam.APIError` with these kinds:

- `request_build`
- `transport`
- `http_status`
- `decode`
- `api_response`

Use `errors.As(err, &apiErr)` to inspect kind, status code, and raw body.

## Examples

- [steamuser](examples/steamuser/main.go)
- [playerservice](examples/playerservice/main.go)
- [steamnews](examples/steamnews/main.go)
- [steamuserstats](examples/steamuserstats/main.go)
