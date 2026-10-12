package api

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"yuchen-panel/backend/internal/model"
	"yuchen-panel/backend/internal/store"
)

func authorizationStore(t *testing.T) *store.Store {
	t.Helper()
	return &store.Store{Path: filepath.Join(t.TempDir(), "db.json"), Data: model.PanelData{
		Clients: map[string]model.Client{}, Nodes: map[string]model.Node{}, RelayRoutes: map[string]model.RelayRoute{}, OperationLogs: map[string]model.OperationLog{},
	}}
}

func TestDeletingLastAuthorizationDisablesOnlyAffectedClient(t *testing.T) {
	for _, kind := range []string{"node", "relay"} {
		t.Run(kind, func(t *testing.T) {
			s := authorizationStore(t)
			s.Data.Nodes["node"] = model.Node{ID: "node"}
			s.Data.RelayRoutes["relay"] = model.RelayRoute{ID: "relay"}
			single := model.Client{Enabled: true}
			mixed := model.Client{Enabled: true, NodeIDs: []string{"node"}, RelayRouteIDs: []string{"relay"}}
			if kind == "node" {
				single.NodeIDs = []string{"node"}
			} else {
				single.RelayRouteIDs = []string{"relay"}
			}
			s.Data.Clients["single"] = single
			s.Data.Clients["mixed"] = mixed
			s.Data.Clients["unrelated"] = model.Client{Enabled: true}
			r := &Router{store: s}
			w := httptest.NewRecorder()
			if kind == "node" {
				r.nodeByID(w, httptest.NewRequest("DELETE", "/api/nodes/node", nil))
			} else {
				r.relayByID(w, httptest.NewRequest("DELETE", "/api/relays/relay", nil))
			}
			if w.Code != 200 {
				t.Fatalf("delete status %d: %s", w.Code, w.Body)
			}
			if s.Data.Clients["single"].Enabled {
				t.Fatal("last authorization removal stayed enabled")
			}
			if !s.Data.Clients["mixed"].Enabled || !s.Data.Clients["unrelated"].Enabled {
				t.Fatal("unrelated grant revoked")
			}
		})
	}
}

func TestSubscriptionUsesExplicitGrantsAndRetiresShadowsocks(t *testing.T) {
	s := authorizationStore(t)
	r := &Router{store: s}
	s.Data.Nodes["node"] = model.Node{ID: "node", Protocol: "vless", Host: "example.com", Port: 23456, Enabled: true}
	s.Data.Nodes["old"] = model.Node{ID: "old", Protocol: "shadowsocks", Host: "example.com", Port: 23457, Enabled: true}
	for _, ids := range [][]string{nil, {"node", "old"}} {
		s.Data.Clients["client"] = model.Client{ID: "client", SubscribeToken: "token", Enabled: true, UUID: "uuid-client", NodeIDs: ids}
		w := httptest.NewRecorder()
		r.subscription(w, httptest.NewRequest("GET", "/sub/token", nil))
		if w.Code != 200 {
			t.Fatal(w.Code)
		}
		if strings.HasPrefix(w.Body.String(), "ss://") || strings.Contains(w.Body.String(), "\nss://") {
			t.Fatal("retired protocol in subscription")
		}
		if strings.Contains(w.Body.String(), "uuid-client") != (len(ids) > 0) {
			t.Fatal("subscription grant differs from core config")
		}
	}
	// Mixed node/relay grants must permit the explicitly granted relay short link.
	s.Data.RelayRoutes["relay"] = model.RelayRoute{ID: "relay", Enabled: true, RouteMode: "socks5_route", RelayHost: "example.com", RelayPort: 23458}
	c := s.Data.Clients["client"]
	c.RelayRouteIDs = []string{"relay"}
	s.Data.Clients["client"] = c
	w := httptest.NewRecorder()
	r.shortShare(w, httptest.NewRequest("GET", "/s/token/relay", nil))
	if w.Code != 200 {
		t.Fatalf("mixed relay denied: %d", w.Code)
	}
}

func TestNewShadowsocksNodeIsRejected(t *testing.T) {
	s := authorizationStore(t)
	s.Data.Servers = map[string]model.Server{"server": {ID: "server", IP: "127.0.0.1", Host: "127.0.0.1"}}
	r := &Router{store: s}
	w := httptest.NewRecorder()
	r.nodes(w, httptest.NewRequest("POST", "/api/nodes", strings.NewReader(`{"name":"legacy","protocol":"shadowsocks","server_id":"server","port":23456}`)))
	if w.Code != 400 || !strings.Contains(w.Body.String(), "面板已停用 Shadowsocks") {
		t.Fatalf("status=%d body=%s", w.Code, w.Body)
	}
}
