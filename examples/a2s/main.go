package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"
	"time"

	a2saddon "github.com/GoFurry/steam-go/addons/a2s"
)

func main() {
	server := flag.String("server", "", "game server address, e.g. 1.2.3.4:27015")
	query := flag.String("query", "info", "query type: info, players, or rules")
	timeout := flag.Duration("timeout", 3*time.Second, "query timeout")
	flag.Parse()

	if *server == "" {
		log.Fatal("missing -server")
	}

	client, err := a2saddon.NewClient(
		*server,
		a2saddon.WithTimeout(*timeout),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()

	switch *query {
	case "info":
		info, err := client.QueryInfo(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf(
			"name=%s map=%s players=%d/%d vac=%t\n",
			info.Name,
			info.Map,
			info.Players,
			info.MaxPlayers,
			info.VAC,
		)
	case "players":
		players, err := client.QueryPlayers(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("players=%d\n", len(players.Players))
		for _, player := range players.Players {
			fmt.Printf("- %s score=%d duration=%.0fs\n", player.Name, player.Score, player.Duration)
		}
	case "rules":
		rules, err := client.QueryRules(ctx)
		if err != nil {
			log.Fatal(err)
		}
		keys := make([]string, 0, len(rules.Items))
		for key := range rules.Items {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		fmt.Printf("rules=%d\n", len(keys))
		for _, key := range keys {
			fmt.Printf("- %s=%s\n", key, rules.Items[key])
		}
	default:
		log.Fatalf("unsupported -query %q", *query)
	}
}
