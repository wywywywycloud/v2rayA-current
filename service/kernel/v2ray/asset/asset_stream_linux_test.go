package asset

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStreamAssetReplacesStagingSymlink(t *testing.T) {
	dir := t.TempDir()
	referent, target := filepath.Join(dir, "previous.dat"), filepath.Join(dir, "asset.dat.new")
	if err := os.WriteFile(referent, []byte("previous"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(referent, target); err != nil {
		t.Fatal(err)
	}
	resp := &http.Response{ContentLength: -1, Body: io.NopCloser(strings.NewReader("new"))}
	if err := writeAssetBody(resp, target, 1024); err != nil {
		t.Fatal(err)
	}
	old, err := os.ReadFile(referent)
	if err != nil || string(old) != "previous" {
		t.Fatalf("referent changed: %q, %v", old, err)
	}
	info, err := os.Lstat(target)
	if err != nil || !info.Mode().IsRegular() {
		t.Fatalf("staging link not replaced: %v", err)
	}
}

func TestStreamAssetBoundaryAndRenameFailure(t *testing.T) {
	for _, size := range []int{0, 1024} {
		dir := t.TempDir()
		target := filepath.Join(dir, "asset")
		resp := &http.Response{ContentLength: -1, Body: io.NopCloser(strings.NewReader(strings.Repeat("x", size)))}
		if err := writeAssetBody(resp, target, 1024); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(target)
		if err != nil || info.Size() != int64(size) {
			t.Fatalf("size %d: %v, %v", size, info, err)
		}
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "occupied")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	resp := &http.Response{ContentLength: -1, Body: io.NopCloser(strings.NewReader("data"))}
	if err := writeAssetBody(resp, target, 1024); err == nil {
		t.Fatal("rename over directory unexpectedly succeeded")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 || !entries[0].IsDir() {
		t.Fatalf("failed rename changed destination or leaked temp: %v, %v", entries, err)
	}
}
