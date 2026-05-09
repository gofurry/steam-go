package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/test/internal/realtest"
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

	fmt.Println("== SteamApps.GetSDRConfig ==")
	sdrConfig, err := client.API.SteamApps.GetSDRConfig(ctx, realtest.DefaultAppID)
	if err != nil {
		realtest.Fatalf("GetSDRConfig failed: %v", err)
	}
	fmt.Printf("revision=%d pops=%d certs=%d success=%t\n",
		sdrConfig.Revision,
		len(sdrConfig.Pops),
		len(sdrConfig.Certs),
		sdrConfig.Success,
	)

	fmt.Println("\n== SteamApps.GetServersAtAddress ==")
	serversAtAddress, err := client.API.SteamApps.GetServersAtAddress(ctx, "45.125.45.95")
	if err != nil {
		realtest.Fatalf("GetServersAtAddress failed: %v", err)
	}
	fmt.Printf("success=%t servers=%d\n",
		serversAtAddress.Response.Success,
		len(serversAtAddress.Response.Servers),
	)

	fmt.Println("\n== SteamApps.UpToDateCheck ==")
	upToDate, err := client.API.SteamApps.UpToDateCheck(ctx, realtest.DefaultAppID, 1)
	if err != nil {
		realtest.Fatalf("UpToDateCheck failed: %v", err)
	}
	fmt.Printf("success=%t up_to_date=%t required_version=%d message=%q\n",
		upToDate.Response.Success,
		upToDate.Response.UpToDate,
		upToDate.Response.RequiredVersion,
		upToDate.Response.Message,
	)
}
