package api

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallerRefreshesWeakSecretWithoutEvaluatingEnv(t *testing.T) {
	raw, err := os.ReadFile("../../../install.sh")
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(string(raw), "<<'PY_CREDENTIALS'\n")
	if len(parts) != 2 {
		t.Fatal("credential preflight not found")
	}
	program := strings.Split(parts[1], "\nPY_CREDENTIALS")[0]
	path := filepath.Join(t.TempDir(), ".env")
	other := "UNRELATED=$(touch /should-never-exist)\n"
	if err = os.WriteFile(path, []byte("YC_JWT_SECRET=short\n"+other), 0644); err != nil {
		t.Fatal(err)
	}
	run := func() {
		cmd := exec.Command("python3", "-", path)
		cmd.Stdin = strings.NewReader(program)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}
	run()
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), other) {
		t.Fatal("unrelated setting changed")
	}
	var secret string
	for _, line := range strings.Split(string(updated), "\n") {
		if strings.HasPrefix(line, "YC_JWT_SECRET=") {
			secret = strings.TrimPrefix(line, "YC_JWT_SECRET=")
		}
	}
	if len(secret) != 64 {
		t.Fatalf("generated secret length %d", len(secret))
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatal("credential file is not private")
	}
	run()
	again, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(updated) != string(again) {
		t.Fatal("strong secret unexpectedly rotated")
	}
}
