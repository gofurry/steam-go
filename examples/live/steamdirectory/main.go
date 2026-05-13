package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/api/steamdirectory"
	"github.com/GoFurry/steam-go/examples/live/internal/realtest"
)

func main() {
	cfg, err := realtest.LoadConfig()
	if err != nil {
		realtest.Fatalf("load config failed: %v", err)
	}

	client, err := realtest.NewClient(cfg)
	if err != nil {
		realtest.Fatalf("create client failed: %v", err)
	}
	defer client.Close()

	ctx := realtest.BackgroundContext()
	realtest.PrintProxy(cfg)

	cellID := uint32(0)
	maxCount := uint32(10)
	qosLevel := uint32(2)
	fmt.Println("== SteamDirectory.GetCMListForConnect ==")
	cmList, err := client.API.SteamDirectory.GetCMListForConnect(ctx, &steamdirectory.GetCMListForConnectOptions{
		CellID:   &cellID,
		CMType:   "websockets",
		Realm:    "steamglobal",
		MaxCount: &maxCount,
		QOSLevel: &qosLevel,
	})
	if err != nil {
		realtest.Fatalf("GetCMListForConnect failed: %v", err)
	}
	fmt.Printf("success=%t servers=%d first=%q\n",
		cmList.Response.Success,
		len(cmList.Response.ServerList),
		cmList.Response.ServerList[0].Endpoint,
	)

	fmt.Println("\n== SteamDirectory.GetSteamPipeDomains ==")
	domains, err := client.API.SteamDirectory.GetSteamPipeDomains(ctx)
	if err != nil {
		realtest.Fatalf("GetSteamPipeDomains failed: %v", err)
	}
	fmt.Printf("result=%d domains=%d first=%q\n",
		domains.Response.Result,
		len(domains.Response.DomainList),
		domains.Response.DomainList[0],
	)
}
