package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReloadFailureRemainsPendingUntilSuccessfulRetry(t *testing.T) {
	cfg := Config{ConfigPath: filepath.Join(t.TempDir(), "config.json"), ReloadCommand: "exit 1"}
	response := SyncResponse{XrayConfig: map[string]any{"inbounds": []any{}}}
	if err := applyXrayConfig(cfg, response); err == nil {
		t.Fatal("failed reload reported success")
	}
	if _, err := os.Stat(cfg.ConfigPath + ".reload-pending"); err != nil {
		t.Fatal("retry marker lost", err)
	}
	cfg.ReloadCommand = "true"
	if err := applyXrayConfig(cfg, response); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cfg.ConfigPath + ".reload-pending"); !os.IsNotExist(err) {
		t.Fatal("successful reload did not clear marker")
	}
}
