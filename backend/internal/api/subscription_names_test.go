package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"yuchen-panel/backend/internal/model"
)

func TestSubscriptionNamesAcrossFourExports(t *testing.T) {
	for _, scenario := range []struct {
		name        string
		input, want []string
	}{
		{"distinct", []string{"美国", "日本", "韩国"}, []string{"美国", "日本", "韩国"}},
		{"duplicates", []string{"美国", "美国", "美国"}, []string{"美国", "美国 2", "美国 3"}},
		{"twelve", make([]string, 12), nil},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			if scenario.name == "twelve" {
				for i := range scenario.input {
					scenario.input[i] = "重复名称"
					scenario.want = append(scenario.want, "重复名称")
					if i > 0 {
						scenario.want[i] += fmt.Sprintf(" %d", i+1)
					}
				}
			}
			nodes := make([]model.Node, len(scenario.input))
			for i, name := range scenario.input {
				nodes[i] = model.Node{ID: fmt.Sprintf("node-%04d", i), Name: name, Protocol: "vless", Host: "192.0.2.1", Port: 20000 + i, Transport: "tcp", Security: "none", Enabled: true}
			}
			original := append([]model.Node(nil), nodes...)
			unique, _ := uniqueSubscriptionNames(nodes, nil)
			if !reflect.DeepEqual(nodes, original) {
				t.Fatal("dedup changed stored node data")
			}
			client := model.Client{UUID: fixtureUUID, Password: "fixture-password"}
			config := buildSingBoxSubscription(nodes, nil, client, model.NetworkPolicy{})
			var tags, names, fragments, ps []string
			for _, raw := range config["outbounds"].([]any) {
				out := raw.(map[string]any)
				if out["type"] == "vless" {
					tags = append(tags, out["tag"].(string))
				}
			}
			clash := buildClashMetaSubscription(unique, nil, client, model.NetworkPolicy{})
			for _, line := range strings.Split(clash, "\n") {
				if strings.HasPrefix(line, "  - name: \"") {
					var name string
					if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "  - name: ")), &name); err != nil {
						t.Fatal(err)
					}
					names = append(names, name)
				}
			}
			for i, node := range unique {
				unchanged := node
				unchanged.Name = original[i].Name
				if unchanged != original[i] {
					t.Fatal("dedup changed fields besides Name")
				}
				link := buildNodeShareLink(node, client)
				u, err := url.Parse(link)
				if err != nil {
					t.Fatal(err)
				}
				fragments = append(fragments, u.Fragment)
				t.Logf("VLESS: %s", link)
				node.Protocol = "trojan"
				u, err = url.Parse(buildNodeShareLink(node, client))
				if err != nil || u.Fragment != scenario.want[i] {
					t.Fatalf("Trojan fragment: %v %v", u, err)
				}
				node.Protocol = "vmess"
				decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(buildNodeShareLink(node, client), "vmess://"))
				if err != nil {
					t.Fatal(err)
				}
				var obj map[string]any
				if err := json.Unmarshal(decoded, &obj); err != nil {
					t.Fatal(err)
				}
				ps = append(ps, obj["ps"].(string))
			}
			for path, got := range map[string][]string{"sing-box tag": tags, "Clash name": names, "fragment": fragments, "VMess ps": ps} {
				if !reflect.DeepEqual(got, scenario.want) {
					t.Fatalf("%s got %q want %q", path, got, scenario.want)
				}
				seen := map[string]bool{}
				for _, name := range got {
					if seen[name] || strings.Contains(name, "#") {
						t.Fatalf("%s duplicate/ID suffix: %q", path, name)
					}
					seen[name] = true
				}
				t.Logf("%s: %q", path, got)
			}
		})
	}
}

func TestNodeShareNameTrimAndEmptyFallback(t *testing.T) {
	for _, sample := range []struct{ name, host, want string }{
		{" 美国 ", "192.0.2.1", "美国"},
		{"", "192.0.2.1", "192.0.2.1:24401"},
		{"   ", "", "127.0.0.1:24401"},
	} {
		if got := nodeShareName(model.Node{ID: "node-de00", Name: sample.name, Host: sample.host, Port: 24401}); got != sample.want {
			t.Fatalf("got %q want %q", got, sample.want)
		}
	}
}
