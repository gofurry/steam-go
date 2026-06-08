package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gofurry/steam-go/addons/vdf"
)

func main() {
	path := flag.String("file", "", "VDF / ACF file path")
	key := flag.String("key", "", "top-level key to print")
	flag.Parse()

	if *path == "" {
		log.Fatal("missing -file")
	}

	doc, err := vdf.ParseFile(*path)
	if err != nil {
		log.Fatal(err)
	}

	if *key != "" {
		node := doc.First(*key)
		if node == nil {
			log.Fatalf("key %q not found", *key)
		}
		fmt.Println(node.Key)
		return
	}

	for _, node := range doc.Nodes {
		if node != nil {
			fmt.Println(node.Key)
		}
	}
}
