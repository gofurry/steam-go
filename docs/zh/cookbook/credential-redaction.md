# Cookbook：凭据脱敏

Steam 凭据经常出现在 URL query string 中，记录日志前应先脱敏。

## URL 脱敏

```go
rawURL := "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/?key=SECRET&steamids=76561198370695025"
safeURL := steam.RedactSensitiveURL(rawURL)

fmt.Println(safeURL)
```

`RedactSensitiveURL(...)` 会移除 URL userinfo，以及 `key`、`access_token` 等已知 Steam 凭据 query 参数。

## Header 脱敏

```go
safeHeaders := steam.RedactSensitiveHeaders(req.Header)
```

`Authorization`、`Proxy-Authorization`、`Cookie`、`Set-Cookie`、`X-WebAPI-Key`、`X-API-Key` 等敏感 header 会被替换为 `[REDACTED]`。

`RedactSensitiveHeaders(...)` 返回 clone，不会修改原始 header map。

## 敏感值

不要记录或粘贴：

- API key 或 publisher key
- access token
- refresh token
- `steamLoginSecure`
- `sessionid`
- 原始 `Cookie` header
- 带用户名或密码的 proxy URL

## 说明

- 生产环境优先使用环境变量或 secret manager 管理凭据。
- 不要通过命令行参数传入 refresh token 或密码。
- live smoke 凭据文件不应进入 Git。
