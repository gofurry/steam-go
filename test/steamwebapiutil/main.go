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

	realtest.PrintProxy(cfg)

	fmt.Println("== SteamWebAPIUtil.GetServerInfo ==")

	info, err := client.API.SteamWebAPIUtil.GetServerInfo(realtest.BackgroundContext())
	if err != nil {
		realtest.Fatalf("GetServerInfo failed: %v", err)
	}
	fmt.Printf("servertime=%d servertimestring=%s\n", info.ServerTime, info.ServerTimeString)

	fmt.Println("== SteamWebAPIUtil.GetSupportedAPIList ==")

	apiList, err := client.API.SteamWebAPIUtil.GetSupportedAPIList(realtest.BackgroundContext())
	if err != nil {
		realtest.Fatalf("GetSupportedAPIList failed: %v", err)
	}
	fmt.Printf("interfaces=%d\n", len(apiList.APIList.Interfaces))
	for i, iface := range apiList.APIList.Interfaces {
		if i >= 10 {
			break
		}
		fmt.Printf("[%d] interface=%s methods=%d\n", i+1, iface.Name, len(iface.Methods))
	}
}
