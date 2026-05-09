package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/api/steamchartsservice"
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

	fmt.Println("== SteamChartsService.GetBestOfYearPages ==")
	bestOfYearPages, err := client.API.SteamChartsService.GetBestOfYearPages(ctx)
	if err != nil {
		realtest.Fatalf("GetBestOfYearPages failed: %v", err)
	}
	fmt.Printf("pages=%d latest=%q\n",
		len(bestOfYearPages.Response.Pages),
		bestOfYearPages.Response.Pages[0].Name,
	)

	fmt.Println("\n== SteamChartsService.GetGamesByConcurrentPlayers ==")
	concurrentPlayers, err := client.API.SteamChartsService.GetGamesByConcurrentPlayers(ctx)
	if err != nil {
		realtest.Fatalf("GetGamesByConcurrentPlayers failed: %v", err)
	}
	fmt.Printf("last_update=%d ranks=%d top_appid=%d\n",
		concurrentPlayers.Response.LastUpdate,
		len(concurrentPlayers.Response.Ranks),
		concurrentPlayers.Response.Ranks[0].AppID,
	)

	rtimeMonth := uint32(1746769043)
	monthLimit := uint32(10)
	includeDLC := true
	fmt.Println("\n== SteamChartsService.GetMonthTopAppReleases ==")
	monthTopReleases, err := client.API.SteamChartsService.GetMonthTopAppReleases(ctx, &steamchartsservice.GetMonthTopAppReleasesOptions{
		RTimeMonth:      &rtimeMonth,
		IncludeDLC:      &includeDLC,
		TopResultsLimit: &monthLimit,
	})
	if err != nil {
		realtest.Fatalf("GetMonthTopAppReleases failed: %v", err)
	}
	fmt.Printf("top_dlc_releases=%d combined=%d\n",
		len(monthTopReleases.Response.TopDLCReleases),
		len(monthTopReleases.Response.TopCombinedAppAndDLCReleases),
	)

	fmt.Println("\n== SteamChartsService.GetMostPlayedGames ==")
	mostPlayedGames, err := client.API.SteamChartsService.GetMostPlayedGames(ctx)
	if err != nil {
		realtest.Fatalf("GetMostPlayedGames failed: %v", err)
	}
	fmt.Printf("rollup_date=%d ranks=%d top_appid=%d\n",
		mostPlayedGames.Response.RollupDate,
		len(mostPlayedGames.Response.Ranks),
		mostPlayedGames.Response.Ranks[0].AppID,
	)

	fmt.Println("\n== SteamChartsService.GetTopReleasesPages ==")
	topReleasesPages, err := client.API.SteamChartsService.GetTopReleasesPages(ctx)
	if err != nil {
		realtest.Fatalf("GetTopReleasesPages failed: %v", err)
	}
	fmt.Printf("pages=%d first=%q\n",
		len(topReleasesPages.Response.Pages),
		topReleasesPages.Response.Pages[0].Name,
	)

	rtimeYear := uint32(1746769043)
	yearLimit := uint32(20)
	fmt.Println("\n== SteamChartsService.GetYearTopAppReleases ==")
	yearTopReleases, err := client.API.SteamChartsService.GetYearTopAppReleases(ctx, &steamchartsservice.GetYearTopAppReleasesOptions{
		RTimeYear:       &rtimeYear,
		IncludeDLC:      &includeDLC,
		TopResultsLimit: &yearLimit,
	})
	if err != nil {
		realtest.Fatalf("GetYearTopAppReleases failed: %v", err)
	}
	fmt.Printf("top_dlc_releases=%d combined=%d top_app_list=%d\n",
		len(yearTopReleases.Response.TopDLCReleases),
		len(yearTopReleases.Response.TopCombinedAppAndDLCReleases),
		len(yearTopReleases.Response.TopAppList),
	)
}
