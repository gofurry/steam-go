package realtest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	steam "github.com/GoFurry/steam-go"
)

const (
	DefaultSteamID = "76561198370695025"
	DefaultAppID   = uint32(550)
)

// Config holds shared smoke-test configuration loaded from test credentials.
type Config struct {
	Key           string
	AccessToken   string
	FamilyGroupID string
	ProxySelector steam.ProxySelector
	ProxyLabel    string
}

// LoadConfig loads shared credentials and optional proxy settings from the test root.
func LoadConfig() (Config, error) {
	proxySelector, proxyLabel, err := loadProxySelector()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Key:           readCredential("key.txt"),
		AccessToken:   readCredential("access-token.txt"),
		FamilyGroupID: readCredential("family-group-id.txt"),
		ProxySelector: proxySelector,
		ProxyLabel:    proxyLabel,
	}, nil
}

// NewClient builds a smoke-test client from the loaded config.
func NewClient(cfg Config) (*steam.Client, error) {
	opts := []steam.Option{
		steam.WithTimeout(30 * time.Second),
		steam.WithRetry(1),
	}
	if cfg.Key != "" {
		opts = append(opts, steam.WithAPIKey(cfg.Key))
	}
	if cfg.AccessToken != "" {
		opts = append(opts, steam.WithAccessToken(cfg.AccessToken))
	}
	if cfg.ProxySelector != nil {
		opts = append(opts, steam.WithProxySelector(cfg.ProxySelector))
	}
	return steam.NewClient(opts...)
}

// BackgroundContext returns the shared context used by smoke tests.
func BackgroundContext() context.Context {
	return context.Background()
}

// PrintProxy prints the active proxy label for a smoke-test run.
func PrintProxy(cfg Config) {
	if cfg.ProxyLabel == "" {
		fmt.Println("proxy=direct")
		return
	}
	fmt.Printf("proxy=%s\n", cfg.ProxyLabel)
}

// RequireAPIKey reports whether a key is available for key-backed endpoints.
func RequireAPIKey(cfg Config) bool {
	if cfg.Key != "" {
		return true
	}
	fmt.Println("skip: test/key.txt is empty")
	return false
}

// RequireAccessToken reports whether an access token is available for token-backed endpoints.
func RequireAccessToken(cfg Config) bool {
	if cfg.AccessToken != "" {
		return true
	}
	fmt.Println("skip: test/access-token.txt is empty")
	return false
}

// RequireFamilyGroupID reports whether a family-group id is available.
func RequireFamilyGroupID(cfg Config) bool {
	if cfg.FamilyGroupID != "" {
		return true
	}
	fmt.Println("skip: test/family-group-id.txt is empty")
	return false
}

// Fatalf exits the smoke-test process with an error message.
func Fatalf(format string, args ...any) {
	fmt.Printf("ERROR: "+format+"\n", args...)
	os.Exit(1)
}

func readCredential(name string) string {
	body, err := os.ReadFile(filepath.Join(testRoot(), name))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(body))
}

func loadProxySelector() (steam.ProxySelector, string, error) {
	raw := strings.TrimSpace(os.Getenv("STEAM_PROXY"))
	if raw == "" {
		raw = readCredential("proxy.txt")
	}
	if raw == "" {
		return nil, "", nil
	}

	selector, err := steam.NewStaticProxySelector(raw)
	if err != nil {
		return nil, "", err
	}
	return selector, raw, nil
}

func testRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "test"
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
