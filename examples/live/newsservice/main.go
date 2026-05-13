package main

import (
	"fmt"

	"github.com/GoFurry/steam-go/api/newsservice"
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

	preserveNewlines := false

	realtest.PrintProxy(cfg)
	fmt.Println("== NewsService.ConvertHTMLToBBCode ==")

	resp, err := client.API.NewsService.ConvertHTMLToBBCode(
		realtest.BackgroundContext(),
		"<strong>ok</strong>",
		&newsservice.ConvertHTMLToBBCodeOptions{PreserveNewlines: &preserveNewlines},
	)
	if err != nil {
		realtest.Fatalf("ConvertHTMLToBBCode failed: %v", err)
	}

	fmt.Printf("found_html=%t converted=%s\n", resp.Response.FoundHTML, resp.Response.ConvertedContent)
}
