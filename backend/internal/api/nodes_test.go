package api

import (
	"encoding/json"
	"testing"
	"yuchen-panel/backend/internal/model"
	"yuchen-panel/backend/internal/store"
	"yuchen-panel/backend/internal/xray"
)

func TestNewNodeSecurityDefaultsAndLegacyNone(t *testing.T) {
	t.Setenv("YC_ADMIN_USERNAME", "test-administrator")
	t.Setenv("YC_ADMIN_PASSWORD", "test-password-123")
	s, err := store.Open(t.TempDir() + "/panel.json")
	if err != nil {
		t.Fatal(err)
	}
	s.Mu.Lock()
	defer s.Mu.Unlock()
	r := &Router{store: s}
	vless := model.Node{Name: "vless", Protocol: "vless", Port: 24001, Enabled: true}
	if err := r.normalizeNodeLocked(&vless, ""); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(xray.GenerateInboundConfig(vless, nil))
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil {
		t.Fatal(err)
	}
	inbound := config["inbounds"].([]any)[0].(map[string]any)
	if got := inbound["streamSettings"].(map[string]any)["security"]; got != "reality" {
		t.Fatalf("generated security=%v", got)
	}
	t.Logf("after: VLESS security=%s; Reality keys generated=%v", vless.Security, vless.RealityPrivateKey != "")
	if vless.Security != "reality" {
		t.Fatalf("new VLESS security=%q, want reality", vless.Security)
	}
	trojan := model.Node{Name: "trojan", Protocol: "trojan", Port: 24002, Enabled: true}
	if err := r.normalizeNodeLocked(&trojan, ""); err != nil {
		t.Fatal(err)
	}
	if trojan.Security != "tls" {
		t.Fatalf("new Trojan security=%q, want tls", trojan.Security)
	}
	legacy := model.Node{ID: "legacy", Name: "legacy", Protocol: "vless", Security: "none", Port: 24003}
	if err := r.normalizeNodeLocked(&legacy, legacy.ID); err != nil {
		t.Fatal(err)
	}
	if legacy.Security != "none" {
		t.Fatalf("legacy explicit none changed to %q", legacy.Security)
	}
}
