package api

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"yuchen-panel/backend/internal/model"
)

func TestClientTrafficLimitValidationMessages(t *testing.T) {
	r := billingRouter(t)
	negative := model.Client{Username: "test", TrafficLimitGB: -1}
	overflow := model.Client{Username: "test", TrafficLimitGB: int64(1) << 33}
	negativeErr := r.validateClientLocked(&negative)
	overflowErr := r.validateClientLocked(&overflow)
	if negativeErr == nil || negativeErr.Error() != "流量限制不能小于 0" {
		t.Fatalf("negative limit: %v", negativeErr)
	}
	if overflowErr == nil || overflowErr.Error() != "流量上限超出允许范围" {
		t.Fatalf("overflow limit: %v", overflowErr)
	}
	if negativeErr.Error() == overflowErr.Error() {
		t.Fatal("negative and overflow errors must differ")
	}
	for _, limit := range []int64{0, (int64(1) << 33) - 1} {
		c := model.Client{Username: "test", TrafficLimitGB: limit}
		if err := r.validateClientLocked(&c); err != nil {
			t.Fatalf("valid limit %d: %v", limit, err)
		}
	}
	t.Logf("negative=%s; overflow=%s", negativeErr, overflowErr)
}

func TestBillingExpiryTransitions(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	for _, scenario := range []struct {
		name                 string
		oldExpiry, newExpiry time.Time
		renew                bool
	}{
		{"lifetime_to_expiring_preserves", time.Time{}, now.Add(48 * time.Hour), false},
		{"extension_resets", now.Add(24 * time.Hour), now.Add(48 * time.Hour), true},
		{"shortening_preserves", now.Add(48 * time.Hour), now.Add(24 * time.Hour), false},
		{"expiring_to_lifetime_preserves", now.Add(24 * time.Hour), time.Time{}, false},
	} {
		for _, enabled := range []bool{false, true} {
			name := scenario.name + "/disabled"
			if enabled {
				name = scenario.name + "/enabled"
			}
			t.Run(name, func(t *testing.T) {
				r := billingRouter(t)
				old := r.store.Data.Clients["c"]
				old.Enabled, old.ExpireAt = enabled, scenario.oldExpiry
				old.TrafficUsedBytes, old.TrafficUsedGB = 2<<30, 2
				old.BillingCycleStart = now.Add(-10 * 24 * time.Hour)
				old.TrafficResetAt = now.Add(-time.Hour)
				old.TrafficCursors = map[string]model.TrafficCursor{"s": {Epoch: "original", Uplink: 123, Downlink: 456, SampleAt: now}}
				r.store.Data.Clients["c"] = old
				body := old
				body.ExpireAt = scenario.newExpiry
				// The API must ignore submitted accounting fields.
				body.TrafficUsedBytes, body.TrafficUsedGB = 0, 0
				body.TrafficCursors = nil
				body.TrafficResetAt, body.BillingCycleStart = time.Time{}, time.Time{}
				raw, err := json.Marshal(body)
				if err != nil {
					t.Fatal(err)
				}
				w := httptest.NewRecorder()
				r.clientByID(w, httptest.NewRequest("PUT", "/api/clients/c", bytes.NewReader(raw)))
				if w.Code != 200 {
					t.Fatalf("PUT %d: %s", w.Code, w.Body.String())
				}
				got := r.store.Data.Clients["c"]
				if !got.ExpireAt.Equal(scenario.newExpiry) || !reflect.DeepEqual(got.TrafficCursors, old.TrafficCursors) {
					t.Fatalf("expiry or cursor changed unexpectedly: %+v", got)
				}
				if scenario.renew {
					if got.TrafficUsedBytes != 0 || got.TrafficUsedGB != 0 || !got.Enabled || !got.TrafficResetAt.After(old.TrafficResetAt) || !got.BillingCycleStart.Equal(scenario.newExpiry.AddDate(0, 0, -30)) {
						t.Fatalf("renewal did not reset and enable: %+v", got)
					}
				} else if got.TrafficUsedBytes != old.TrafficUsedBytes || got.TrafficUsedGB != old.TrafficUsedGB || got.Enabled != enabled || !got.TrafficResetAt.Equal(old.TrafficResetAt) || !got.BillingCycleStart.Equal(old.BillingCycleStart) {
					t.Fatalf("nonrenewal altered accounting or enabled state: %+v", got)
				}
				t.Logf("used_bytes=%d enabled=%t renewal=%t", got.TrafficUsedBytes, got.Enabled, scenario.renew)
			})
		}
	}
}
