package main

import (
	"encoding/json"
	"runtime"
	"testing"
)

func TestMemoryCapacityReusesUsageRead(t *testing.T) {
	usage, total := parseMemoryMetrics("MemTotal: 1024000 kB\nMemAvailable: 512000 kB\n")
	if usage != 50 || total != 1048576000 {
		t.Fatalf("usage=%v bytes=%d", usage, total)
	}
	for _, raw := range []string{"", "MemTotal: -1 kB", "MemTotal: broken kB"} {
		usage, total := parseMemoryMetrics(raw)
		if total != 0 || usage != 0 {
			t.Fatal("invalid memory metrics reported")
		}
	}
}
func TestHeartbeatReportsSpecs(t *testing.T) {
	hb := collectHeartbeat(Config{}, "fixture-host", "", "fixture")
	if hb.CPUCores != runtime.NumCPU() {
		t.Fatal("CPU count missing")
	}
	raw, err := json.Marshal(hb)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err = json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if _, ok := fields["cpu_cores"]; !ok {
		t.Fatal("capacity not serialized")
	}
}
