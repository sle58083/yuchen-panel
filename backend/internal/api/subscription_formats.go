package api

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"yuchen-panel/backend/internal/model"
)

func subscriptionFormat(req *http.Request, parts []string) (string, string) {
	value := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("format")))
	source := "explicit"
	if value == "" {
		value = strings.ToLower(strings.TrimSpace(req.URL.Query().Get("target")))
	}
	if value == "" && len(parts) > 1 {
		value = strings.ToLower(strings.TrimSpace(parts[1]))
	}
	if value == "" {
		source = "user-agent"
		ua := strings.ToLower(req.UserAgent())
		switch {
		case strings.Contains(ua, "clash"), strings.Contains(ua, "mihomo"), strings.Contains(ua, "stash"):
			return "clash", source
		case strings.Contains(ua, "sing-box"), strings.Contains(ua, "singbox"), strings.Contains(ua, "sfa"), strings.Contains(ua, "sfi"):
			return "singbox", source
		default:
			return "v2rayn", source
		}
	}
	switch value {
	case "clash", "clash-meta", "mihomo", "yaml", "yml":
		return "clash", source
	case "sing-box", "singbox", "sb", "json":
		return "singbox", source
	case "v2rayn", "v2rayng", "shadowrocket", "base64":
		return "v2rayn", source
	case "raw":
		return "raw", source // Explicit compatibility for previous unencoded subscriptions.
	default:
		return "v2rayn", "unknown-format-fallback"
	}
}

// Copy subscription inputs under the lock; encoding and format logging run after it is released.
func (r *Router) subscriptionSnapshot(token string) (model.Client, []model.Node, []model.RelayRoute, model.NetworkPolicy, int) {
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	var client model.Client
	found := false
	for _, c := range r.store.Data.Clients {
		if c.SubscribeToken == token {
			client = c
			found = true
			break
		}
	}
	if !found {
		return client, nil, nil, model.NetworkPolicy{}, http.StatusNotFound
	}
	if !client.Enabled || (!client.ExpireAt.IsZero() && !time.Now().Before(client.ExpireAt)) {
		return client, nil, nil, model.NetworkPolicy{}, http.StatusForbidden
	}
	allowed := model.AllowedNodes(client)
	nodes := []model.Node{}
	for _, node := range r.store.Data.Nodes {
		if node.Enabled && allowed[node.ID] && isShareableNodeProtocol(node.Protocol) {
			nodes = append(nodes, node)
		}
	}
	relays := []model.RelayRoute{}
	for _, id := range client.RelayRouteIDs {
		if relay, ok := r.store.Data.RelayRoutes[id]; ok && relay.Enabled && relay.RouteMode == "socks5_route" {
			relays = append(relays, relay)
		}
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	sort.Slice(relays, func(i, j int) bool { return relays[i].ID < relays[j].ID })
	return client, nodes, relays, r.store.Data.NetworkPolicy, http.StatusOK
}

func uniqueSubscriptionNames(nodes []model.Node, relays []model.RelayRoute) ([]model.Node, []model.RelayRoute) {
	nodes = append([]model.Node(nil), nodes...)
	relays = append([]model.RelayRoute(nil), relays...)
	used := map[string]bool{"PROXY": true, "direct": true, "DIRECT": true}
	name := func(base string) string {
		next := base
		for n := 2; used[next]; n++ {
			next = base + " #" + strconv.Itoa(n)
		}
		used[next] = true
		return next
	}
	for i := range nodes {
		nodes[i].Name = name(nodeShareName(nodes[i]))
	}
	for i := range relays {
		relays[i].Name = name(valueOr(relays[i].Name, relays[i].ID))
	}
	return nodes, relays
}
