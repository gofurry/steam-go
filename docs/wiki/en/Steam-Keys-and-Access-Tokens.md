# Steam Keys and Access Tokens

This page explains common credential types you may encounter when using `steam-go` or the Steam Web API: Steam Web API Key, Access Token, Publisher Web API Key / Partner Key, and OpenID identity verification.

> Access token retrieval depends on Steam Store / Community web login state. It is web-surface behavior and may change. Do not treat it as a long-term stable replacement for official Web API keys.

## Quick Summary

| Credential | Purpose | Long-Term Stable? | Where to Keep It |
|---|---|---|---|
| Steam Web API Key | Calls Steam Web API methods that require `key` | Usually long-lived until reset or revoked | Backend config / server environment variables |
| Access Token | Authenticates some Store / Community web APIs for a logged-in user | Expires | Temporary use only; avoid long-term storage |
| Publisher Web API Key / Partner Key | Steamworks partner / publisher backend APIs | Highly sensitive, higher privilege | Trusted servers only |
| OpenID | Lets users sign in through Steam | Not an API key | Used during sign-in flow |

```text
Web API Key   = key for official Steam Web API calls
Access Token  = short-lived token from Steam web login state
Publisher Key = high-privilege key for Steamworks partner backends
OpenID        = proves who the user is
```

## 1. What Is a Steam Web API Key?

A Steam Web API Key is the most common credential used when calling the Steam Web API.

Many endpoints require:

```text
key=YOUR_STEAM_WEB_API_KEY
```

In `steam-go`, configure it like this:

```go
client, err := steam.NewClient(
    steam.WithAPIKey("your-steam-web-api-key"),
)
```

Typical uses include player summaries, friend lists, achievements, owned games, and methods that require a normal Web API key.

Important:

```text
A Steam Web API Key is not a user login session.
A Steam Web API Key is not a Publisher Key.
```

## 2. How to Get a Steam Web API Key

A normal Steam Web API Key is usually obtained here:

```text
https://steamcommunity.com/dev/apikey
```

General steps:

1. Sign in to your Steam account.
2. Open `https://steamcommunity.com/dev/apikey`.
3. Register a Web API Key.
4. Enter a domain name.
5. Accept the terms.
6. Copy the generated key.

For backend services, keep the key in `.env`, server environment variables, secret management systems, or CI/CD secrets. Do not hardcode it into source code.

```bash
STEAM_WEB_API_KEY=your-key-here
```

```go
key := os.Getenv("STEAM_WEB_API_KEY")

client, err := steam.NewClient(
    steam.WithAPIKey(key),
)
```

## 3. Steam Web API Authentication Levels

From a user perspective, Steam Web API authentication can be grouped into three levels:

| Type | Meaning |
|---|---|
| Public methods | Return public data and usually require no authentication |
| User Web API key methods | Require a normal user Web API key |
| Protected / publisher methods | Require a Publisher Web API Key and are intended for trusted backends |

The official Steamworks Web API documentation describes both public methods and protected methods. Protected methods require authentication and are intended to be called from trusted backend applications.

## 4. How the API Key Is Sent to Steam

The most common way is a query parameter:

```text
https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/?key=YOUR_KEY&steamids=7656119...
```

Some references also mention an HTTP header:

```text
x-webapi-key: YOUR_KEY
```

For most users, the `key` query parameter is still the most common and compatible approach.

In `steam-go`, you usually do not need to build the query yourself:

```go
steam.WithAPIKey("your-key")
```

## 5. What Is an Access Token?

An Access Token is not the same thing as a Web API Key. It is closer to:

```text
a user-authenticated token used by Steam Store / Community web pages.
```

It is commonly used by Steam's own Store and Community web APIs for authenticated users.

Characteristics:

- usually tied to the currently logged-in Steam user;
- expires;
- not supported by every API;
- some APIs only support keys;
- some APIs only support access tokens;
- usually passed as the `access_token` parameter.

In `steam-go`, configure it like this:

```go
client, err := steam.NewClient(
    steam.WithAccessToken("your-access-token"),
)
```

Multiple tokens:

```go
client, err := steam.NewClient(
    steam.WithAccessTokens("token-a", "token-b"),
)
```

## 6. How to Get an Access Token

> Access tokens are sensitive login-state credentials. Only inspect them in your own browser, with your own Steam account, in your own test environment. Do not paste tokens into untrusted websites or share them with others.

### 6.1 Store Token

After signing in to Steam in your browser, open:

```text
https://store.steampowered.com/pointssummary/ajaxgetasyncconfig
```

Copy the value of:

```text
webapi_token
```

This token is usually used for Store-related authenticated web APIs.

### 6.2 Community Token

After signing in to Steam Community, open:

```text
https://steamcommunity.com/my/edit/info
```

Then run this in the browser DevTools Console:

```js
JSON.parse(application_config.dataset.loyalty_webapi_token)
```

You may also manually copy the value from the `application_config` element:

```text
data-loyalty_webapi_token
```

This token is usually used for Community-related authenticated web APIs.

## 7. Important Access Token Properties

The biggest difference between an Access Token and a normal API key is:

```text
Access Tokens expire.
```

You may see information like:

```text
Currently entered token is for web:store with steamid 76561198370695025
and expires on May 14, 2026, 14:32.
```

This means the token may have a scope, a SteamID, and an expiration time. You need a new token after expiration, and a leaked token can be abused before it expires.

Better fits:

```text
temporary testing
manual smoke validation
debugging Store / Community authenticated web APIs
```

Poor fits:

```text
long-term .env configuration
database storage as a permanent credential
frontend code
sharing with third-party tools
```

## 8. What Is a Publisher Web API Key / Partner Key?

A Publisher Web API Key is a higher-privilege key used by Steamworks partners, publishers, and secure backend servers.

It is typically used for publisher-only methods, sensitive data access, protected actions, and Steamworks backend server calls.

Normal key:

```text
for normal Steam Web API use
```

Publisher key:

```text
for Steamworks Partner / Publisher backend APIs
```

## 9. Public Host and Partner Host

Common public Steam Web API host:

```text
https://api.steampowered.com
```

Some public Web API traffic may also use:

```text
https://community.steam-api.com
```

Partner-only host:

```text
https://partner.steam-api.com
```

The partner host has stricter requirements:

- HTTPS only;
- every request requires a valid Publisher Web API Key;
- even methods that normally do not need a key on the public host require a publisher key on the partner host;
- missing or invalid publisher keys may return `403`;
- repeated `403` responses can trigger strict rate limits or temporary deny listing for the connecting IP;
- do not connect directly by IP; use the DNS name.

```text
Do not use a normal Web API Key against partner.steam-api.com.
```

## 10. How Is OpenID Different From Keys?

Steam can act as an OpenID provider. OpenID is used for:

```text
letting users sign in to your website through Steam
```

It confirms:

```text
this user owns this SteamID64.
```

It is not a Web API Key, Access Token, Publisher Key, or your application session.

Typical flow:

```text
1. User clicks "Sign in through Steam"
2. Redirect to Steam OpenID login
3. Steam redirects back to your callback
4. Your server verifies the callback
5. You obtain the user's SteamID64
6. You create your own application session
```

In `steam-go`, this is handled by:

```text
addons/openid
```

## 11. Key and Token Security Recommendations

### Do Not Commit Keys to Git

Avoid:

```go
client, _ := steam.NewClient(
    steam.WithAPIKey("ABCDEF123456"),
)
```

Prefer:

```go
client, _ := steam.NewClient(
    steam.WithAPIKey(os.Getenv("STEAM_WEB_API_KEY")),
)
```

### Do Not Put Keys in Frontend Code

Do not put keys in Vue / React / Nuxt frontend code, browser localStorage, public JavaScript bundles, or GitHub Pages static sites.

Recommended architecture:

```text
Browser -> Your Backend -> steam-go -> Steam Web API
```

Avoid:

```text
Browser -> Steam Web API + raw key
```

### Do Not Log Raw URLs

Steam Web API keys and access tokens are often passed through query parameters.

Dangerous log:

```text
https://api.steampowered.com/xxx?key=SECRET&access_token=TOKEN
```

Use `steam-go` redaction:

```go
safeURL := steam.RedactSensitiveURL(rawURL)
```

### Do Not Treat Access Tokens as Long-Term Credentials

Access Tokens expire and are tied to web login state. Do not store them long-term in `.env`, databases, or server-side permanent configuration.

### Keep Publisher Keys on Trusted Servers Only

Publisher Web API Keys have higher privilege and must be stored only in trusted backend servers, secure CI/CD secrets, or dedicated secret management systems. Do not distribute them in game clients, desktop clients, mobile apps, frontend pages, or public repositories.

## 12. Which Credential Should I Use in steam-go?

### Normal Web API Requests

```go
client, err := steam.NewClient(
    steam.WithAPIKey(os.Getenv("STEAM_WEB_API_KEY")),
)
```

Useful for:

```text
GetPlayerSummaries
GetFriendList
GetOwnedGames
GetPlayerAchievements
GetNewsForApp
```

### APIs That Require Access Tokens

```go
client, err := steam.NewClient(
    steam.WithAccessToken(os.Getenv("STEAM_ACCESS_TOKEN")),
)
```

Useful for some Store / Community authenticated web APIs, or APIs that only accept `access_token`.

### Multiple Keys

```go
client, err := steam.NewClient(
    steam.WithAPIKeys(
        os.Getenv("STEAM_WEB_API_KEY_A"),
        os.Getenv("STEAM_WEB_API_KEY_B"),
    ),
)
```

Use this when you legitimately own multiple keys and want to distribute request load.

### Multiple Keys with Cooldown

```go
client, err := steam.NewClient(
    steam.WithHealthCheckedAPIKeys(
        steam.DefaultAPIKeyHealthConfig(),
        os.Getenv("STEAM_WEB_API_KEY_A"),
        os.Getenv("STEAM_WEB_API_KEY_B"),
    ),
    steam.WithRetry(2),
)
```

Useful when one key hits `401` / `429` and should be temporarily avoided. This is for resilience, not for bypassing Steam limits.

## 13. FAQ

### Q: Does a Web API Key expire?

A normal Web API Key is usually more like a long-term credential until you reset, revoke, or lose access to it. It must still be protected as a sensitive credential.

### Q: Does an Access Token expire?

Yes. Access Tokens usually contain an expiration time and must be retrieved again after expiration.

### Q: Can an Access Token replace a Web API Key?

Not generally. Some APIs support access tokens, some only support Web API keys, and some require specific credential types.

### Q: Can I put a Publisher Key in a client application?

No. Publisher Web API Keys must remain on trusted backend servers. Do not distribute them in game clients, desktop clients, mobile apps, or frontend pages.

### Q: Does OpenID login allow me to call all Web APIs?

No. OpenID only proves user identity and usually gives you the user's SteamID64. Calling Web APIs still depends on the endpoint's required credential type: key, access token, or publisher key.

### Q: Should I publish my key in a README?

Never. If a key leaks, reset or revoke it immediately and check logs, CI secrets, deployment configs, and repositories for remaining copies.

## 14. Recommended Practices

| Scenario | Recommendation |
|---|---|
| Normal public data queries | Use a normal Steam Web API Key |
| Backend service | Store the key in environment variables or secrets |
| User login | Use Steam OpenID |
| Temporary Store / Community authenticated testing | Use Access Token, but do not store it long-term |
| Publisher backend APIs | Use Publisher Web API Key |
| Production logs | Redact with `steam.RedactSensitiveURL(...)` |
| High request volume | Use `WithSafeDefaults()`, rate limiting, and retries |
| Multiple keys | Use health-checked key provider, but do not bypass limits |

## 15. References

- Steam Web API Key registration: `https://steamcommunity.com/dev/apikey`
- Steam Web API Terms of Use: `https://steamcommunity.com/dev/apiterms`
- Steamworks Web API Overview: `https://partner.steamgames.com/doc/webapi_overview`
- Steam Web API Explorer / xPaw: `https://steamapi.xpaw.me/`
