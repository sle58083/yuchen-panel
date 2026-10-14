package xray

import (
	"encoding/json"
	"os"
	"testing"

	"yuchen-panel/backend/internal/model"
)

func TestStatsConfigAndUniqueClientEmail(t *testing.T) {
	node := model.Node{ID: "n", Protocol: "vless", Port: 24401, Transport: "tcp", Security: "none", Enabled: true}
	clients := []model.Client{{ID: "first", Username: "same", UUID: "test-uuid-1", Password: "fixture-password", Enabled: true, NodeIDs: []string{"n"}}, {ID: "second", Username: "same", UUID: "test-uuid-2", Password: "fixture-password-2", Enabled: true, NodeIDs: []string{"n"}}}
	for _, protocol := range []string{"vless", "vmess", "trojan"} {
		node.Protocol = protocol
		cfg := GenerateInboundConfig(node, clients)
		if _, ok := cfg["stats"]; !ok {
			t.Fatal("no stats")
		}
		level := cfg["policy"].(map[string]any)["levels"].(map[string]any)["0"].(map[string]any)
		if level["statsUserUplink"] != true || level["statsUserDownlink"] != true {
			t.Fatal("user stats disabled")
		}
		api := cfg["api"].(map[string]any)
		if api["tag"] != "api" || api["services"].([]string)[0] != "StatsService" {
			t.Fatal("no StatsService")
		}
		inbounds := cfg["inbounds"].([]any)
		users := inbounds[0].(map[string]any)["settings"].(map[string]any)["clients"].([]map[string]any)
		if len(users) != 2 || users[0]["email"] != "client_first" || users[1]["email"] != "client_second" {
			t.Fatalf("identity %v", users)
		}
		stats := inbounds[len(inbounds)-1].(map[string]any)
		if stats["listen"] != "127.0.0.1" || stats["port"] != 10000 || stats["tag"] != "api" {
			t.Fatalf("unsafe API %v", stats)
		}
		if _, ok := stats["sniffing"]; ok {
			t.Fatal("internal stats inbound should not sniff")
		}
		first := cfg["routing"].(map[string]any)["rules"].([]any)[0].(map[string]any)
		if first["outboundTag"] != "api" {
			t.Fatal("Stats API is not routed first")
		}
		if protocol == "vless" {
			raw, _ := json.MarshalIndent(cfg, "", "  ")
			t.Logf("generated Stats config:\n%s", raw)
			if path := os.Getenv("YC_STATS_FIXTURE"); path != "" {
				if err := os.WriteFile(path, raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestStatsPortAvoidsNodeCollision(t *testing.T) {
	cfg := GenerateInboundConfig(model.Node{ID: "n", Protocol: "vless", Port: 10000, Enabled: true}, nil)
	in := cfg["inbounds"].([]any)
	stats := in[len(in)-1].(map[string]any)
	if stats["port"] != 10001 || stats["listen"] != "127.0.0.1" {
		t.Fatal(stats)
	}
}

func TestStatsRelayEmailUsesClientID(t *testing.T) {
	users := makeRelayVlessClients([]model.Client{{ID: "first", Username: "same", UUID: "one", Enabled: true, RelayRouteIDs: []string{"r"}}, {ID: "second", Username: "same", UUID: "two", Enabled: true, RelayRouteIDs: []string{"r"}}}, "r")
	if len(users) != 2 || users[0]["email"] != "client_first" || users[1]["email"] != "client_second" {
		t.Fatal(users)
	}
}
