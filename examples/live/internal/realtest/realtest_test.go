package realtest

import "testing"

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
