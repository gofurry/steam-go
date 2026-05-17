package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofurry/steam-go/examples/live/internal/realtest"
)

func main() {
	cfg, err := realtest.LoadConfig()
	if err != nil {
		realtest.Fatalf("load config: %v", err)
	}
	client, err := realtest.NewClient(cfg)
	if err != nil {
		realtest.Fatalf("new client: %v", err)
	}
	defer client.Close()

	accountName := strings.TrimSpace(os.Getenv("STEAM_AUTH_ACCOUNT_NAME"))
	if accountName == "" {
		fmt.Println("skip: STEAM_AUTH_ACCOUNT_NAME is empty")
		return
	}

	realtest.PrintProxy(cfg)

	key, err := client.API.AuthenticationService.GetPasswordRSAPublicKey(realtest.BackgroundContext(), accountName)
	if err != nil {
		realtest.Fatalf("GetPasswordRSAPublicKey: %v", err)
	}
	fmt.Printf("rsa_key mod_len=%d exp=%s timestamp=%d\n", len(key.PublicKeyMod), key.PublicKeyExp, key.Timestamp)
}
