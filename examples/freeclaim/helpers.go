package main

import "github.com/gofurry/steam-go/examples/internal/secretinput"

func resolveRefreshToken(resolver secretinput.Resolver) (string, error) {
	return resolver.ResolveSensitive("STEAM_REFRESH_TOKEN", "Steam refresh token (input hidden): ")
}
