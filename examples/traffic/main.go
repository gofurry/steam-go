package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	steam "github.com/GoFurry/steam-go"
)

func main() {
	profile := steam.DefaultPublicStoreHeaderProfileEN()

	client, err := steam.NewClient(
		steam.WithAPIKey("your-key"),
		steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
			HeaderProfile: &profile,
			TransportHook: steam.TransportHookFunc(func(class steam.TrafficClass, base *http.Client) (*http.Client, error) {
				cloned := *base
				if transport, ok := base.Transport.(*http.Transport); ok {
					custom := transport.Clone()
					custom.TLSClientConfig = &tls.Config{
						MinVersion: tls.VersionTLS12,
					}
					cloned.Transport = custom
				}
				return &cloned, nil
			}),
			Cache: &steam.TrafficCachePolicy{TTL: time.Minute},
		}),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	fmt.Println("traffic client with public store-page transport hook is ready")
}
