package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestNoHardcodedBackendReleaseVersions(t *testing.T) {
	// go test runs from this package's directory: backend/cmd/server.
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("locate backend module: %v", err)
	}
	versionPattern := regexp.MustCompile(`1\.0\.\d+`)
	var violations []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if filepath.ToSlash(rel) == "internal/version/version.go" {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for line := 1; scanner.Scan(); line++ {
			if versionPattern.MatchString(scanner.Text()) {
				violations = append(violations, fmt.Sprintf("backend/%s:%d: %s", filepath.ToSlash(rel), line, strings.TrimSpace(scanner.Text())))
			}
		}
		return scanner.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Fatalf("hardcoded release versions outside internal/version/version.go:\n%s", strings.Join(violations, "\n"))
	}
	t.Log("0 hardcoded release versions outside the single version source")
}
