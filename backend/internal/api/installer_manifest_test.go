package api

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallerRejectsMaliciousManifestBeforeExtraction(t *testing.T) {
	script, err := filepath.Abs("../../../install.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, key, value, root string
		valid                  bool
	}{
		{"valid", "version", "1.0.3-stable", "", true},
		{"version-traversal", "version", "../../danger", "", false},
		{"shell", "version", "1.0.3-$(id)", "", false},
		{"package-traversal", "package", "../bad.zip", "", false},
		{"package-dotdot", "package", "bad..zip", "", false},
		{"http", "download_url", "http://example.com/p.zip", "", false},
		{"root", "version", "1.0.3", "/", false},
		{"multiple-root", "version", "1.0.3", "////", false},
		{"relative-root", "version", "1.0.3", "relative", false},
		{"parent-root", "version", "1.0.3", "/tmp/../danger", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			temp := t.TempDir()
			root := filepath.Join(temp, "install root")
			if tc.root != "" {
				root = tc.root
			}
			fixture := filepath.Join(temp, "manifest.json")
			fields := map[string]string{"version": "1.0.3", "package": "panel.zip", "download_url": "https://example.com/p.zip", "sha256": strings.Repeat("0", 64)}
			fields[tc.key] = tc.value
			raw, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(fixture, raw, 0600); err != nil {
				t.Fatal(err)
			}
			program := `set -Eeuo pipefail
source "$1"
curl() {
  local output=''
  while (( $# )); do
    if [[ $1 == -o ]]; then output=$2; shift 2; else shift; fi
  done
  cp -- "$2_FIXTURE" "$output"
}
`
			// Positional parameters are copied before mock curl receives its own argv.
			program = strings.ReplaceAll(program, `"$2_FIXTURE"`, `"$fixture_path"`)
			program += "fixture_path=$2\nfetch_manifest\n"
			cmd := exec.Command("bash", "-c", program, "test", script, fixture)
			cmd.Env = append(os.Environ(), "TMPDIR="+temp, "YC_INSTALL_ROOT="+root)
			out, err := cmd.CombinedOutput()
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v err=%v output=%s", tc.valid, err, out)
			}
			if _, err := os.Stat(root); err == nil && tc.root == "" {
				t.Fatal("validation created installation directory")
			}
		})
	}
}

func TestInstallerExtractsIntoFreshDirectory(t *testing.T) {
	temp := t.TempDir()
	root := filepath.Join(temp, "install")
	old := filepath.Join(root, "yuchen-panel-1.0.3")
	if err := os.MkdirAll(old, 0700); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(old, "preserve")
	if err := os.WriteFile(sentinel, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	script, err := filepath.Abs("../../../install.sh")
	if err != nil {
		t.Fatal(err)
	}
	program := `set -Eeuo pipefail
source "$1"
VERSION=1.0.3
PACKAGE_NAME=panel.zip
DOWNLOAD_URL=https://example.com/p.zip
PACKAGE_FILE=unused
unzip() { mkdir -p "$5/deploy"; touch "$5/deploy/install.sh"; }
extract_release
printf '%s' "$RELEASE_ROOT"
`
	cmd := exec.Command("bash", "-c", program, "test", script)
	cmd.Env = append(os.Environ(), "TMPDIR="+temp, "YC_INSTALL_ROOT="+root)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatal("existing release was removed")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatal("fresh staging directory not created")
	}
}
