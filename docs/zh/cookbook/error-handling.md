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
