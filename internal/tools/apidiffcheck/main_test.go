package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Exercise real Git worktree registration and cleanup without fetching apidiff
// or touching the developer's checkout. The stand-in tool only uses the standard library.
func TestCheckAPICleansTemporaryWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is required for worktree cleanup tests")
	}
	repo := t.TempDir()
	for name, body := range map[string]string{
		"go.mod":     "module example.invalid/apidiff-fixture\n\ngo 1.25.0\n",
		"fixture.go": "package fixture\n\nfunc Example() {}\n",
	} {
		if err := os.WriteFile(filepath.Join(repo, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{
		{"init"},
		{"add", "go.mod", "fixture.go"},
		{"-c", "user.name=Test", "-c", "user.email=test@example.invalid", "-c", "core.hooksPath=" + t.TempDir(), "commit", "--no-gpg-sign", "-m", "test fixture"},
	} {
		if _, err := run(repo, "git", args...); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(repo)
	t.Setenv("GOWORK", "off")
	tool := filepath.Join(t.TempDir(), "apidiff.go")
	if err := os.WriteFile(tool, []byte(fakeAPIDiff), 0600); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name string
		base string
		fail string
		want string
	}{
		{"success", "HEAD", "", ""},
		{"invalid base", "refs/heads/missing-audit-base", "", "create temporary worktree"},
		{"base export fails", "HEAD", "old.apidiff", "export base API"},
		{"current export fails", "HEAD", "new.apidiff", "export current API"},
		{"comparison fails", "HEAD", "compare", "compare API"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			scratch := t.TempDir()
			for _, key := range []string{"TMPDIR", "TMP", "TEMP"} {
				t.Setenv(key, scratch)
			}
			t.Setenv("STEAM_GO_APIDIFF_TEST_FAIL", tt.fail)
			err := checkAPI(tt.base, true, tool)
			if tt.want == "" && err != nil {
				t.Fatal(err)
			}
			if tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
			worktrees, err := run(repo, "git", "worktree", "list", "--porcelain")
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(worktrees, "worktree ") != 1 {
				t.Fatalf("temporary worktree still registered:\n%s", worktrees)
			}
			leftovers, err := filepath.Glob(filepath.Join(scratch, "steam-go-apidiff-*"))
			if err != nil {
				t.Fatal(err)
			}
			if len(leftovers) != 0 {
				t.Fatalf("temporary directories remain: %v", leftovers)
			}
		})
	}
}

const fakeAPIDiff = `package main

import (
	"flag"
	"os"
	"path/filepath"
)

func main() {
	output := flag.String("w", "", "export path")
	flag.Bool("m", false, "module mode")
	flag.Bool("incompatible", false, "incompatible only")
	flag.Parse()
	stage := "compare"
	if *output != "" {
		stage = filepath.Base(*output)
	}
	if os.Getenv("STEAM_GO_APIDIFF_TEST_FAIL") == stage {
		os.Exit(1)
	}
	if *output != "" {
		if err := os.WriteFile(*output, []byte("fixture"), 0600); err != nil {
			panic(err)
		}
	}
}
`
