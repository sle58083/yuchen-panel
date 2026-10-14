package api

import (
	"math"
	"time"

	"yuchen-panel/backend/internal/model"
)

func migrateTraffic(c *model.Client) {
	if c.TrafficUsedBytes == 0 && c.TrafficUsedGB > 0 && c.TrafficUsedGB <= math.MaxInt64>>30 {
		c.TrafficUsedBytes = c.TrafficUsedGB << 30
	}
}

func addTraffic(a, b int64) int64 {
	if b > math.MaxInt64-a {
		return math.MaxInt64
	}
	return a + b
}

// Authorization is scoped to the authenticated source server, never to a
// username or machine-wide network counter.
func (r *Router) trafficSourceAllowed(c model.Client, serverID string) bool {
	for _, id := range c.NodeIDs {
		if n, ok := r.store.Data.Nodes[id]; ok && n.ServerID == serverID {
			return true
		}
	}
	for _, id := range c.RelayRouteIDs {
		if rr, ok := r.store.Data.RelayRoutes[id]; ok && rr.RelayServerID == serverID {
			return true
		}
	}
	return false
}

func applyTrafficSample(c *model.Client, source string, sample model.ClientTraffic, epoch string, at time.Time) {
	if epoch == "" || at.IsZero() || sample.Uplink < 0 || sample.Downlink < 0 {
		return
	}
	old, exists := c.TrafficCursors[source]
	if exists && !at.After(old.SampleAt) {
		return
	}
	if !c.TrafficResetAt.IsZero() && !at.After(c.TrafficResetAt) {
		return
	}
	migrateTraffic(c)
	u, d := sample.Uplink, sample.Downlink
	if exists && old.Epoch == epoch {
		// A decrease inside one process is a counter reset. Subsequent samples
		// use the new watermark, so repeating the same sample adds zero.
		if u >= old.Uplink {
			u -= old.Uplink
		}
		if d >= old.Downlink {
			d -= old.Downlink
		}
	}
	// With no post-renewal baseline for a source, retain its first absolute
	// sample as baseline rather than charging the previous cycle again.
	if !exists && !c.TrafficResetAt.IsZero() {
		u, d = 0, 0
	}
	c.TrafficUsedBytes = addTraffic(c.TrafficUsedBytes, addTraffic(u, d))
	c.TrafficUsedGB = c.TrafficUsedBytes >> 30 // legacy integer display field
	copyCursors := make(map[string]model.TrafficCursor, len(c.TrafficCursors)+1)
	for id, cursor := range c.TrafficCursors {
		copyCursors[id] = cursor
	}
	copyCursors[source] = model.TrafficCursor{Epoch: epoch, Uplink: sample.Uplink, Downlink: sample.Downlink, SampleAt: at}
	c.TrafficCursors = copyCursors
}

func (r *Router) enforceBillingLocked(now time.Time) bool {
	changed := false
	for id, c := range r.store.Data.Clients {
		migrateTraffic(&c)
		expired := !c.ExpireAt.IsZero() && !now.Before(c.ExpireAt)
		limited := c.TrafficLimitGB > 0 && (c.TrafficLimitGB <= math.MaxInt64>>30 && c.TrafficUsedBytes >= c.TrafficLimitGB<<30)
		if c.Enabled && (expired || limited) {
			c.Enabled = false
			c.UpdatedAt = now
			changed = true
		}
		r.store.Data.Clients[id] = c
	}
	return changed
}

func (r *Router) accountTrafficLocked(hb model.AgentHeartbeat) {
	for id, sample := range hb.TrafficByClient {
		c, ok := r.store.Data.Clients[id]
		if !ok || !r.trafficSourceAllowed(c, hb.ServerID) {
			continue
		}
		applyTrafficSample(&c, hb.ServerID, sample, hb.TrafficEpoch, hb.TrafficSampleAt)
		r.store.Data.Clients[id] = c
	}
	r.enforceBillingLocked(time.Now())
}

// Called only after a PUT has been validated. Accounting is server-owned.
func preserveAndRenewBilling(body *model.Client, old model.Client, now time.Time) {
	migrateTraffic(&old)
	body.TrafficUsedBytes, body.TrafficUsedGB = old.TrafficUsedBytes, old.TrafficUsedGB
	body.TrafficCursors = old.TrafficCursors
	body.TrafficResetAt = old.TrafficResetAt
	body.BillingCycleStart = old.BillingCycleStart
	if body.BillingCycleDays == 0 {
		body.BillingCycleDays = old.BillingCycleDays
	}
	if body.BillingCycleDays == 0 {
		body.BillingCycleDays = 30
	}
	if body.ExpireAt.After(old.ExpireAt) {
		body.TrafficUsedBytes, body.TrafficUsedGB = 0, 0
		body.BillingCycleStart = body.ExpireAt.AddDate(0, 0, -body.BillingCycleDays)
		body.TrafficResetAt = now
		body.Enabled = true
	}
}
