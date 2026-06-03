# Cookbook：错误处理

SDK 错误可以通过 `errors.As` 识别和分类。

## 分类 SDK 错误

```go
resp, err := client.API.SteamUser.GetPlayerSummaries(ctx, []string{"76561198370695025"})
if err != nil {
	var apiErr *steam.APIError
	if errors.As(err, &apiErr) {
		fmt.Printf("kind=%s status=%d body=%s\n", apiErr.Kind, apiErr.StatusCode, apiErr.BodyPreview(512))
		return
	}
	panic(err)
}

fmt.Println(resp.Response.Players)
```

## 错误类别

- `ErrorKindRequestBuild`：输入非法或请求构造失败
- `ErrorKindTransport`：网络、超时或响应读取失败
- `ErrorKindHTTPStatus`：HTTP 非成功状态
- `ErrorKindDecode`：JSON 解码失败
- `ErrorKindAPIResponse`：Steam API 响应级失败

## 说明

- 日志中使用 `BodyPreview(max)` 记录有界响应体预览，不要打印完整响应体。
- transport failure、`429` 和部分 `5xx` 可按业务策略重试。
- 鉴权失败通常应视为凭据问题，除非重试策略明确覆盖。

## 重试建议

通常可以重试：

- `ErrorKindTransport`
- `ErrorKindHTTPStatus` 且状态为 `429`
- 临时 `5xx` 响应

通常不应在不更换凭据或用户状态的情况下重试：

- `401` 或 `403`
- 请求构造错误
- 上游 payload 形状变化导致的 decode 错误
- 明确指向账号、权限或 token 问题的 API response 错误

记录错误时，应把 `errors.As`、有界 `BodyPreview(max)`、URL/header 脱敏结合起来。不要打印原始响应体、原始请求 URL、cookie 或 authorization header。
