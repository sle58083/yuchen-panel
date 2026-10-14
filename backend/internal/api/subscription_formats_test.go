package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"yuchen-panel/backend/internal/model"
	"yuchen-panel/backend/internal/security"
	"yuchen-panel/backend/internal/xray"
)

const fixtureUUID = "f9b77af1-736a-4f5b-8bc5-273ba9b552ea"

func TestSubscriptionFormatRouting(t *testing.T) {
	s := authorizationStore(t)
	s.Data.Clients["c"] = model.Client{ID: "c", Username: "fixture", UUID: fixtureUUID, Enabled: true, NodeIDs: []string{"n", "old"}, SubscribeToken: "private-token-do-not-log"}
	s.Data.Nodes["n"] = model.Node{ID: "n", Name: "node", Protocol: "vless", Transport: "tcp", Security: "none", Host: "192.0.2.1", Port: 23456, Enabled: true}
	s.Data.Nodes["old"] = model.Node{ID: "old", Protocol: "shadowsocks", Enabled: true}
	r := &Router{store: s}
	for _, tc := range []struct{ query, ua, format string }{
		{"?format=singbox", "", "singbox"}, {"?format=sing-box", "", "singbox"}, {"?format=sb", "", "singbox"}, {"?format=json", "", "singbox"},
		{"?format=clash", "", "clash"}, {"?format=v2rayn", "sing-box", "v2rayn"}, {"?format=unknown", "Clash", "v2rayn"},
		{"", "Clash Verge", "clash"}, {"", "Mihomo", "clash"}, {"", "Stash", "clash"}, {"", "sing-box", "singbox"}, {"", "SFA/1.12", "singbox"}, {"", "SFI/1.12", "singbox"},
		{"", "v2rayNG", "v2rayn"}, {"", "", "v2rayn"}, {"?format=raw", "", "raw"},
	} {
		req := httptest.NewRequest("GET", "/sub/private-token-do-not-log"+tc.query, nil)
		req.Header.Set("User-Agent", tc.ua)
		w := httptest.NewRecorder()
		r.subscription(w, req)
		if w.Code != 200 {
			t.Fatalf("%+v status %d", tc, w.Code)
		}
		contentType := w.Header().Get("Content-Type")
		body := w.Body.Bytes()
		switch tc.format {
		case "singbox":
			if !strings.HasPrefix(contentType, "application/json") {
				t.Fatal(contentType)
			}
			if err := validateSingBoxSchema(body); err != nil {
				t.Fatal(err)
			}
			var config struct {
				Outbounds []json.RawMessage `json:"outbounds"`
			}
			if err := json.Unmarshal(body, &config); err != nil {
				t.Fatal(err)
			}
			if len(config.Outbounds) != 3 {
				t.Fatal("visible nodes + selector + direct count incorrect")
			}
		case "clash":
			if !strings.HasPrefix(contentType, "text/yaml") || !strings.Contains(string(body), "proxies:") {
				t.Fatal("invalid YAML response")
			}
			if count := clashProxyEntryCount(string(body)); count != 1 {
				t.Fatalf("unexpected YAML proxies: %d", count)
			}
		case "v2rayn":
			decoded, err := base64.StdEncoding.DecodeString(string(body))
			if err != nil || !strings.HasPrefix(string(decoded), "vless://") {
				t.Fatal("invalid base64 subscription")
			}
		case "raw":
			if !strings.HasPrefix(string(body), "vless://") {
				t.Fatal("raw compatibility lost")
			}
		}
	}
	for _, log := range s.Data.OperationLogs {
		raw, _ := json.Marshal(log)
		if strings.Contains(string(raw), "private-token-do-not-log") {
			t.Fatal("subscription token logged")
		}
	}
}

// clashProxyEntryCount checks the stable structural contract needed here
// without spawning Python/PyYAML inside the builder image. Full YAML parsing is
// deliberately left to the client compatibility matrix, while this test keeps
// the API test hermetic and available in a minimal Go container.
func clashProxyEntryCount(raw string) int {
	inProxies := false
	count := 0
	for _, line := range strings.Split(raw, "\n") {
		if line == "proxies:" {
			inProxies = true
			continue
		}
		if inProxies && line != "" && !strings.HasPrefix(line, " ") {
			break
		}
		if inProxies && strings.HasPrefix(line, "  - name: ") {
			count++
		}
	}
	if !inProxies {
		return -1
	}
	return count
}

// Typed structural comparison for the generated subset of the official 1.12+
// schema. This does not replace sing-box check or actual client/device tests.
func validateSingBoxSchema(raw []byte) error {
	type tls struct {
		Enabled    bool   `json:"enabled"`
		ServerName string `json:"server_name"`
		UTLS       *struct {
			Enabled     bool   `json:"enabled"`
			Fingerprint string `json:"fingerprint"`
		} `json:"utls"`
		Reality *struct {
			Enabled   bool   `json:"enabled"`
			PublicKey string `json:"public_key"`
			ShortID   string `json:"short_id"`
		} `json:"reality"`
	}
	type outbound struct {
		Type           string `json:"type"`
		Tag            string `json:"tag"`
		Server         string `json:"server"`
		ServerPort     int    `json:"server_port"`
		UUID           string `json:"uuid"`
		Password       string `json:"password"`
		Security       string `json:"security"`
		PacketEncoding string `json:"packet_encoding"`
		TLS            *tls   `json:"tls"`
		Transport      *struct {
			Type        string `json:"type"`
			Path        string `json:"path"`
			ServiceName string `json:"service_name"`
		} `json:"transport"`
		Outbounds []string `json:"outbounds"`
		Default   string   `json:"default"`
	}
	var config struct {
		Log struct {
			Level string `json:"level"`
		} `json:"log"`
		DNS struct {
			Servers []struct {
				Type   string `json:"type"`
				Tag    string `json:"tag"`
				Server string `json:"server"`
			} `json:"servers"`
			Final    string `json:"final"`
			Strategy string `json:"strategy"`
		} `json:"dns"`
		Inbounds []struct {
			Type        string   `json:"type"`
			Tag         string   `json:"tag"`
			Listen      string   `json:"listen"`
			ListenPort  int      `json:"listen_port"`
			Address     []string `json:"address"`
			MTU         int      `json:"mtu"`
			AutoRoute   bool     `json:"auto_route"`
			StrictRoute bool     `json:"strict_route"`
			Stack       string   `json:"stack"`
		} `json:"inbounds"`
		Outbounds []outbound `json:"outbounds"`
		Route     struct {
			Final      string `json:"final"`
			Resolver   string `json:"default_domain_resolver"`
			AutoDetect bool   `json:"auto_detect_interface"`
			Rules      []struct {
				Port   int    `json:"port"`
				Action string `json:"action"`
			} `json:"rules"`
		} `json:"route"`
	}
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&config); err != nil {
		return err
	}
	dns := map[string]bool{}
	for _, server := range config.DNS.Servers {
		if server.Type != "udp" || server.Tag == "" || net.ParseIP(server.Server) == nil || dns[server.Tag] {
			return fmt.Errorf("invalid modern UDP DNS server")
		}
		dns[server.Tag] = true
	}
	if !dns[config.DNS.Final] || !dns[config.Route.Resolver] {
		return fmt.Errorf("unresolved DNS reference")
	}
	if config.DNS.Strategy != "ipv4_only" {
		return fmt.Errorf("unexpected DNS strategy")
	}
	mixed := 0
	for _, inbound := range config.Inbounds {
		switch inbound.Type {
		case "mixed":
			if inbound.Listen != "127.0.0.1" || inbound.ListenPort < 1 {
				return fmt.Errorf("invalid local mixed inbound")
			}
			mixed++
		case "tun":
			if !inbound.AutoRoute || !inbound.StrictRoute || inbound.MTU < 1280 || inbound.MTU > 1500 || inbound.Stack != "system" || !config.Route.AutoDetect {
				return fmt.Errorf("invalid TUN profile")
			}
			if len(inbound.Address) == 0 {
				return fmt.Errorf("missing TUN address")
			}
			for _, address := range inbound.Address {
				if _, _, err := net.ParseCIDR(address); err != nil {
					return err
				}
			}
			if len(config.Route.Rules) != 1 || config.Route.Rules[0].Port != 53 || config.Route.Rules[0].Action != "hijack-dns" {
				return fmt.Errorf("missing TUN DNS hijack")
			}
		default:
			return fmt.Errorf("unsupported inbound")
		}
	}
	if mixed != 1 {
		return fmt.Errorf("missing/duplicate local mixed inbound")
	}
	tags := map[string]bool{}
	for _, out := range config.Outbounds {
		if out.Tag == "" || tags[out.Tag] {
			return fmt.Errorf("duplicate outbound tag")
		}
		tags[out.Tag] = true
		switch out.Type {
		case "direct", "selector":
		case "vless", "vmess", "trojan":
			if out.Server == "" || out.ServerPort < 1 || out.ServerPort > 65535 {
				return fmt.Errorf("invalid remote endpoint")
			}
			if out.Type == "trojan" {
				if out.Password == "" || out.TLS == nil {
					return fmt.Errorf("missing Trojan credentials/TLS")
				}
			} else if out.UUID == "" {
				return fmt.Errorf("missing UUID")
			}
			if out.Type == "vmess" && out.Security != "auto" {
				return fmt.Errorf("invalid VMess security")
			}
			if out.TLS != nil {
				if !out.TLS.Enabled || out.TLS.UTLS == nil || !out.TLS.UTLS.Enabled {
					return fmt.Errorf("invalid TLS")
				}
				if reality := out.TLS.Reality; reality != nil && reality.Enabled {
					key, err := base64.RawURLEncoding.DecodeString(reality.PublicKey)
					if err != nil || len(key) != 32 {
						return fmt.Errorf("invalid Reality public key")
					}
				}
			}
			if transport := out.Transport; transport != nil {
				switch transport.Type {
				case "ws":
					if transport.Path == "" {
						return fmt.Errorf("empty WS path")
					}
				case "grpc":
					if transport.ServiceName == "" {
						return fmt.Errorf("empty gRPC service")
					}
				default:
					return fmt.Errorf("unsupported transport")
				}
			}
		default:
			return fmt.Errorf("unsupported outbound type")
		}
	}
	if !tags[config.Route.Final] {
		return fmt.Errorf("unresolved route final")
	}
	for _, out := range config.Outbounds {
		if out.Type == "selector" {
			choices := map[string]bool{}
			for _, tag := range out.Outbounds {
				if !tags[tag] || tag == out.Tag {
					return fmt.Errorf("invalid selector reference")
				}
				choices[tag] = true
			}
			if !choices[out.Default] {
				return fmt.Errorf("default is not a selector choice")
			}
		}
	}
	return nil
}

func TestSingBoxDNSUsesModernSchema(t *testing.T) {
	nodes, client := clientFixture(t)
	for _, input := range [][]model.Node{nil, nodes} {
		config := buildSingBoxSubscription(input, nil, client, model.NetworkPolicy{SingBoxIncludeQuad9: true})
		raw, err := json.Marshal(config)
		if err != nil {
			t.Fatal(err)
		}
		if err = validateSingBoxSchema(raw); err != nil {
			t.Fatal(err)
		}
		dns := config["dns"].(map[string]any)
		servers := dns["servers"].([]any)
		for _, server := range servers {
			entry := server.(map[string]any)
			if entry["type"] == nil || entry["address"] != nil {
				t.Fatal("legacy DNS generated")
			}
		}
		servers[0] = map[string]any{"tag": "cloudflare", "address": "1.1.1.1"}
		legacy, err := json.Marshal(config)
		if err != nil {
			t.Fatal(err)
		}
		if validateSingBoxSchema(legacy) == nil {
			t.Fatal("legacy DNS schema accepted")
		}
	}
}

func TestMobileURIPasswordAndIPv6RoundTrip(t *testing.T) {
	client := model.Client{UUID: fixtureUUID, Password: `&=+,'" /中文😀:@`}
	node := model.Node{Protocol: "trojan", Host: "2001:db8::1", Port: 23456, Name: "测试 节点 #1", Transport: "grpc", Security: "tls"}
	raw := buildNodeShareLink(node, client)
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Hostname() != node.Host || parsed.User.Username() != client.Password || parsed.Fragment != nodeShareName(node) {
		t.Fatalf("round-trip failed: %s", raw)
	}
	if !strings.Contains(raw, "%26") || !strings.Contains(raw, "%3D") {
		t.Fatal("userinfo was not strictly percent encoded")
	}
	if parsed.Query().Get("serviceName") != "yuchen" {
		t.Fatal("gRPC default differs from core")
	}
	node.Protocol = "vless"
	parsed, err = url.Parse(buildNodeShareLink(node, client))
	if err != nil || parsed.Hostname() != node.Host || parsed.User.Username() != client.UUID {
		t.Fatal("VLESS IPv6 round-trip failed")
	}
	node.Protocol = "vmess"
	encoded := strings.TrimPrefix(buildNodeShareLink(node, client), "vmess://")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	var vmess map[string]any
	if err = json.Unmarshal(decoded, &vmess); err != nil || vmess["net"] != "grpc" || vmess["path"] != "yuchen" {
		t.Fatal("VMess gRPC default differs from core")
	}
	yaml := strings.Join(buildClashProxyLinesForNode(node, client), "\n")
	if !strings.Contains(yaml, `network: "grpc"`) || !strings.Contains(yaml, `grpc-service-name: "yuchen"`) {
		t.Fatal("Clash VMess gRPC transport omitted")
	}
}

func TestNodeShareNameMatchesFrontendFixture(t *testing.T) {
	// frontend/tests/share.test.mjs uses the same node fixture and expectation.
	node := model.Node{ID: "node-de00", Name: "重复名称", Host: "192.0.2.1", Port: 24401}
	if got, want := nodeShareName(node), "重复名称 #de00"; got != want {
		t.Fatalf("nodeShareName = %q, want %q", got, want)
	}
}

func TestDisplayNameHasOneShortIDSuffixAcrossFormats(t *testing.T) {
	node := model.Node{ID: "node-de00", Name: "重复名称", Protocol: "vless", Host: "192.0.2.1", Port: 24401, Transport: "tcp", Security: "none", Enabled: true}
	client := model.Client{UUID: fixtureUUID, Password: "password"}
	want := "重复名称 #de00"
	if got := nodeShareName(node); got != want || strings.Count(got, "#de00") != 1 {
		t.Fatalf("node name = %q, want one %q suffix", got, want)
	}
	config := buildSingBoxSubscription([]model.Node{node}, nil, client, model.NetworkPolicy{})
	for _, raw := range config["outbounds"].([]any) {
		out := raw.(map[string]any)
		if out["type"] == "vless" && out["tag"] != want {
			t.Fatalf("sing-box tag = %q, want %q", out["tag"], want)
		}
	}
	clash := strings.Join(buildClashProxyLinesForNode(node, client), "\n")
	if strings.Count(clash, want) != 1 {
		t.Fatalf("Clash name count = %d in %q", strings.Count(clash, want), clash)
	}
	link := buildNodeShareLink(node, client)
	parsed, err := url.Parse(link)
	if err != nil || parsed.Fragment != want || strings.Count(parsed.Fragment, "#") != 1 {
		t.Fatalf("share fragment = %q, want %q", parsed.Fragment, want)
	}
	node.Protocol = "vmess"
	encoded := strings.TrimPrefix(buildNodeShareLink(node, client), "vmess://")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	var vmess map[string]any
	if err := json.Unmarshal(decoded, &vmess); err != nil || vmess["ps"] != want {
		t.Fatalf("v2rayN ps = %v, want %q", vmess["ps"], want)
	}
}

func TestSubscriptionNamesAreUniqueAndTraceable(t *testing.T) {
	nodes := make([]model.Node, 12)
	for i := range nodes {
		nodes[i] = model.Node{ID: fmt.Sprintf("node-%02d", i), Name: "重复名称", Protocol: "vless", Host: "192.0.2.1", Port: 20000 + i, Enabled: true}
	}
	client := model.Client{UUID: fixtureUUID}
	config := buildSingBoxSubscription(nodes, nil, client, model.NetworkPolicy{})
	seen := map[string]bool{}
	for _, raw := range config["outbounds"].([]any) {
		out := raw.(map[string]any)
		if out["type"] == "selector" || out["type"] == "direct" {
			continue
		}
		tag := out["tag"].(string)
		if seen[tag] {
			t.Fatalf("duplicate tag %q", tag)
		}
		seen[tag] = true
		matched := false
		for _, node := range nodes {
			short := strings.Map(func(r rune) rune {
				if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
					return r
				}
				return -1
			}, node.ID)
			if len(short) > 4 {
				short = short[len(short)-4:]
			}
			if strings.HasSuffix(tag, "#"+strings.ToLower(short)) {
				matched = true
				break
			}
		}
		if !matched {
			t.Fatalf("tag not traceable: %s", tag)
		}
	}
}

func clientFixture(t *testing.T) ([]model.Node, model.Client) {
	t.Helper()
	_, public, err := xray.NewRealityKeys()
	if err != nil {
		t.Fatal(err)
	}
	nodes := []model.Node{}
	for i := 0; i < 12; i++ {
		node := model.Node{ID: fmt.Sprintf("node-%02d", i), Name: "重复名称", Protocol: "vless", Host: "192.0.2.1", Port: 23456 + i, Transport: "tcp", Security: "none", Enabled: true}
		switch i {
		case 1:
			node.Security = "tls"
			node.SNI = "example.com"
		case 2:
			node.Security = "reality"
			node.SNI = "www.intel.com"
			node.RealityPublicKey = public
			node.RealityShortID = "abcd1234"
		case 3:
			node.Transport = "grpc"
		case 4:
			node.Protocol = "vmess"
		case 5:
			node.Protocol = "trojan"
			node.Security = "tls"
			node.Host = "2001:db8::1"
			node.SNI = "example.com"
		case 6:
			node.Transport = "ws"
			node.Path = "/yuchen"
		}
		nodes = append(nodes, node)
	}
	client := model.Client{ID: "fixture-client", Username: "fixture", UUID: fixtureUUID, Password: "p@ss w#rd&test=1", Enabled: true, SubscribeToken: "audit-demo-local"}
	for _, node := range nodes {
		client.NodeIDs = append(client.NodeIDs, node.ID)
	}
	return nodes, client
}

func TestExportClientFixtures(t *testing.T) {
	output := os.Getenv("YC_CLIENT_FIXTURES_DIR")
	if output == "" {
		t.Skip("fixture export not requested")
	}
	if err := os.MkdirAll(output, 0700); err != nil {
		t.Fatal(err)
	}
	nodes, client := clientFixture(t)
	data := model.PanelData{Nodes: map[string]model.Node{}, Clients: map[string]model.Client{client.ID: client}, Admins: map[string]model.AdminUser{}, OperationLogs: map[string]model.OperationLog{}}
	hash, err := security.HashPassword("fixture-password-long")
	if err != nil {
		t.Fatal(err)
	}
	data.Admins["fixture-admin"] = model.AdminUser{ID: "fixture-admin", Username: "fixture-admin", PasswordHash: hash, Role: "super_admin", Enabled: true}
	for _, node := range nodes {
		data.Nodes[node.ID] = node
	}
	legacyNode := model.Node{ID: "legacy-ss", Name: "旧节点迁移示例", Protocol: "shadowsocks", Port: 25000, Enabled: true}
	data.Nodes[legacyNode.ID] = legacyNode
	files := map[string]any{"fixture-db.json": data, "singbox.json": buildSingBoxSubscription(nodes, nil, client, model.NetworkPolicy{})}
	for name, value := range files {
		raw, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(output, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
