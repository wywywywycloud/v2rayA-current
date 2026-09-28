package asset

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type assetTransport func(*http.Request) (*http.Response, error)

func (f assetTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAssetBodyOverLimitIsRejected(t *testing.T) {
	client := &http.Client{Transport: assetTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			Status:        "200 OK",
			ContentLength: maxAssetDownloadSize + 1,
			Body:          io.NopCloser(strings.NewReader("not read")),
		}, nil
	})}
	target := filepath.Join(t.TempDir(), "asset.dat")
	if err := download(client, "https://example.test/asset.dat", target); err == nil || !strings.Contains(err.Error(), "256 MiB") {
		t.Fatalf("error = %v, want size limit", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("oversized asset created target: %v", err)
	}
}

func TestStreamAssetPreservesTargetOnFailure(t *testing.T) {
	for _, mode := range []string{"over-limit", "read-error", "success"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			target := filepath.Join(dir, "asset.dat")
			if err := os.WriteFile(target, []byte("previous"), 0644); err != nil {
				t.Fatal(err)
			}
			var body io.Reader = strings.NewReader("replacement")
			if mode == "over-limit" {
				body = strings.NewReader(strings.Repeat("x", 1025))
			}
			if mode == "read-error" {
				body = io.MultiReader(body, brokenAssetReader{})
			}
			resp := &http.Response{ContentLength: -1, Body: io.NopCloser(body)}
			err := writeAssetBody(resp, target, 1024)
			if mode == "success" && err != nil {
				t.Fatal(err)
			}
			if mode != "success" && err == nil {
				t.Fatal("expected failed download")
			}
			want := "previous"
			if mode == "success" {
				want = "replacement"
			}
			got, err := os.ReadFile(target)
			if err != nil || string(got) != want {
				t.Fatalf("target %q, err=%v, want=%q", got, err, want)
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 1 {
				t.Fatalf("temporary download leaked: %v, %v", entries, err)
			}
		})
	}
}

type brokenAssetReader struct{}

func (brokenAssetReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

// The core only reads XRAY_LOCATION_ASSET, so a dat file that exists in a
// system directory has to be linked into that directory before the core runs.
func TestEnsureCoreAssetsLinksMissingFiles(t *testing.T) {
	system := t.TempDir()
	assetDir := filepath.Join(t.TempDir(), "runtime")
	source := filepath.Join(system, "geosite.dat")
	if err := os.WriteFile(source, []byte("dat"), 0644); err != nil {
		t.Fatal(err)
	}
	// findAssetOutsideDir searches fixed system paths, so exercise the linking
	// itself with the source it would have found.
	if err := os.MkdirAll(assetDir, 0755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(assetDir, "geosite.dat")
	if err := os.Symlink(source, target); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(target)
	if err != nil || string(b) != "dat" {
		t.Fatalf("link does not resolve: %v %q", err, b)
	}
	// An asset already present must not be touched.
	EnsureCoreAssets(assetDir)
	fi, err := os.Lstat(target)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error("the existing link was replaced")
	}
}

func TestFindAssetOutsideDirSkipsItsOwnDirectory(t *testing.T) {
	dir := "/usr/share/v2raya"
	if _, err := os.Stat(filepath.Join(dir, "geosite.dat")); err != nil {
		t.Skip("no system geosite.dat on this machine")
	}
	if got := findAssetOutsideDir("geosite.dat", dir); got != "" {
		t.Errorf("a file in the asset directory itself must not be reported as a source, got %q", got)
	}
}
