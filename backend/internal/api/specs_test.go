package api

import (
	"net/http/httptest"
	"strings"
	"testing"
	"yuchen-panel/backend/internal/model"
	"yuchen-panel/backend/internal/store"
)

func TestAgentSpecsRoundtripAndOlderHeartbeat(t *testing.T) {
	t.Setenv("YC_ADMIN_USERNAME", "specs-admin")
	t.Setenv("YC_ADMIN_PASSWORD", "specs-password-123")
	s, err := store.Open(t.TempDir() + "/panel.json")
	if err != nil {
		t.Fatal(err)
	}
	s.Data.Servers = map[string]model.Server{"selected": {ID: "selected", AgentToken: "fixture-token"}, "unknown": {ID: "unknown"}}
	r := &Router{store: s}
	for _, body := range []string{
		`{"server_id":"selected","cpu_cores":2,"memory_total_bytes":1007681536}`,
		`{"server_id":"selected","cpu_usage":12,"memory_usage":25}`,
	} {
		request := httptest.NewRequest("POST", "/api/agent/heartbeat", strings.NewReader(body))
		request.Header.Set("X-Agent-Token", "fixture-token")
		w := httptest.NewRecorder()
		r.agentHeartbeat(w, request)
		if w.Code != 200 {
			t.Fatalf("heartbeat status=%d body=%s", w.Code, w.Body.String())
		}
	}
	if s.Data.Servers["selected"].CPUCores != 2 || s.Data.Servers["selected"].MemTotalBytes != 1007681536 {
		t.Fatal("old heartbeat erased capacities")
	}
	specs := r.serverSpecs()
	if specs["selected"]["cpu_cores"] != 2 || specs["selected"]["memory_total_bytes"] != int64(1007681536) {
		t.Fatal("system/checks specs differ from Agent values")
	}
	if _, ok := specs["unknown"]; ok {
		t.Fatal("fabricated unknown server capacity")
	}
}
