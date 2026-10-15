package api

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"yuchen-panel/backend/internal/model"
	"yuchen-panel/backend/internal/store"
)

func socksUsernameRouter(t *testing.T) *Router {
	t.Helper()
	t.Setenv("YC_ADMIN_USERNAME", "socks-test-admin")
	t.Setenv("YC_ADMIN_PASSWORD", "socks-test-password")
	s, err := store.Open(t.TempDir() + "/panel.json")
	if err != nil {
		t.Fatal(err)
	}
	s.Data.Servers["s"] = model.Server{ID: "s", Host: "example.com"}
	return &Router{store: s}
}

func createSocksUsernameNode(t *testing.T, r *Router, port int, username string) model.Node {
	t.Helper()
	body := model.Node{Name: "SOCKS test", ServerID: "s", Port: port, Protocol: "socks5", SocksUsername: username, SocksPassword: "existing-test-password"}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r.nodes(w, httptest.NewRequest("POST", "/api/nodes", bytes.NewReader(raw)))
	if w.Code != 201 {
		t.Fatalf("POST %d: %s", w.Code, w.Body.String())
	}
	var node model.Node
	if err := json.Unmarshal(w.Body.Bytes(), &node); err != nil {
		t.Fatal(err)
	}
	return node
}

func TestSocksEmptyUsernameGeneratedAndDistinct(t *testing.T) {
	r := socksUsernameRouter(t)
	first := createSocksUsernameNode(t, r, 24501, "")
	second := createSocksUsernameNode(t, r, 24502, "")
	for _, node := range []model.Node{first, second} {
		if len(node.SocksUsername) != 12 {
			t.Fatalf("username length=%d, want 12", len(node.SocksUsername))
		}
		for _, c := range node.SocksUsername {
			if !strings.ContainsRune("abcdefghijkmnpqrstuvwxyz23456789", c) {
				t.Fatalf("username contains disallowed character %q", c)
			}
		}
		if node.SocksPassword != "existing-test-password" || node.Protocol != "socks" {
			t.Fatal("username generation changed password or SOCKS normalization")
		}
	}
	if first.SocksUsername == second.SocksUsername {
		t.Fatal("two created nodes share the same generated username")
	}
	t.Log("two created SOCKS5 nodes: 12 allowed characters each, distinct usernames")
}

func TestSocksStoredUsernamePreserved(t *testing.T) {
	for _, username := range []string{"yuchen", "custom-account"} {
		t.Run(username, func(t *testing.T) {
			r := socksUsernameRouter(t)
			node := createSocksUsernameNode(t, r, 24501, username)
			if node.SocksUsername != username {
				t.Fatal("creation overwrote explicit username")
			}
			node.Remark = "edit without rotating credentials"
			raw, err := json.Marshal(node)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			r.nodeByID(w, httptest.NewRequest("PUT", "/api/nodes/"+node.ID, bytes.NewReader(raw)))
			if w.Code != 200 {
				t.Fatalf("PUT %d: %s", w.Code, w.Body.String())
			}
			got := r.store.Data.Nodes[node.ID]
			if got.SocksUsername != username || got.SocksPassword != node.SocksPassword {
				t.Fatal("editing rotated existing credentials")
			}
			t.Log("explicit username and password preserved on creation and edit")
		})
	}
}
