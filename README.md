# steam-go

<p align="center">
  <img src="https://img.shields.io/badge/License-MIT-6C757D?style=flat&color=3B82F6" alt="License">&nbsp&nbsp&nbsp
  <img src="https://img.shields.io/github/v/release/gofurry/steam-go?style=flat&color=blue" alt="Release">&nbsp&nbsp&nbsp
  <img src="https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=flat&logo=go&logoColor=white" alt="Go Version">&nbsp&nbsp&nbsp
  <a href="https://goreportcard.com/report/github.com/gofurry/steam-go">
    <img src="https://goreportcard.com/badge/github.com/gofurry/steam-go" alt="Go Report Card">
  </a>&nbsp&nbsp&nbsp
  <img src="https://img.shields.io/badge/weekend-project-8B5CF6?style=flat" alt="Weekend Project">&nbsp&nbsp&nbsp
  <img src="https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange" alt="Made with Love">
</p>

<p align="center">
  ⭐
  <a href="https://github.com/gofurry/steam-go">English</a>
  &nbsp;|&nbsp;
  <a href="https://github.com/gofurry/steam-go/wiki/Steam-Keys-and-Access-Tokens">Steam Keys and Access Tokens</a>&nbsp;|&nbsp;
  <a href="https://github.com/gofurry/steam-go/wiki/%E9%A6%96%E9%A1%B5">中文Wiki</a>
  &nbsp;|&nbsp;
  <a href="https://github.com/gofurry/steam-go/wiki/Steam-Key-%E4%B8%8E-Access-Token">Steam Key 与 Access Token</a>
  ⭐
</p>

```text
				____ ____ ____ _  _ ____ ____ _   _   / ____ ___ ____ ____ _  _    ____ ____ 
				| __ |  | |___ |  | |__/ |__/  \_/   /  [__   |  |___ |__| |\/| __ | __ |  | 
				|__] |__| |    |__| |  \ |  \   |   /   ___]  |  |___ |  | |  |    |__] |__| 
                                                                             
```

`steam-go` is a stable Go SDK for the official Steam Web API, with practical request controls and carefully scoped read-only Steam Web helpers.

## Why Use It

- Stable root `Client` with grouped official API access under `client.API.*`
- Read-only `client.Web.*` helpers for Storefront, Community inventory, and Market JSON endpoints
- Functional options for API key, access token, timeout, retry, rate limit, proxy, cookies, and traffic policy
- Safe defaults for external traffic with bounded response bodies and URL redaction helpers
- Typed responses for stable payloads, with raw methods and `json.RawMessage` for volatile subtrees
- Rotating and health-checked API key providers for resilient `401/429` handling
- Addons for OpenID, web sessions, assets, free-claim workflows, and A2S without bloating the core SDK

## Installation

```bash
go get github.com/gofurry/steam-go@latest
```

`steam-go` currently requires Go 1.25+.

## Quick Start: Official API

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

More official API examples: [docs/cookbook/basic-api.md](docs/cookbook/basic-api.md).

## Quick Start: Steam OpenID

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

In a real app, store the `state` in a secure cookie or server-side session before redirecting the user. Cookbook: [docs/cookbook/auth-openid.md](docs/cookbook/auth-openid.md).

## Quick Start: Read-Only Web Query

```go
client, err := steam.NewClient(steam.WithSafeDefaults())
if err != nil {
	panic(err)
}
defer client.Close()

reviews, err := client.Web.Storefront.GetAppReviews(context.Background(), 440, &storefront.GetAppReviewsOptions{
	Language:   "english",
	NumPerPage: 20,
})
if err != nil {
	panic(err)
}

fmt.Println(reviews.QuerySummary.TotalReviews)
```

`client.Web.*` is read-only and never injects Steam Web API `key` or `access_token`. Cookbook: [docs/cookbook/web-readonly.md](docs/cookbook/web-readonly.md).

## Production Notes

- Start with `WithSafeDefaults()` for real external traffic, then tune timeout, retry, and rate limit per workload.
- Use `WithProxySelector(...)` or `WithTrafficPolicy(...)` when region, host, or session behavior needs different network paths.
- Do not log raw URLs that may contain `key`, `access_token`, cookies, or proxy credentials. Use `steam.RedactSensitiveURL(...)`.
- Use `WithMaxResponseBodyBytes(...)` when callers need a stricter response body cap.
- Keep live smoke credentials and web cookies out of Git; examples use environment variables or hidden terminal prompts for sensitive values.

## Stability Boundary

- `client.API.*` is the official Steam Web API surface and the main stable `v1` contract.
- `client.Web.*` exposes stable Go method signatures, but the upstream Store / Community / Market payloads are unofficial and may drift.
- Volatile nested payloads may remain `json.RawMessage` instead of being forced into brittle typed structs.
- The core package does not include browser fallback, purchase, sell, trade, or bulk account automation.

## Addons

| Addon | Use it for |
|---|---|
| `addons/openid` | Steam OpenID login verification |
| `addons/websession` | Manual Steam web-login session flow |
| `addons/freeclaim` | Read-only free promotion discovery plus explicit single-package claim |
| `addons/assets` | Store / Library asset URL construction, verification, reading, and downloading |
| `addons/a2s` | A2S server queries through `github.com/gofurry/a2s-go` |

Detailed addon notes: [docs/addons/reference.md](docs/addons/reference.md).

## Documentation

- [Documentation index](docs/README.md)
- [API reference](docs/api/reference.md)
- [Web reference](docs/web/reference.md)
- [Addon reference](docs/addons/reference.md)
- [Cookbook](docs/cookbook/basic-api.md)
- [Compatibility policy](docs/governance/compatibility.md)
- [Endpoint stability](docs/governance/endpoint-stability.md)
- [Endpoint coverage](docs/governance/endpoint-coverage.md)
- [Release checklist](docs/releases/checklist.md)
