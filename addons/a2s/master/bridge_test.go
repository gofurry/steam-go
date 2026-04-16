package master_test

import (
	"testing"
	"time"

	masteraddon "github.com/GoFurry/steam-go/addons/a2s/master"
)

func TestNewClientBridgesToMaster(t *testing.T) {
	t.Parallel()

	client, err := masteraddon.NewClient(
		masteraddon.WithBaseAddress("127.0.0.1:27011"),
		masteraddon.WithTimeout(2*time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	if client == nil {
		t.Fatal("expected client")
	}
}

func TestStartCursorBridge(t *testing.T) {
	t.Parallel()

	cursor := masteraddon.StartCursor()
	if !cursor.IsZero() {
		t.Fatal("expected zero cursor")
	}
}
