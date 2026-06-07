package main

import (
	"fmt"

	"github.com/gofurry/steam-go/addons/markup"
)

func main() {
	input := `[p]\[ PATCH ][/p][list][*][b]New map[/b][/*][/list]`

	html, err := markup.CleanSteamContent(input)
	if err != nil {
		panic(err)
	}
	summary, err := markup.Summary(input, 80)
	if err != nil {
		panic(err)
	}

	fmt.Println(html)
	fmt.Println(summary)
}
