package main

import (
	"fmt"
	"log"
	"net/http"

	steam "github.com/GoFurry/steam-go"
)

func main() {
	staticSelector, err := steam.NewStaticProxySelector("http://127.0.0.1:7897")
	if err != nil {
		log.Fatal(err)
	}

	roundRobinSelector, err := steam.NewRoundRobinProxySelector(
		"http://127.0.0.1:7897",
		"http://127.0.0.1:7898",
	)
	if err != nil {
		log.Fatal(err)
	}

	healthSelector, err := steam.NewHealthCheckedRoundRobinProxySelector(
		steam.DefaultProxyHealthConfig(),
		"http://127.0.0.1:7897",
		"http://127.0.0.1:7898",
	)
	if err != nil {
		log.Fatal(err)
	}

	routingSelector, err := steam.NewRoutingProxySelector(
		steam.ProxyRoute{
			Host:       "api.steampowered.com",
			PathPrefix: "/ISteamUser/",
			ProxyURL:   "http://127.0.0.1:7897",
		},
		steam.ProxyRoute{
			Host:       "steamcommunity.com",
			PathPrefix: "/openid/",
			ProxyURL:   "",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	requests := []*http.Request{
		mustRequest("https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/"),
		mustRequest("https://steamcommunity.com/openid/login"),
		mustRequest("https://store.steampowered.com/"),
	}

	fmt.Println("== static selector ==")
	printSelections(staticSelector, requests)

	fmt.Println("\n== round-robin selector ==")
	printSelections(roundRobinSelector, requests[:2])
	printSelections(roundRobinSelector, requests[:2])

	fmt.Println("\n== health-checked round-robin selector ==")
	printSelections(healthSelector, requests[:2])
	printSelections(healthSelector, requests[:2])
	printHealthMetrics(healthSelector)

	fmt.Println("\n== routing selector ==")
	printSelections(routingSelector, requests)
}

func printSelections(selector steam.ProxySelector, requests []*http.Request) {
	for _, req := range requests {
		proxyURL, err := selector.Next(req)
		if err != nil {
			log.Fatal(err)
		}
		if proxyURL == nil {
			fmt.Printf("%s -> direct\n", req.URL.String())
			continue
		}
		fmt.Printf("%s -> %s\n", req.URL.String(), proxyURL.String())
	}
}

func printHealthMetrics(selector steam.ProxySelector) {
	metricsProvider, ok := selector.(steam.ProxyMetricsProvider)
	if !ok {
		fmt.Println("metrics unavailable")
		return
	}

	snapshot := metricsProvider.ProxyMetricsSnapshot()
	fmt.Printf("healthy=%d cooling=%d total=%d\n", snapshot.HealthyProxies, snapshot.CoolingProxies, snapshot.TotalProxies)
	for _, proxy := range snapshot.Proxies {
		fmt.Printf(
			"%s selected=%d success=%d failure=%d cooldowns=%d\n",
			proxy.ProxyURL,
			proxy.SelectionCount,
			proxy.SuccessCount,
			proxy.FailureCount,
			proxy.CooldownCount,
		)
	}
}

func mustRequest(rawURL string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	return req
}
