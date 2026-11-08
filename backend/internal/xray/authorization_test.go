package xray

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
	"yuchen-panel/backend/internal/model"
)

func TestGenerateServerConfigExcludesUnauthorizedClients(t *testing.T) {
	node := model.Node{ID: "node", Protocol: "vless", Enabled: true, Port: 23456, Transport: "tcp", Security: "none"}
	clients := []model.Client{
		{ID: "empty", UUID: "uuid-empty", Enabled: true},
		{ID: "fixed", UUID: "uuid-fixed", Enabled: true, RelayRouteIDs: []string{"relay"}},
		{ID: "disabled", UUID: "uuid-disabled", NodeIDs: []string{"node"}, Enabled: false},
		{ID: "expired", UUID: "uuid-expired", NodeIDs: []string{"node"}, Enabled: true, ExpireAt: time.Now().Add(-time.Hour)},
		{ID: "other", UUID: "uuid-other", NodeIDs: []string{"other"}, Enabled: true},
		{ID: "valid", UUID: "uuid-valid", NodeIDs: []string{"node"}, Enabled: true},
	}
	raw, err := json.Marshal(GenerateServerConfig([]model.Node{node}, clients, nil, nil, model.NetworkPolicy{}))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range clients[:5] {
		if strings.Contains(string(raw), c.UUID) {
			t.Errorf("unauthorized UUID: %s", c.UUID)
		}
	}
	if !strings.Contains(string(raw), "uuid-valid") {
		t.Fatal("valid client missing")
	}
}

func TestGenerateShadowsocksInbound(t *testing.T) {
	node := model.Node{ID: "ss-node", Protocol: "shadowsocks", Enabled: true, Port: 23456, SSMethod: "aes-256-gcm", SSPassword: "ss-password"}
	cfg := GenerateServerConfig([]model.Node{node}, nil, nil, nil, model.NetworkPolicy{})
	inbounds := cfg["inbounds"].([]any)
	var inbound map[string]any
	for _, raw := range inbounds {
		candidate := raw.(map[string]any)
		if candidate["tag"] == node.ID {
			inbound = candidate
			break
		}
	}
	if inbound == nil {
		t.Fatal("Shadowsocks inbound missing")
	}
	if inbound["protocol"] != "shadowsocks" || inbound["streamSettings"] != nil {
		t.Fatalf("unexpected Shadowsocks transport config: %#v", inbound)
	}
	settings := inbound["settings"].(map[string]any)
	if settings["method"] != node.SSMethod || settings["password"] != node.SSPassword {
		t.Fatalf("settings=%#v", settings)
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), node.SSPassword) {
		t.Fatal("Shadowsocks password missing from generated config")
	}
}
