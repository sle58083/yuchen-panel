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
	if !ok || v != Version || channel != "stable" {
		t.Fatalf("ParseReleaseID=%q,%q,%v", v, channel, ok)
	}
}
