# Cookbook：基础官方 API

使用 `client.API.*` 调用官方 `api.steampowered.com` 方法。

## 获取玩家摘要

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

## 说明

- `client.API.*` 是官方 Steam Web API surface。
- `WithAPIKey(...)` 会在 endpoint 请求未显式设置 `key` 时注入 API key。
- `WithRetry(...)` 会重试 transport failure 和可重试 HTTP 状态。
- 可用服务分组见 [API 参考](../api/reference.md)。
