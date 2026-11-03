package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"v1.2.0", "v1.1.9", 1},
		{"1.0.0", "v1.0.0", 0},
		{"v1.0.0", "v1.0.1", -1},
		{"v2.0", "v1.9.9", 1},
		{"v1.0.0-rc1", "v1.0.0", 0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_vs_%s", tt.a, tt.b), func(t *testing.T) {
			got := compareVersions(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestIsReleaseVersion(t *testing.T) {
	tests := []struct {
		v    string
		want bool
	}{
		{"v1.2.3", true},
		{"1.2.3", true},
		{"dev", false},
		{"f70b7f4", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.v, func(t *testing.T) {
			got := isReleaseVersion(tt.v)
			if got != tt.want {
				t.Errorf("isReleaseVersion(%q) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestAssetName(t *testing.T) {
	tests := []struct {
		goos, goarch string
		want         []string
	}{
		{"windows", "amd64", []string{"acronomicon-windows-amd64.exe"}},
		{"darwin", "arm64", []string{"acronomicon-darwin-arm64", "acronomicon-macos-universal"}},
		{"linux", "amd64", []string{"acronomicon-linux-amd64"}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.goos, tt.goarch), func(t *testing.T) {
			got := assetName(tt.goos, tt.goarch)
			if len(got) != len(tt.want) {
				t.Errorf("assetName(%q, %q) returned %d items, want %d", tt.goos, tt.goarch, len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("assetName(%q, %q)[%d] = %q, want %q", tt.goos, tt.goarch, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSwapBinaryUnix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping swap test in short mode")
	}

	dir := t.TempDir()
	dst := filepath.Join(dir, "binary")
	src := filepath.Join(dir, "binary.new")

	if err := os.WriteFile(dst, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := swapBinaryFor("linux", src, dst); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Errorf("dst content = %q, want %q", string(content), "new")
	}

	if _, err := os.Stat(src); err == nil {
		t.Errorf("src file should be removed")
	}
}

func TestSwapBinaryWindows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping swap test in short mode")
	}

	dir := t.TempDir()
	dst := filepath.Join(dir, "binary.exe")
	src := filepath.Join(dir, "binary.exe.new")

	if err := os.WriteFile(dst, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := swapBinaryFor("windows", src, dst); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Errorf("dst content = %q, want %q", string(content), "new")
	}

	oldContent, err := os.ReadFile(dst + ".old")
	if err != nil {
		t.Fatal(err)
	}
	if string(oldContent) != "old" {
		t.Errorf("dst.old content = %q, want %q", string(oldContent), "old")
	}

	cleanupOld(dst)
	if _, err := os.Stat(dst + ".old"); err == nil {
		t.Errorf("dst.old should be removed by cleanupOld")
	}
}

func TestCheckRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/espresso20/acronomicon/releases/latest" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"tag_name":"v1.3.0","assets":[]}`)
			return
		}
		w.WriteHeader(404)
	}))
	defer server.Close()

	u := &Updater{
		APIBase: server.URL,
		Client:  server.Client(),
		Version: "v1.2.0",
		RepoDir: "",
		Out:     io.Discard,
	}

	st, err := u.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if !st.Available {
		t.Errorf("Available = false, want true")
	}
	if st.Latest != "v1.3.0" {
		t.Errorf("Latest = %q, want %q", st.Latest, "v1.3.0")
	}
	if st.Method != MethodRelease {
		t.Errorf("Method = %v, want %v", st.Method, MethodRelease)
	}

	u.Version = "v1.3.0"
	st, err = u.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.Available {
		t.Errorf("Available = true, want false")
	}
}

func TestCheckCompare(t *testing.T) {
	rev := "aaaaaaa111111111111111111111111111111111"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/espresso20/acronomicon/releases/latest" {
			w.WriteHeader(404)
			return
		}
		if r.URL.Path == fmt.Sprintf("/repos/espresso20/acronomicon/compare/%s...main", rev) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":"ahead","ahead_by":2,"commits":[{"sha":"aaaaaaa111"},{"sha":"bbbbbbb222"}]}`)
			return
		}
		if r.URL.Path == "/repos/espresso20/acronomicon/compare/0000000...main" {
			w.WriteHeader(404)
			return
		}
		w.WriteHeader(404)
	}))
	defer server.Close()

	u := &Updater{
		APIBase:  server.URL,
		Client:   server.Client(),
		Version:  "dev",
		RepoDir:  "",
		Out:      io.Discard,
		revision: rev,
	}

	st, err := u.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if !st.Available {
		t.Errorf("Available = false, want true")
	}
	if st.Latest != "bbbbbbb" {
		t.Errorf("Latest = %q, want %q", st.Latest, "bbbbbbb")
	}

	// Test 404 case
	u.revision = "0000000000000000000000000000000000000000"
	st, err = u.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.Available {
		t.Errorf("Available = true, want false (404 case)")
	}
}

func TestCheckCached(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/espresso20/acronomicon/releases/latest" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"tag_name":"v1.3.0","assets":[]}`)
			return
		}
		w.WriteHeader(404)
	}))
	defer server.Close()

	dir := t.TempDir()
	cacheFile := filepath.Join(dir, "update-check.json")

	u := &Updater{
		APIBase:   server.URL,
		Client:    server.Client(),
		Version:   "v1.2.0",
		RepoDir:   "",
		Out:       io.Discard,
		CacheFile: cacheFile,
	}

	st1, err := u.CheckCached(context.Background(), 1*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	st2, err := u.CheckCached(context.Background(), 1*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	if st1.Latest != st2.Latest {
		t.Errorf("cache mismatch: %q vs %q", st1.Latest, st2.Latest)
	}

	// Check with ttl = 0 should hit server again
	st3, err := u.CheckCached(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}

	if st3.Latest != "v1.3.0" {
		t.Errorf("Latest = %q, want %q", st3.Latest, "v1.3.0")
	}
}

func BenchmarkCompareVersions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compareVersions("v1.2.3", "v1.2.4")
	}
}
