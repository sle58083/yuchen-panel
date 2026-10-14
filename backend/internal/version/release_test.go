package version

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVersionFileParsesVersionAndChannel(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	v, channel, ok := ParseReleaseID(string(raw))
	if !ok || v != Version || channel != Channel {
		t.Fatalf("ParseReleaseID=%q,%q,%v", v, channel, ok)
	}
}

func TestReleaseIDPreservesRC(t *testing.T) {
	v, ch, ok := ParseReleaseID("1.0.3-rc.1")
	if !ok || v != "1.0.3-rc.1" || ch != "rc" {
		t.Fatalf("lost RC: %q %q", v, ch)
	}
	v, ch, ok = ParseReleaseID("1.0.2-stable")
	if !ok || v != "1.0.2" || ch != "stable" {
		t.Fatal("legacy channel parsing")
	}
}
