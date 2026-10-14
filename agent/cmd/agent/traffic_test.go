package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStatsTrafficParserAcceptsProtoStringsAndNumbers(t *testing.T) {
	raw := []byte(`{"stat":[{"name":"user>>>client_id1>>>traffic>>>uplink","value":"4030040"},{"name":"user>>>client_id1>>>traffic>>>downlink","value":91196267},{"name":"user>>>same-username>>>traffic>>>uplink","value":"999"},{"name":"inbound>>>node>>>traffic>>>uplink","value":"999"}]}`)
	got, err := parseTraffic(raw)
	if err != nil || len(got) != 1 || got["id1"].Uplink != 4030040 || got["id1"].Downlink != 91196267 {
		t.Fatalf("%v %v", got, err)
	}
	t.Logf("traffic_by_client=%+v (machine/username counters ignored)", got)
}

func TestStatsDefaultZeroValueInProtoJSON(t *testing.T) {
	got, err := parseTraffic([]byte(`{"stat":[{"name":"user>>>client_a>>>traffic>>>uplink"},{"name":"user>>>client_a>>>traffic>>>downlink","value":"1"}]}`))
	if err != nil || got["a"].Uplink != 0 || got["a"].Downlink != 1 {
		t.Fatal(got, err)
	}
}

func TestStatsCollectorWithRealCore(t *testing.T) {
	binary := os.Getenv("YC_XRAY_TEST_BINARY")
	if binary == "" {
		t.Skip("real core integration requires explicitly staged YC_XRAY_TEST_BINARY")
	}
	binary, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", filepath.Dir(binary)+string(os.PathListSeparator)+os.Getenv("PATH"))
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	path := filepath.Join(t.TempDir(), "stats-only.json")
	// Only a loopback gRPC API, no proxy listener, routing or external node use.
	raw := fmt.Sprintf(`{"log":{"loglevel":"warning"},"stats":{},"policy":{"levels":{"0":{"statsUserUplink":true,"statsUserDownlink":true}}},"api":{"tag":"api","services":["StatsService"]},"inbounds":[{"tag":"api","listen":"127.0.0.1","port":%d,"protocol":"dokodemo-door","settings":{"address":"127.0.0.1"}}],"outbounds":[{"protocol":"blackhole","tag":"blocked"}],"routing":{"rules":[{"type":"field","inboundTag":["api"],"outboundTag":"api"}]}}`, port)
	if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(binary, "run", "-config", path)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { cmd.Process.Kill(); cmd.Wait() }()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		traffic, epoch, err := collectXrayTraffic(Config{ConfigPath: path})
		if err == nil {
			if epoch == "" || len(traffic) != 0 {
				t.Fatal("unexpected stats-only result", traffic, epoch)
			}
			t.Logf("real local StatsService queried on 127.0.0.1:%d; process epoch identified; no proxy traffic generated", port)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("real StatsService integration timeout")
}

func TestStatsEpochSurvivesAgentRestartButDetectsXrayRestart(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "sys/kernel/random"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "4321"), 0700); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(root, "sys/kernel/random/boot_id"), []byte("boot-id\n"), 0600)
	os.WriteFile(filepath.Join(root, "4321/cmdline"), []byte("/usr/local/bin/xray\x00run\x00-config\x00/config.json\x00"), 0600)
	statPath := filepath.Join(root, "4321/stat")
	os.WriteFile(statPath, []byte("4321 (xray) S "+strings.Repeat("0 ", 18)+"101 0"), 0600)
	first, err := processEpoch(root, "/config.json")
	if err != nil {
		t.Fatal(err)
	}
	second, err := processEpoch(root, "/config.json")
	if err != nil || second != first {
		t.Fatal("Agent restart changed epoch", second, err)
	}
	os.WriteFile(statPath, []byte("4321 (xray) S "+strings.Repeat("0 ", 18)+"202 0"), 0600)
	third, err := processEpoch(root, "/config.json")
	if err != nil || third == first {
		t.Fatal("Xray restart missed", third, err)
	}
	if first != "boot-id:4321:101" || third != "boot-id:4321:202" {
		t.Fatal(first, third)
	}
}

func TestStatsTrafficParserRejectsInvalidCounters(t *testing.T) {
	for _, raw := range []string{
		`not JSON`,
		`{"stat":[{"name":"user>>>client_a>>>traffic>>>uplink","value":"-1"}]}`,
		`{"stat":[{"name":"user>>>client_a>>>traffic>>>uplink","value":"9223372036854775808"}]}`,
		`{"stat":[{"name":"user>>>client_a>>>traffic>>>uplink","value":"1"},{"name":"user>>>client_a>>>traffic>>>uplink","value":"1"}]}`,
	} {
		if _, err := parseTraffic([]byte(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}

func TestStatsAPIAddressLoopbackOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	for _, listen := range []string{"127.0.0.1", "0.0.0.0", "::"} {
		os.WriteFile(path, []byte(`{"inbounds":[{"tag":"api","listen":"`+listen+`","port":10001}]}`), 0600)
		address, err := statsAddress(path)
		if listen == "127.0.0.1" {
			if err != nil || address != "127.0.0.1:10001" {
				t.Fatalf("%s %v", address, err)
			}
		} else if err == nil {
			t.Fatal("public stats endpoint accepted")
		}
	}
}

func TestStatsPartialDirectionDoesNotInventCounterReset(t *testing.T) {
	got, err := parseTraffic([]byte(`{"stat":[{"name":"user>>>client_a>>>traffic>>>uplink","value":"100"}]}`))
	if err != nil || len(got) != 0 {
		t.Fatalf("partial counter treated as complete: %v %v", got, err)
	}
}
