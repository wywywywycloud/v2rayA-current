package asset

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Compare the r20 read-then-write algorithm with bounded streaming. The
// immutable source is allocated before timing, identically for both variants.
func BenchmarkAssetBodyStorage(b *testing.B) {
	const size = 8 << 20
	body := strings.Repeat("x", size)
	for _, stream := range []bool{false, true} {
		name := "r20-readall"
		if stream {
			name = "stream"
		}
		b.Run(name, func(b *testing.B) {
			target := filepath.Join(b.TempDir(), "asset.dat.new")
			b.ReportAllocs()
			b.SetBytes(size)
			b.ResetTimer()
			for range b.N {
				resp := &http.Response{ContentLength: size, Body: io.NopCloser(strings.NewReader(body))}
				if stream {
					if err := writeAssetBody(resp, target, maxAssetDownloadSize); err != nil {
						b.Fatal(err)
					}
				} else {
					data, err := io.ReadAll(io.LimitReader(resp.Body, maxAssetDownloadSize+1))
					if err != nil {
						b.Fatal(err)
					}
					if err = os.WriteFile(target, data, 0644); err != nil {
						b.Fatal(err)
					}
				}
				resp.Body.Close()
			}
		})
	}
}
