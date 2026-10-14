package api

import (
	"strings"
	"yuchen-panel/backend/internal/model"
	"yuchen-panel/backend/internal/xray"
)

func grpcService(path string) string { return valueOr(strings.Trim(path, "/"), "yuchen") }

// Optional VPN profile. Exporting it changes no host routes; the client must
// explicitly grant VPN/admin permissions when it actually runs the config.
func addSingBoxTunProfile(config map[string]any) {
	config["inbounds"] = append([]any{map[string]any{"type": "tun", "tag": "tun-in", "address": []string{"172.19.0.1/30", "fdfe:dcba:9876::1/126"}, "mtu": 1400, "auto_route": true, "strict_route": true, "stack": "system"}}, config["inbounds"].([]any)...)
	route := config["route"].(map[string]any)
	route["auto_detect_interface"] = true
	route["rules"] = []any{map[string]any{"port": 53, "action": "hijack-dns"}}
}

// Minimum schema: sing-box 1.12. No legacy DNS server fields or multiplex defaults.
func buildSingBoxSubscription(nodes []model.Node, relays []model.RelayRoute, client model.Client, policy model.NetworkPolicy) map[string]any {
	visible := []model.Node{}
	for _, node := range nodes {
		if node.Enabled && isShareableNodeProtocol(node.Protocol) {
			visible = append(visible, node)
		}
	}
	activeRelays := []model.RelayRoute{}
	for _, relay := range relays {
		if relay.Enabled && relay.RouteMode == "socks5_route" {
			activeRelays = append(activeRelays, relay)
		}
	}
	visible, activeRelays = uniqueSubscriptionNames(visible, activeRelays)
	outbounds := []any{}
	tags := []string{}
	add := func(node model.Node) {
		outbound := singBoxNode(node, client)
		outbounds = append(outbounds, outbound)
		tags = append(tags, node.Name)
	}
	for _, node := range visible {
		add(node)
	}
	for _, relay := range activeRelays {
		add(model.Node{Name: relay.Name, Protocol: "vless", Host: valueOr(relay.RelayHost, "127.0.0.1"), Port: relay.RelayPort, Transport: "tcp", Security: "reality", SNI: valueOr(relay.RelaySNI, "www.intel.com"), Fingerprint: valueOr(relay.RelayFingerprint, "chrome"), RealityPublicKey: relay.RelayRealityPublicKey, RealityShortID: relay.RelayRealityShortID})
	}
	choices := append(append([]string(nil), tags...), "direct")
	selected := "direct"
	if len(tags) > 0 {
		selected = tags[0]
	}
	selector := map[string]any{"type": "selector", "tag": "PROXY", "outbounds": choices, "default": selected}
	outbounds = append([]any{selector}, outbounds...)
	outbounds = append(outbounds, map[string]any{"type": "direct", "tag": "direct"})
	servers := []any{map[string]any{"type": "udp", "tag": "cloudflare", "server": "1.1.1.1"}, map[string]any{"type": "udp", "tag": "google", "server": "8.8.8.8"}}
	if policy.SingBoxIncludeQuad9 {
		servers = append(servers, map[string]any{"type": "udp", "tag": "quad9", "server": "9.9.9.9"})
	}
	return map[string]any{
		"log":       map[string]any{"level": "info"},
		"dns":       map[string]any{"servers": servers, "final": "cloudflare", "strategy": "ipv4_only"},
		"inbounds":  []any{map[string]any{"type": "mixed", "tag": "mixed-in", "listen": "127.0.0.1", "listen_port": 2080}},
		"outbounds": outbounds,
		"route":     map[string]any{"final": "PROXY", "default_domain_resolver": "cloudflare"},
	}
}

func singBoxNode(node model.Node, client model.Client) map[string]any {
	protocol := strings.ToLower(strings.TrimSpace(node.Protocol))
	out := map[string]any{"type": protocol, "tag": node.Name, "server": strings.Trim(nodeHost(node), "[]"), "server_port": node.Port}
	if protocol == "trojan" {
		out["password"] = xray.TrojanClientPassword(client)
	} else {
		out["uuid"] = client.UUID
		out["packet_encoding"] = "xudp"
		if protocol == "vmess" {
			out["security"] = "auto"
		}
	}
	security := strings.ToLower(node.Security)
	if security == "tls" || security == "reality" || protocol == "trojan" {
		tls := map[string]any{"enabled": true, "server_name": valueOr(node.SNI, nodeHost(node)), "utls": map[string]any{"enabled": true, "fingerprint": valueOr(node.Fingerprint, "chrome")}}
		if security == "reality" {
			tls["reality"] = map[string]any{"enabled": true, "public_key": node.RealityPublicKey, "short_id": node.RealityShortID}
		}
		out["tls"] = tls
	}
	switch strings.ToLower(node.Transport) {
	case "ws":
		out["transport"] = map[string]any{"type": "ws", "path": valueOr(node.Path, "/yuchen")}
	case "grpc":
		out["transport"] = map[string]any{"type": "grpc", "service_name": grpcService(node.Path)}
	}
	return out
}
