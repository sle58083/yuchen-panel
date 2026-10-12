package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteUpgradeRunnerEscapesShellMetachars(t *testing.T) {
	for _, value := range []string{"1.0.3-$(id)", "`id`", "single'quote", "double\"quote"} {
		t.Run(value, func(t *testing.T) {
			root := t.TempDir()
			marker := filepath.Join(root, "must-not-exist")
			literal := value + "$(touch " + marker + ")`touch " + marker + "`"
			task := upgradeTaskState{ID: "test", DownloadURL: literal, TargetVersion: literal, Package: "test.zip", SHA256: strings.Repeat("0", 64), LogFile: filepath.Join(root, "log"), WorkDir: filepath.Join(root, "work")}
			path, err := writeUpgradeRunnerAt(task, "https://example.com/manifest.json", root)
			if err != nil {
				t.Fatal(err)
			}
			if out, err := exec.Command("bash", "-n", path).CombinedOutput(); err != nil {
				t.Fatalf("bash -n: %v: %s", err, out)
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			prefix := strings.Split(string(raw), "\nmkdir -p")[0]
			cmd := exec.Command("bash", "-c", prefix+"\nprintf '%s' \"$DOWNLOAD_URL\"")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("literal round-trip: %v: %s", err, out)
			}
			if string(out) != literal {
				t.Fatalf("literal changed: %q", out)
			}
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatal("shell input executed")
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != 0700 {
				t.Fatal("runner permissions are not private")
			}
		})
	}
	for _, value := range []string{"line\nbreak", "line\rbreak", "nul\x00byte"} {
		task := upgradeTaskState{ID: "test", DownloadURL: value}
		if _, err := writeUpgradeRunnerAt(task, "https://example.com/m", t.TempDir()); err == nil {
			t.Fatal("control character accepted")
		}
	}
}

func TestFetchManifestRequiresHTTPSAndSameOriginRedirects(t *testing.T) {
	for _, url := range []string{"http://example.com/m", "https://user:password@example.com/m", "file:///tmp/m"} {
		if _, err := fetchUpdateManifest(context.Background(), url); err == nil {
			t.Fatal("unsafe manifest URL accepted")
		}
	}
	var origin string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/valid":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"version":"1.0.3"}`))
		case "/same":
			http.Redirect(w, r, origin+"/valid", 302)
		case "/other":
			http.Redirect(w, r, "https://example.com/m", 302)
		case "/downgrade":
			http.Redirect(w, r, "http://example.com/m", 302)
		default:
			http.Redirect(w, r, origin+"/loop", 302)
		}
	}))
	defer server.Close()
	origin = server.URL
	old := http.DefaultTransport
	http.DefaultTransport = server.Client().Transport
	t.Cleanup(func() { http.DefaultTransport = old })
	for _, path := range []string{"/valid", "/same"} {
		if _, err := fetchUpdateManifest(context.Background(), origin+path); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{"/other", "/downgrade", "/loop"} {
		if _, err := fetchUpdateManifest(context.Background(), origin+path); err == nil {
			t.Fatalf("redirect %s accepted", path)
		}
	}
}
