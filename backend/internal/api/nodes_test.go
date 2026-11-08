package api

import (
	"testing"
	"yuchen-panel/backend/internal/model"
	"yuchen-panel/backend/internal/store"
	"yuchen-panel/backend/internal/xray"
)

func TestNodePresetDefaultsAndStoredSecurity(t *testing.T) {
	t.Setenv("YC_ADMIN_USERNAME", "test-administrator")
	t.Setenv("YC_ADMIN_PASSWORD", "test-password-123")
	s, err := store.Open(t.TempDir() + "/panel.json")
	if err != nil {
		t.Fatal(err)
	}
	r := &Router{store: s}
	cases := []struct{ name, protocol, security, id, want string }{
		// User confirmed mainly mainland clients: case 1 intentionally remains Reality.
		{"C24_1_NewVLESS_UserConfirmedRealityDefault", "vless", "", "", "reality"},
		{"C24_2_NewTrojan_TLS", "trojan", "", "", "tls"},
		{"C24_3_NewVMess_None", "vmess", "", "", "none"},
		{"C24_4_StoredRealityPreserved", "vless", "reality", "stored-reality", "reality"},
		{"C24_5_StoredNonePreserved", "vless", "none", "stored-none", "none"},
		{"ExplicitNewVLESSNoneRemainsSelectable", "vless", "none", "", "none"},
		{"ExplicitVMessRealityCompatibility", "vmess", "reality", "", "none"},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s.Mu.Lock()
			defer s.Mu.Unlock()
			n := model.Node{ID: tc.id, Name: tc.name, Protocol: tc.protocol, Security: tc.security, Port: 24001 + i, Enabled: true}
			if tc.id != "" {
				s.Data.Nodes[tc.id] = n
			}
			if err := r.normalizeNodeLocked(&n, tc.id); err != nil {
				t.Fatal(err)
			}
			if n.Security != tc.want {
				t.Fatalf("security=%q want %q", n.Security, tc.want)
			}
			config := xray.GenerateInboundConfig(n, nil)
			in := config["inbounds"].([]any)[0].(map[string]any)
			if got := in["streamSettings"].(map[string]any)["security"]; got != tc.want {
				t.Fatalf("generated security=%v want %q", got, tc.want)
			}
			if tc.want == "reality" && (n.RealityPrivateKey == "" || n.RealityPublicKey == "") {
				t.Fatal("Reality keys missing")
			}
			t.Logf("protocol=%s security=%s generated=%s", n.Protocol, n.Security, tc.want)
		})
	}
}

func TestShadowsocksMethodValidation(t *testing.T) {
	t.Setenv("YC_ADMIN_USERNAME", "test-administrator")
	t.Setenv("YC_ADMIN_PASSWORD", "test-password-123")
	s, err := store.Open(t.TempDir() + "/panel.json")
	if err != nil {
		t.Fatal(err)
	}
	r := &Router{store: s}
	for _, method := range []string{"aes-128-gcm", "aes-256-gcm", "chacha20-ietf-poly1305", "2022-blake3-aes-128-gcm", "2022-blake3-aes-256-gcm"} {
		s.Mu.Lock()
		n := model.Node{Name: method, Protocol: "shadowsocks", SSMethod: method, SSPassword: "password", Port: 25000, Enabled: true}
		err = r.normalizeNodeLocked(&n, "")
		s.Mu.Unlock()
		if err != nil {
			t.Fatalf("method %q rejected: %v", method, err)
		}
	}
	s.Mu.Lock()
	n := model.Node{Name: "invalid", Protocol: "shadowsocks", SSMethod: "rc4-md5", SSPassword: "password", Port: 25001, Enabled: true}
	err = r.normalizeNodeLocked(&n, "")
	s.Mu.Unlock()
	if err == nil {
		t.Fatal("unsupported Shadowsocks method accepted")
	}
}
