package realtest

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadCredentialEnvOrFilePrefersEnvironment(t *testing.T) {
	t.Setenv("STEAM_GO_TEST_CREDENTIAL", "  environment-value  ")
	if got := readCredentialEnvOrFile("STEAM_GO_TEST_CREDENTIAL", "missing-test-credential.txt"); got != "environment-value" {
		t.Fatalf("credential = %q", got)
	}
}

func TestReadCredentialEnvOrFileMissing(t *testing.T) {
	t.Setenv("STEAM_GO_TEST_CREDENTIAL", "")
	if got := readCredentialEnvOrFile("STEAM_GO_TEST_CREDENTIAL", "missing-test-credential.txt"); got != "" {
		t.Fatalf("credential = %q, want empty", got)
	}
}

func TestPrintProxyRedactsCredentials(t *testing.T) {
	for _, tt := range []struct {
		name  string
		label string
		want  string
	}{
		{"direct", "", "proxy=direct\n"},
		{"proxy", "http://proxy.example:8080", "proxy=http://proxy.example:8080\n"},
		{"credentials", "http://example-user:example-password@proxy.example:8080?access_token=example-token", "proxy=http://proxy.example:8080\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			output, err := os.CreateTemp(t.TempDir(), "stdout")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			original := os.Stdout
			os.Stdout = output
			defer func() { os.Stdout = original }()
			PrintProxy(Config{ProxyLabel: tt.label})
			if _, err := output.Seek(0, io.SeekStart); err != nil {
				t.Fatal(err)
			}
			got, err := io.ReadAll(output)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Fatalf("proxy output = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadProxySelectorRedactsLabel(t *testing.T) {
	t.Setenv("STEAM_PROXY", "http://example-user:example-password@proxy.example:8080")
	selector, label, err := loadProxySelector()
	if err != nil {
		t.Fatal(err)
	}
	if selector == nil || label != "http://proxy.example:8080" {
		t.Fatalf("proxy selector missing or display label is not redacted: %q", label)
	}
	proxy, err := selector.Next(httptest.NewRequest("GET", "https://example.invalid/", nil))
	if err != nil {
		t.Fatal(err)
	}
	if proxy == nil || proxy.User == nil || proxy.User.Username() != "example-user" {
		t.Fatal("display redaction changed the proxy authentication username")
	}
	if password, ok := proxy.User.Password(); !ok || password != "example-password" {
		t.Fatal("display redaction changed the proxy authentication password")
	}
}

func TestLegacyTestRootIsAtRepositoryRoot(t *testing.T) {
	root := legacyTestRoot()
	if filepath.Base(root) != "test" {
		t.Fatalf("legacy credential directory = %q, want test", root)
	}
	module, err := os.ReadFile(filepath.Join(filepath.Dir(root), "go.mod"))
	if err != nil {
		t.Fatalf("legacy credential directory must be beside the repository go.mod: %v", err)
	}
	if !strings.HasPrefix(string(module), "module github.com/gofurry/steam-go\n") &&
		!strings.HasPrefix(string(module), "module github.com/gofurry/steam-go\r\n") {
		t.Fatal("legacy credential directory is outside the steam-go repository")
	}
}

func TestReadCredentialFromRoots(t *testing.T) {
	for _, tt := range []struct {
		name    string
		current *string
		legacy  *string
		want    string
	}{
		{"current wins", stringPointer(" current-value \n"), stringPointer("legacy-value"), "current-value"},
		{"missing current", nil, stringPointer(" legacy-value \n"), "legacy-value"},
		{"empty current", stringPointer(" \n"), stringPointer("legacy-value"), "legacy-value"},
		{"missing both", nil, nil, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			roots := []string{t.TempDir(), t.TempDir()}
			for i, content := range []*string{tt.current, tt.legacy} {
				if content != nil {
					if err := os.WriteFile(filepath.Join(roots[i], "credential.txt"), []byte(*content), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			if got := readCredentialFromRoots("credential.txt", roots); got != tt.want {
				t.Fatalf("credential = %q, want %q", got, tt.want)
			}
		})
	}
}

func stringPointer(value string) *string { return &value }
