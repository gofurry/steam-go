package scanner_test

import (
	"testing"
	"time"

	scanneraddon "github.com/GoFurry/steam-go/addons/a2s/scanner"
)

func TestNewClientBridgesToScanner(t *testing.T) {
	t.Parallel()

	client, err := scanneraddon.NewClient(
		scanneraddon.WithConcurrency(2),
		scanneraddon.WithTimeout(2*time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client")
	}
}

func TestParseAddressBridge(t *testing.T) {
	t.Parallel()

	server, err := scanneraddon.ParseAddress("127.0.0.1")
	if err != nil {
		t.Fatalf("ParseAddress returned error: %v", err)
	}
	if got := server.String(); got != "127.0.0.1:27015" {
		t.Fatalf("unexpected server: %s", got)
	}
}
