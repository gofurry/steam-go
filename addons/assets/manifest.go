package assets

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// NewURLManifest creates a manifest from constructed URL items.
func NewURLManifest(items []URLItem) Manifest {
	return Manifest{URLs: items}
}

// NewDownloadManifest creates a manifest from download results.
func NewDownloadManifest(results []DownloadResult) Manifest {
	return Manifest{Downloads: results}
}

// MarshalManifestJSON encodes a manifest as indented JSON.
func MarshalManifestJSON(manifest Manifest) ([]byte, error) {
	return json.MarshalIndent(manifest, "", "  ")
}

// WriteManifestJSON writes a manifest as indented JSON.
func WriteManifestJSON(path string, manifest Manifest) error {
	data, err := MarshalManifestJSON(manifest)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
