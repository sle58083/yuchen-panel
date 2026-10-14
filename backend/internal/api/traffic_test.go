package api

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"yuchen-panel/backend/internal/model"
	"yuchen-panel/backend/internal/store"
	"yuchen-panel/backend/internal/xray"
)

func billingRouter(t *testing.T) *Router {
	t.Helper()
	t.Setenv("YC_ADMIN_USERNAME", "billing-admin")
	t.Setenv("YC_ADMIN_PASSWORD", "billing-test-password")
	s, err := store.Open(t.TempDir() + "/panel.json")
	if err != nil {
		t.Fatal(err)
	}
	s.Data.Servers = map[string]model.Server{"s": {ID: "s", AgentToken: "token"}, "other": {ID: "other", AgentToken: "other-token"}}
	s.Data.Nodes = map[string]model.Node{"n": {ID: "n", ServerID: "s", Protocol: "vless", Port: 24401, Transport: "tcp", Security: "none", Enabled: true}}
	s.Data.Clients = map[string]model.Client{"c": {ID: "c", Username: "same", UUID: "fixture-uuid", Enabled: true, NodeIDs: []string{"n"}, TrafficLimitGB: 1}}
	return &Router{store: s}
}

func TestTrafficPersistenceFailureRollsBackBeforeRetry(t *testing.T) {
	r := billingRouter(t)
	validPath := r.store.Path
	blocked := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocked, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	r.store.Path = filepath.Join(blocked, "panel.json")
	at := time.Now()
	raw, _ := json.Marshal(model.AgentHeartbeat{ServerID: "s", TrafficEpoch: "xray", TrafficSampleAt: at, TrafficByClient: map[string]model.ClientTraffic{"c": {Uplink: 100, Downlink: 200}}})
	q := httptest.NewRequest("POST", "/api/agent/heartbeat", bytes.NewReader(raw))
	q.Header.Set("X-Agent-Token", "token")
	w := httptest.NewRecorder()
	r.agentHeartbeat(w, q)
	if w.Code != 500 || r.store.Data.Clients["c"].TrafficUsedBytes != 0 {
		t.Fatalf("failed save altered ledger: %d %+v", w.Code, r.store.Data.Clients["c"])
	}
	r.store.Path = validPath
	postTraffic(t, r, 100, 200, "xray", at)
	if r.store.Data.Clients["c"].TrafficUsedBytes != 300 {
		t.Fatal("retry lost or doubled sample")
	}
}

func postTraffic(t *testing.T, r *Router, up, down int64, epoch string, at time.Time) {
	t.Helper()
	raw, _ := json.Marshal(model.AgentHeartbeat{ServerID: "s", TrafficEpoch: epoch, TrafficSampleAt: at, TrafficByClient: map[string]model.ClientTraffic{"c": {Uplink: up, Downlink: down}}})
	q := httptest.NewRequest("POST", "/api/agent/heartbeat", bytes.NewReader(raw))
	q.Header.Set("X-Agent-Token", "token")
	w := httptest.NewRecorder()
	r.agentHeartbeat(w, q)
	if w.Code != 200 {
		t.Fatalf("heartbeat %d %s", w.Code, w.Body.String())
	}
}

func TestTrafficNormalBytesAndConversion(t *testing.T) {
	r := billingRouter(t)
	postTraffic(t, r, 4030040, 91196267, "process1", time.Now())
	c := r.store.Data.Clients["c"]
	if c.TrafficUsedBytes != 95226307 || !c.Enabled {
		t.Fatalf("client %+v", c)
	}
	t.Logf("traffic_used_bytes=%d used_GB=%.6f", c.TrafficUsedBytes, float64(c.TrafficUsedBytes)/(1<<30))
}

func TestTrafficRestartAndRepeatedValueAreIdempotent(t *testing.T) {
	r := billingRouter(t)
	start := time.Now()
	postTraffic(t, r, 100, 200, "first", start)
	postTraffic(t, r, 10, 20, "second", start.Add(time.Second))
	if got := r.store.Data.Clients["c"].TrafficUsedBytes; got != 330 {
		t.Fatalf("restart got %d want 330", got)
	}
	postTraffic(t, r, 10, 20, "second", start.Add(2*time.Second))
	if got := r.store.Data.Clients["c"].TrafficUsedBytes; got != 330 {
		t.Fatalf("repeat got %d want 330", got)
	}
	postTraffic(t, r, 20, 30, "second", start.Add(3*time.Second))
	if got := r.store.Data.Clients["c"].TrafficUsedBytes; got != 350 {
		t.Fatalf("increment got %d", got)
	}
	t.Log("restart: 300+30=330; repeat=330; subsequent increment=350")
	postTraffic(t, r, 1000, 1000, "third", start.Add(4*time.Second))
	if got := r.store.Data.Clients["c"].TrafficUsedBytes; got != 2350 {
		t.Fatalf("restart with larger counters got %d", got)
	}
}

func TestTrafficMachineTotalsNeverBillClients(t *testing.T) {
	r := billingRouter(t)
	raw, _ := json.Marshal(model.AgentHeartbeat{ServerID: "s", UploadTotal: 100 << 30, DownloadTotal: 100 << 30})
	q := httptest.NewRequest("POST", "/api/agent/heartbeat", bytes.NewReader(raw))
	q.Header.Set("X-Agent-Token", "token")
	w := httptest.NewRecorder()
	r.agentHeartbeat(w, q)
	if w.Code != 200 || r.store.Data.Clients["c"].TrafficUsedBytes != 0 || !r.store.Data.Clients["c"].Enabled {
		t.Fatal("machine totals billed client")
	}
}

func TestTrafficDuplicateHeartbeatPersistsAcrossPanelRestart(t *testing.T) {
	r := billingRouter(t)
	at := time.Now()
	postTraffic(t, r, 100, 200, "xray", at)
	// Reopen the real JSON store, proving the watermark is durable.
	s, err := store.Open(r.store.Path)
	if err != nil {
		t.Fatal(err)
	}
	r.store = s
	postTraffic(t, r, 100, 200, "xray", at.Add(time.Second))
	if r.store.Data.Clients["c"].TrafficUsedBytes != 300 {
		t.Fatal("duplicate billed after panel restart")
	}
}

func TestTrafficLimitStopsAndSyncRemovesClient(t *testing.T) {
	r := billingRouter(t)
	cfg := xray.GenerateInboundConfig(r.store.Data.Nodes["n"], []model.Client{r.store.Data.Clients["c"]})
	before := xray.ConfigHash(cfg)
	postTraffic(t, r, 1<<29, 1<<29, "xray", time.Now())
	if r.store.Data.Clients["c"].Enabled {
		t.Fatal("limit failed")
	}
	raw, _ := json.Marshal(model.AgentSyncRequest{ServerID: "s", ConfigHash: before})
	q := httptest.NewRequest("POST", "/api/agent/sync", bytes.NewReader(raw))
	q.Header.Set("X-Agent-Token", "token")
	w := httptest.NewRecorder()
	r.agentSync(w, q)
	var response model.AgentSyncResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || !response.RestartRequired {
		t.Fatalf("sync %d %s", w.Code, w.Body.String())
	}
	raw, _ = json.Marshal(response.XrayConfig)
	if bytes.Contains(raw, []byte("fixture-uuid")) {
		t.Fatal("stopped client still configured")
	}
	t.Log("Enabled=false; sync RestartRequired=true; client credential removed")
}

func TestTrafficLimitUnderAndUnlimited(t *testing.T) {
	for _, limit := range []int64{1, 0} {
		r := billingRouter(t)
		c := r.store.Data.Clients["c"]
		c.TrafficLimitGB = limit
		r.store.Data.Clients["c"] = c
		amount := int64(10)
		if limit == 0 {
			amount = 2 << 30
		}
		postTraffic(t, r, amount, 0, "xray", time.Now())
		if !r.store.Data.Clients["c"].Enabled {
			t.Fatalf("limit %d incorrectly stopped", limit)
		}
	}
}

func TestExpireBillingStopsAndSubscriptionGuardRemains(t *testing.T) {
	r := billingRouter(t)
	c := r.store.Data.Clients["c"]
	c.ExpireAt = time.Now().Add(-time.Minute)
	c.SubscribeToken = "expired"
	r.store.Data.Clients["c"] = c
	postTraffic(t, r, 0, 0, "xray", time.Now())
	if r.store.Data.Clients["c"].Enabled {
		t.Fatal("expired not disabled")
	}
	q := httptest.NewRequest("GET", "/sub/expired?format=raw", nil)
	w := httptest.NewRecorder()
	r.subscription(w, q)
	if w.Code != 403 {
		t.Fatalf("expired subscription %d", w.Code)
	}
}

func TestBillingRenewalBoundaryCases(t *testing.T) {
	for _, scenario := range []struct {
		name                             string
		enabled, expired, renew, shorten bool
	}{
		{"Billing11_extend_resets", true, false, true, false},
		{"Billing12_disabled_renew_restores", false, false, true, false},
		{"Billing13_remark_preserves", true, false, false, false},
		{"Billing14_shorter_preserves", true, false, false, true},
		{"Billing15_expired_renew_restores", false, true, true, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			r := billingRouter(t)
			now := time.Now().UTC().Truncate(time.Second)
			c := r.store.Data.Clients["c"]
			c.Enabled = scenario.enabled
			c.ExpireAt = now.Add(24 * time.Hour)
			if scenario.expired {
				c.ExpireAt = now.Add(-24 * time.Hour)
			}
			c.TrafficUsedBytes = 123456
			c.BillingCycleStart = now.Add(-30 * 24 * time.Hour)
			r.store.Data.Clients["c"] = c
			body := c
			body.Remark = "updated only"
			body.TrafficUsedBytes = 0 // submitted accounting must be ignored
			if scenario.renew {
				body.ExpireAt = now.Add(31 * 24 * time.Hour)
			}
			if scenario.shorten {
				body.ExpireAt = c.ExpireAt.Add(-time.Hour)
			}
			raw, _ := json.Marshal(body)
			q := httptest.NewRequest("PUT", "/api/clients/c", bytes.NewReader(raw))
			w := httptest.NewRecorder()
			r.clientByID(w, q)
			if w.Code != 200 {
				t.Fatalf("PUT %d %s", w.Code, w.Body.String())
			}
			got := r.store.Data.Clients["c"]
			if scenario.renew {
				if got.TrafficUsedBytes != 0 || !got.Enabled || !got.BillingCycleStart.Equal(body.ExpireAt.AddDate(0, 0, -30)) {
					t.Fatalf("renew %+v", got)
				}
			} else if got.TrafficUsedBytes != 123456 || !got.BillingCycleStart.Equal(c.BillingCycleStart) {
				t.Fatalf("nonrenew %+v", got)
			}
			t.Logf("used_bytes=%d enabled=%t cycle_start=%s", got.TrafficUsedBytes, got.Enabled, got.BillingCycleStart.Format(time.RFC3339))
		})
	}
}

func TestTrafficMultiServerAndStaleSourceIsolation(t *testing.T) {
	r := billingRouter(t)
	at := time.Now()
	postTraffic(t, r, 100, 0, "one", at)
	c := r.store.Data.Clients["c"]
	c.NodeIDs = append(c.NodeIDs, "n2")
	r.store.Data.Clients["c"] = c
	r.store.Data.Nodes["n2"] = model.Node{ID: "n2", ServerID: "other"}
	r.accountTrafficLocked(model.AgentHeartbeat{ServerID: "other", TrafficEpoch: "two", TrafficSampleAt: at, TrafficByClient: map[string]model.ClientTraffic{"c": {Uplink: 200}}})
	postTraffic(t, r, 999, 0, "old", at.Add(-time.Minute))
	if r.store.Data.Clients["c"].TrafficUsedBytes != 300 {
		t.Fatal("multi-server or stale charge incorrect")
	}
	r.accountTrafficLocked(model.AgentHeartbeat{ServerID: "unbound", TrafficEpoch: "three", TrafficSampleAt: at, TrafficByClient: map[string]model.ClientTraffic{"c": {Uplink: 10000}}})
	if r.store.Data.Clients["c"].TrafficUsedBytes != 300 {
		t.Fatal("unbound server charged client")
	}
}

func TestBillingRenewalSameCumulativeValueDoesNotRebill(t *testing.T) {
	r := billingRouter(t)
	at := time.Now()
	postTraffic(t, r, 100, 200, "xray", at)
	c := r.store.Data.Clients["c"]
	next := c
	next.ExpireAt = at.Add(30 * 24 * time.Hour)
	preserveAndRenewBilling(&next, c, at.Add(time.Second))
	r.store.Data.Clients["c"] = next
	postTraffic(t, r, 100, 200, "xray", at.Add(2*time.Second))
	if r.store.Data.Clients["c"].TrafficUsedBytes != 0 {
		t.Fatal("old cycle rebilled")
	}
	postTraffic(t, r, 150, 250, "xray", at.Add(3*time.Second))
	if r.store.Data.Clients["c"].TrafficUsedBytes != 100 {
		t.Fatal("new cycle increment incorrect")
	}
}
