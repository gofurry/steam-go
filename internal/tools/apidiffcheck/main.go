package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const defaultAPIDiffTool = "golang.org/x/exp/cmd/apidiff@v0.0.0-20260529124908-c761662dc8c9"

func main() {
	var (
		baseRef      string
		incompatible bool
		apidiffTool  string
	)
	flag.StringVar(&baseRef, "base", "", "base git ref or tag to compare against")
	flag.BoolVar(&incompatible, "incompatible", false, "show only incompatible API changes")
	flag.StringVar(&apidiffTool, "apidiff", defaultAPIDiffTool, "apidiff command module passed to go run")
	flag.Parse()

	if strings.TrimSpace(baseRef) == "" {
		fatalf("-base is required, for example: go run ./internal/tools/apidiffcheck -base v1.2.2")
	}

	root, err := gitOutput("", "rev-parse", "--show-toplevel")
	if err != nil {
		fatalf("resolve git root: %v", err)
	}
	root = strings.TrimSpace(root)

	modulePath, err := goOutput(root, "list", "-m")
	if err != nil {
		fatalf("resolve module path: %v", err)
	}
	modulePath = strings.TrimSpace(modulePath)
	if modulePath == "" {
		fatalf("empty module path")
	}

	tmpRoot, err := os.MkdirTemp("", "steam-go-apidiff-*")
	if err != nil {
		fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpRoot)

	baseDir := filepath.Join(tmpRoot, "base")
	if _, err := run(root, "git", "worktree", "add", "--detach", baseDir, baseRef); err != nil {
		fatalf("create temporary worktree for %q: %v", baseRef, err)
	}
	defer func() {
		_, _ = run(root, "git", "worktree", "remove", "--force", baseDir)
	}()

	oldExport := filepath.Join(tmpRoot, "old.apidiff")
	newExport := filepath.Join(tmpRoot, "new.apidiff")

	if _, err := run(baseDir, "go", "run", apidiffTool, "-m", "-w", oldExport, modulePath); err != nil {
		fatalf("export base API: %v", err)
	}
	if _, err := run(root, "go", "run", apidiffTool, "-m", "-w", newExport, modulePath); err != nil {
		fatalf("export current API: %v", err)
	}

	args := []string{"run", apidiffTool}
	if incompatible {
		args = append(args, "-incompatible")
	}
	args = append(args, "-m", oldExport, newExport)

	out, err := run(root, "go", args...)
	if out != "" {
		fmt.Print(out)
	}
	if err != nil {
		fatalf("compare API: %v", err)
	}
	if noMeaningfulDiff(out) {
		fmt.Printf("No API differences found against %s.\n", baseRef)
	}
}

func gitOutput(dir string, args ...string) (string, error) {
	return run(dir, "git", args...)
}

func goOutput(dir string, args ...string) (string, error) {
	return run(dir, "go", args...)
}

func run(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := stdout.String()
	if stderr.Len() > 0 {
		out += stderr.String()
	}
	if err != nil {
		return out, fmt.Errorf("%s %s failed: %w\n%s", name, strings.Join(args, " "), err, out)
	}
	return out, nil
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func noMeaningfulDiff(out string) bool {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Ignoring internal package ") {
			continue
		}
		return false
	}
	return true
}
