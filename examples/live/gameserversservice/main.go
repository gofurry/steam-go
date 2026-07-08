package main

import (
	"fmt"

	"github.com/gofurry/steam-go/api/gameserversservice"
	"github.com/gofurry/steam-go/examples/live/internal/realtest"
)

func main() {
	cfg, err := realtest.LoadConfig()
	if err != nil {
		realtest.Fatalf("load config failed: %v", err)
	}
	if !realtest.RequireAPIKey(cfg) {
		return
	}

	client, err := realtest.NewClient(cfg)
	if err != nil {
		realtest.Fatalf("create client failed: %v", err)
	}
	defer client.Close()

	ctx := realtest.BackgroundContext()
	realtest.PrintProxy(cfg)

	fmt.Println("== GameServersService.GetServerList ==")

	filter := gameserversservice.NewFilter().AppID(realtest.DefaultAppID)
	resp, err := client.API.GameServersService.GetServerList(ctx, &gameserversservice.GetServerListOptions{
		Filter: filter.String(),
		Limit:  100,
	})
	if err != nil {
		realtest.Fatalf("GetServerList failed: %v", err)
	}

	fmt.Printf("filter=%s limit=%d servers=%d\n", filter.String(), 100, len(resp.Response.Servers))
	for i, server := range resp.Response.Servers {
		if i >= 10 {
			break
		}
		fmt.Printf(
			"[%d] addr=%s appid=%d name=%q map=%s players=%d/%d bots=%d secure=%t dedicated=%t\n",
			i+1,
			server.Addr,
			server.AppID,
			server.Name,
			server.Map,
			server.Players,
			server.MaxPlayers,
			server.Bots,
			server.Secure,
			server.Dedicated,
		)
	}
}
