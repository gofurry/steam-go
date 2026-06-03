# Cookbook：Steam OpenID

需要浏览器式 Steam 登录识别时，使用 `addons/openid`。

## 构造登录 URL

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

## 校验回调

```go
identity, err := verifier.VerifyRequest(r.Context(), r)
if err != nil {
	http.Error(w, "invalid Steam login", http.StatusUnauthorized)
	return
}

fmt.Fprintf(w, "SteamID64: %s", identity.SteamID)
```

## 说明

- 跳转前把 `state` 存到安全 cookie 或服务端 session。
- 回调后比较 `identity.State` 和已保存的值。
- OpenID 只证明身份，不替代 Web API key 或 access token。
- 完整示例：`go run ./examples/openid`。
