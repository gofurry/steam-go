# Cookbook: Credential Redaction

Steam credentials often appear in URL query strings. Redact before logging.

## Redact URLs

```go
rawURL := "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/?key=SECRET&steamids=76561198370695025"
safeURL := steam.RedactSensitiveURL(rawURL)

fmt.Println(safeURL)
```

`RedactSensitiveURL(...)` removes URL userinfo plus known Steam credential query parameters such as `key` and `access_token`.

## Sensitive Values

Do not log or paste:

- API keys or publisher keys
- access tokens
- refresh tokens
- `steamLoginSecure`
- `sessionid`
- raw `Cookie` headers
- proxy URLs with usernames or passwords

## Notes

- Prefer environment variables or secret managers for credentials.
- Do not pass refresh tokens or passwords through command-line flags.
- Keep live smoke credential files out of Git.
