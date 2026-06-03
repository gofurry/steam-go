# Cookbook: Error Handling

SDK errors can be inspected through `errors.As`.

## Classify SDK Errors

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

## Error Kinds

- `ErrorKindRequestBuild`: invalid input or request construction failure
- `ErrorKindTransport`: network, timeout, or response read failure
- `ErrorKindHTTPStatus`: non-success HTTP response
- `ErrorKindDecode`: JSON decode failure
- `ErrorKindAPIResponse`: Steam API response-level failure

## Notes

- Log bounded previews with `BodyPreview(max)`, not full response bodies.
- Retry transport failures, `429`, and some `5xx` responses according to your workload.
- Treat authentication and authorization failures as credential problems unless a retry policy says otherwise.
