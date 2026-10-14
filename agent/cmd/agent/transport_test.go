package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestAgentTokenIsNotForwardedAcrossOrigins(t *testing.T) {
	var hits atomic.Int32
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits.Add(1) }))
	defer destination.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, destination.URL, 302) }))
	defer origin.Close()
	if err := postJSON(origin.URL, "fixture-token", []byte(`{}`), nil); err == nil {
		t.Fatal("cross-origin redirect allowed")
	}
	if hits.Load() != 0 {
		t.Fatal("agent request reached redirect destination")
	}
}

func TestRemoteAgentRequiresHTTPS(t *testing.T) {
	for _, tc := range []struct {
		url     string
		allowed bool
	}{
		{"http://127.0.0.1:8088", true}, {"http://[::1]:8088", true}, {"https://example.com/api", true},
		{"http://example.com", false}, {"http://192.0.2.1", false}, {"https://user:password@example.com", false}, {"file:///tmp/panel", false},
	} {
		_, err := validatePanelURL(tc.url)
		if (err == nil) != tc.allowed {
			t.Errorf("%s allowed=%v error=%v", tc.url, tc.allowed, err)
		}
	}
}
