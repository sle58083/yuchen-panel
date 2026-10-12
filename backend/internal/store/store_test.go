package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"yuchen-panel/backend/internal/model"
)

func TestSeedDefaultAdminRequiresExplicitCredentials(t *testing.T) {
	for _, tc := range []struct{ user, password string }{
		{"", ""}, {"admin", ""}, {"admin", "admin123"}, {"admin", "change-me"}, {"", "long-explicit-password"},
	} {
		t.Setenv("YC_ADMIN_USERNAME", tc.user)
		t.Setenv("YC_ADMIN_PASSWORD", tc.password)
		if _, err := Open(filepath.Join(t.TempDir(), "db.json")); err == nil {
			t.Error("weak initial credentials accepted")
		}
	}
	t.Setenv("YC_ADMIN_USERNAME", "operator")
	t.Setenv("YC_ADMIN_PASSWORD", "long-explicit-password")
	s, err := Open(filepath.Join(t.TempDir(), "db.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Data.Admins) != 1 {
		t.Fatal("administrator not initialized")
	}
}

func TestMissingHistoricalHashDisablesAccount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db.json")
	data := newData()
	data.Admins["old"] = model.AdminUser{ID: "old", Enabled: true, Role: "super_admin"}
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if a := s.Data.Admins["old"]; a.Enabled || a.PasswordHash != "" {
		t.Fatal("account silently regained credentials")
	}
}
