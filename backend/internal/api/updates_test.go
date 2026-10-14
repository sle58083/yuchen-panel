// SPDX-License-Identifier: AGPL-3.0-only
package api

import "testing"

func TestComparePanelVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"0.7.6.0-base-stable-agent-xray", "0.7.5.9.1-qr-flow-compatibility-fix-agent-xray", 1},
		{"1.0.2-clash-import-polish-agent-xray", "0.7.6.0-base-stable-agent-xray", 1},
		{"v1.0.2", "1.0.2-clash-import-polish-agent-xray", 1},
		{"0.7.5.9.1", "0.7.6.0", -1},
		{"0.7.6", "0.7.6.0", 0},
	}
	for _, tc := range cases {
		got := comparePanelVersions(tc.a, tc.b)
		if got < 0 {
			got = -1
		} else if got > 0 {
			got = 1
		}
		if got != tc.want {
			t.Fatalf("comparePanelVersions(%q, %q)=%d want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestUpdateAvailableDoesNotAllowDowngrade(t *testing.T) {
	if isRemoteVersionNewer("0.7.5.9.1-qr-flow-compatibility-fix-agent-xray", "0.7.6.0-base-stable-agent-xray") {
		t.Fatal("older remote version must not be treated as update")
	}
	if !isRemoteVersionNewer("1.0.2-clash-import-polish-agent-xray", "0.7.6.0-base-stable-agent-xray") {
		t.Fatal("newer remote version should be treated as update")
	}
}

func TestVersionMatchesTargetSemanticBoundaries(t *testing.T) {
	cases := []struct {
		current, target string
		want            bool
	}{
		{"1.0.2", "1.0.20", false},
		{"1.0.2", "1.0.2-stable", true},
		{"1x0y2", "1.0.2", false},
		{"1.10.0", "1.9.0", false},
		{"1.0", "1.0.0", true},
		{"v1.0.2", "1.0.2", true},
		{"", "1.0.2", false},
		{"1.0.2", "", false},
		{"1.0.2-rc.1", "1.0.2", false},
		{"1.0.2-rc.1", "v1.0.2-rc.1", true},
	}
	for _, tc := range cases {
		if got := versionMatchesTarget(tc.current, tc.target); got != tc.want {
			t.Errorf("versionMatchesTarget(%q, %q)=%v want %v", tc.current, tc.target, got, tc.want)
		}
	}
}

func TestCompareSemanticVersionNumericOrdering(t *testing.T) {
	a, _ := parseSemanticVersion("1.10.0")
	b, _ := parseSemanticVersion("1.9.0")
	if got := compareSemanticVersion(a, b); got <= 0 {
		t.Fatalf("1.10.0 must be greater than 1.9.0: %d", got)
	}
}

func TestCompareSemanticVersionPrereleaseNumbers(t *testing.T) {
	a, _ := parseSemanticVersion("1.0.2-rc.10")
	b, _ := parseSemanticVersion("1.0.2-rc.9")
	if compareSemanticVersion(a, b) <= 0 {
		t.Fatal("rc.10 must follow rc.9")
	}
}
