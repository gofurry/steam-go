package assets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestJSON(t *testing.T) {
	manifest := NewDownloadManifest([]DownloadResult{
		{
			AppID:  550,
			Kind:   KindHeader,
			URL:    "https://example.test/header.jpg",
			Path:   "550/header.jpg",
			Status: DownloadStatusDownloaded,
		},
	})

	data, err := MarshalManifestJSON(manifest)
	if err != nil {
		t.Fatalf("MarshalManifestJSON returned error: %v", err)
	}
	if !strings.Contains(string(data), `"downloads"`) || !strings.Contains(string(data), `"downloaded"`) {
		t.Fatalf("manifest JSON = %s", data)
	}

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := WriteManifestJSON(path, manifest); err != nil {
		t.Fatalf("WriteManifestJSON returned error: %v", err)
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if !strings.HasSuffix(string(written), "\n") {
		t.Fatalf("manifest should end with newline")
	}
}
