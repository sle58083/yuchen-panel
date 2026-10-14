package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ClientTraffic struct {
	Uplink   int64 `json:"uplink"`
	Downlink int64 `json:"downlink"`
}

func statsAddress(configPath string) (string, error) {
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	var cfg struct {
		Inbounds []struct {
			Tag, Listen string
			Port        int
		}
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return "", err
	}
	for _, in := range cfg.Inbounds {
		if in.Tag == "api" && in.Listen == "127.0.0.1" && in.Port > 0 && in.Port <= 65535 {
			return fmt.Sprintf("127.0.0.1:%d", in.Port), nil
		}
	}
	return "", fmt.Errorf("loopback StatsService not configured")
}

func parseTraffic(raw []byte) (map[string]ClientTraffic, error) {
	var response struct {
		Stat []struct {
			Name  string
			Value json.RawMessage
		}
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, err
	}
	out := map[string]ClientTraffic{}
	seen := map[string]bool{}
	directions := map[string]int{}
	for _, stat := range response.Stat {
		parts := strings.Split(stat.Name, ">>>")
		if len(parts) != 4 || parts[0] != "user" || !strings.HasPrefix(parts[1], "client_") || parts[2] != "traffic" {
			continue
		}
		id := strings.TrimPrefix(parts[1], "client_")
		if id == "" {
			continue
		}
		if parts[3] != "uplink" && parts[3] != "downlink" {
			continue
		}
		if seen[stat.Name] {
			return nil, fmt.Errorf("duplicate stats counter")
		}
		seen[stat.Name] = true
		var value int64
		var err error
		// Protobuf JSON may omit an int64 field when it has the default 0.
		if len(stat.Value) > 0 {
			value, err = strconv.ParseInt(strings.Trim(string(stat.Value), "\""), 10, 64)
		}
		if err != nil || value < 0 {
			return nil, fmt.Errorf("invalid stats counter")
		}
		c := out[id]
		if parts[3] == "uplink" {
			c.Uplink = value
			directions[id] |= 1
		} else {
			c.Downlink = value
			directions[id] |= 2
		}
		if c.Uplink > math.MaxInt64-c.Downlink {
			return nil, fmt.Errorf("stats total overflow")
		}
		out[id] = c
	}
	// Do not convert an omitted direction into a fake reset of that counter.
	for id := range out {
		if directions[id] != 3 {
			delete(out, id)
		}
	}
	return out, nil
}

// /proc identity includes boot ID, PID and process start ticks; Agent restarts
// keep this epoch stable while Xray restarts always change it.
func xrayEpoch(configPath string) (string, error) {
	return processEpoch("/proc", configPath)
}

func processEpoch(procRoot, configPath string) (string, error) {
	entries, err := os.ReadDir(procRoot)
	if err != nil {
		return "", err
	}
	boot, err := os.ReadFile(filepath.Join(procRoot, "sys/kernel/random/boot_id"))
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if _, err := strconv.Atoi(entry.Name()); err != nil {
			continue
		}
		dir := filepath.Join(procRoot, entry.Name())
		cmd, _ := os.ReadFile(filepath.Join(dir, "cmdline"))
		args := strings.Split(strings.TrimRight(string(cmd), "\x00"), "\x00")
		if len(args) < 2 || filepath.Base(args[0]) != "xray" {
			continue
		}
		found := false
		for _, arg := range args[1:] {
			if arg == configPath || arg == "-config="+configPath {
				found = true
			}
		}
		if !found {
			continue
		}
		stat, err := os.ReadFile(filepath.Join(dir, "stat"))
		if err != nil {
			continue
		}
		end := strings.LastIndex(string(stat), ")")
		if end < 0 {
			continue
		}
		fields := strings.Fields(string(stat)[end+1:])
		if len(fields) < 20 {
			continue
		}
		return strings.TrimSpace(string(boot)) + ":" + entry.Name() + ":" + fields[19], nil
	}
	return "", fmt.Errorf("Xray process identity unavailable")
}

func collectXrayTraffic(cfg Config) (map[string]ClientTraffic, string, error) {
	address, err := statsAddress(cfg.ConfigPath)
	if err != nil {
		return nil, "", err
	}
	epoch, err := xrayEpoch(cfg.ConfigPath)
	if err != nil {
		return nil, "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Xray's bundled CLI is the gRPC client: no grpcurl or Go modules needed.
	raw, err := exec.CommandContext(ctx, "xray", "api", "statsquery", "--server="+address, "-pattern", "user>>>client_", "-reset=false").Output()
	if err != nil {
		return nil, "", fmt.Errorf("StatsService query failed: %w", err)
	}
	traffic, err := parseTraffic(raw)
	if err != nil {
		return nil, "", err
	}
	after, err := xrayEpoch(cfg.ConfigPath)
	if err != nil || after != epoch {
		return nil, "", fmt.Errorf("Xray restarted during stats query")
	}
	return traffic, epoch, nil
}
