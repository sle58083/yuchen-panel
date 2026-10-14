package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"yuchen-panel/backend/internal/model"
	"yuchen-panel/backend/internal/security"
	"yuchen-panel/backend/internal/store"
)

func TestNewRouterRejectsWeakSecret(t *testing.T) {
	for _, secret := range []string{"", "short", "dev-secret-change-me", "change-me-please", "please-change-this-secret-before-public-deploy", strings.Repeat(" ", 40)} {
		t.Run(secret, func(t *testing.T) {
			t.Setenv("YC_JWT_SECRET", secret)
			defer func() {
				if recover() == nil {
					t.Error("expected fail-fast panic")
				}
			}()
			NewRouter(&store.Store{})
		})
	}
	t.Setenv("YC_JWT_SECRET", strings.Repeat("s", 64))
	if NewRouter(&store.Store{}) == nil {
		t.Fatal("valid secret rejected")
	}
}

func TestWithAuthRejectsUnknownOrDisabledAccount(t *testing.T) {
	secret := strings.Repeat("s", 64)
	for _, tc := range []struct {
		name    string
		account model.AdminUser
		present bool
		status  int
	}{
		{"unknown", model.AdminUser{}, false, 403},
		{"disabled", model.AdminUser{Enabled: false, Role: "super_admin"}, true, 403},
		{"wrong-role", model.AdminUser{Enabled: true, Role: "viewer"}, true, 403},
		{"valid", model.AdminUser{Username: "database-name", Enabled: true, Role: "super_admin"}, true, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &store.Store{Data: model.PanelData{Admins: map[string]model.AdminUser{}}}
			if tc.present {
				s.Data.Admins["account"] = tc.account
			}
			r := &Router{store: s, jwtSecret: secret}
			token, err := security.SignJWT(secret, security.Claims{Sub: "account", Username: "forged-name", Role: "forged-role", Exp: time.Now().Add(time.Hour).Unix()})
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			r.withAuth(func(w http.ResponseWriter, req *http.Request) {
				claims := currentClaims(req)
				if claims.Username != "database-name" || claims.Role != "super_admin" {
					t.Error("token identity was trusted")
				}
				w.WriteHeader(200)
			})(response, req)
			if response.Code != tc.status {
				t.Fatalf("status %d, want %d", response.Code, tc.status)
			}
		})
	}
}
