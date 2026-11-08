package xray

import (
	"testing"

	"yuchen-panel/backend/internal/model"
)

func TestTCPRelayPreservesFixedDestination(t *testing.T) {
	landing := model.Node{ID: "landing", Host: "192.0.2.20", Port: 24443, Enabled: true}
	relay := model.RelayRoute{ID: "relay", LandingNodeID: landing.ID, RelayPort: 24444, Enabled: true, RouteMode: "tcp_forward"}
	cfg := GenerateServerConfig(nil, nil, []model.RelayRoute{relay}, map[string]model.Node{landing.ID: landing}, model.NetworkPolicy{})
	inbounds := cfg["inbounds"].([]any)
	if len(inbounds) != 1 {
		t.Fatalf("expected one relay inbound, got %d", len(inbounds))
	}
	inbound := inbounds[0].(map[string]any)
	if inbound["sniffing"].(map[string]any)["enabled"] != false {
		t.Fatal("fixed relay must not sniff and override the REALITY destination")
	}
	settings := inbound["settings"].(map[string]any)
	if settings["address"] != landing.Host || settings["port"] != landing.Port || settings["network"] != "tcp" {
		t.Fatalf("unexpected relay destination: %#v", settings)
	}
}
