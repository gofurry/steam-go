package live_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gofurry/steam-go/examples/live/internal/realtest"
)

func TestLiveSmokeOptIn(t *testing.T) {
	if os.Getenv("STEAM_GO_LIVE") != "1" {
		t.Skip("set STEAM_GO_LIVE=1 to run live Steam smoke checks")
	}

	cfg, err := realtest.LoadConfig()
	if err != nil {
		t.Fatalf("load live config: %v", err)
	}
	client, err := realtest.NewClient(cfg)
	if err != nil {
		t.Fatalf("create live client: %v", err)
	}
	defer client.Close()

	t.Run("steamwebapiutil", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		info, err := client.API.SteamWebAPIUtil.GetServerInfo(ctx)
		if err != nil {
			t.Fatalf("GetServerInfo failed: %v", err)
		}
		if info.ServerTime == 0 && info.ServerTimeString == "" {
			t.Fatalf("empty server info response: %#v", info)
		}
	})

	t.Run("webstorefront", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		details, err := client.Web.Storefront.GetAppDetails(ctx, realtest.DefaultAppID, nil)
		if err != nil {
			t.Fatalf("GetAppDetails failed: %v", err)
		}
		app := details["550"]
		if !app.Success || app.Data.SteamAppID != realtest.DefaultAppID {
			t.Fatalf("unexpected appdetails response: %#v", app)
		}
	})
}
