package model

import "time"

// AllowedNodes returns only explicit grants for active clients. Empty bindings
// grant nothing, including clients authorized solely for a fixed-exit relay.
func AllowedNodes(client Client) map[string]bool {
	allowed := map[string]bool{}
	if !client.Enabled || (!client.ExpireAt.IsZero() && !time.Now().Before(client.ExpireAt)) {
		return allowed
	}
	for _, id := range client.NodeIDs {
		if id != "" {
			allowed[id] = true
		}
	}
	return allowed
}
