package conf

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	multiobs "github.com/v2rayA/v2raya-core/hint/app/observatory/multiobservatory"
	"github.com/xtls/xray-core/app/observatory"
	"github.com/xtls/xray-core/app/observatory/burst"
	"github.com/xtls/xray-core/common/cmdarg"
	xcore "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/extension"
)

func TestNativeBurstAliasRuntime(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method=%s want GET", r.Method)
		}
		requests.Add(1)
		time.Sleep(2 * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	// No selectors: invoke this sole production observer synchronously, without scheduler noise.
	raw := fmt.Sprintf(`{"outbounds":[{"tag":"direct","protocol":"freedom"}],"burstObservatory":{"pingConfig":{"destination":%q,"interval":"10s","samplingCount":2,"httpMethod":"GET"}}}`, server.URL)
	cfg, err := xcore.LoadConfig("json", strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if ping := nativeForwardedPing(t, cfg); ping.SamplingCount != 2 {
		t.Fatalf("sampling=%d", ping.SamplingCount)
	}
	instance, err := xcore.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer instance.Close()
	if err := instance.Start(); err != nil {
		t.Fatal(err)
	}
	observer := instance.GetFeature(extension.ObservatoryType()).(*burst.Observer)
	for i := 0; i < 4; i++ {
		observer.Check([]string{"direct"})
	}
	result, err := observer.GetObservation(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	status := result.(*observatory.ObservationResult).Status
	if len(status) != 1 || status[0].HealthPing.All != 2 || !status[0].Alive {
		t.Fatalf("status=%v", status)
	}
	if requests.Load() != 4 {
		t.Fatalf("requests=%d want 4", requests.Load())
	}
	t.Logf("native owners=1 multi owners=0 endpoint requests=%d sampling window=%d", requests.Load(), status[0].HealthPing.All)
}

func nativeForwardedPing(t *testing.T, cfg *xcore.Config) *burst.HealthPingConfig {
	t.Helper()
	var ping *burst.HealthPingConfig
	for _, app := range cfg.App {
		value, err := app.GetInstance()
		if err != nil {
			t.Fatal(err)
		}
		switch value := value.(type) {
		case *burst.Config:
			if ping != nil {
				t.Fatal("duplicate native Burst")
			}
			ping = value.PingConfig
		case *multiobs.Config:
			t.Fatal("top-level Burst unexpectedly adapted to Multi")
		}
	}
	if ping == nil {
		t.Fatal("missing native Burst")
	}
	return ping
}

func TestNativeBurstForwardedAliases(t *testing.T) {
	for _, tc := range []struct {
		name, sampling string
		want           int32
	}{
		{"native", `,"sampling":3`, 3},
		{"v5", `,"samplingCount":3`, 3},
		{"equal", `,"sampling":3,"samplingCount":3`, 3},
		{"absent", "", 0},
		{"zero", `,"samplingCount":0`, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := fmt.Sprintf(`{"burstObservatory":{"subjectSelector":["direct"],"pingConfig":{"destination":"http://localhost/probe","connectivity":"http://localhost/connect","interval":"17s","timeout":"1750ms","httpMethod":"GET"%s}}}`, tc.sampling)
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
				t.Fatal(err)
			}
			for source, load := range map[string]func() (*xcore.Config, error){
				"reader": func() (*xcore.Config, error) { return xcore.LoadConfig("json", strings.NewReader(raw)) },
				"files": func() (*xcore.Config, error) {
					return buildConfigFromFiles([]*xcore.ConfigSource{{Name: path, Format: "json"}})
				},
				"args": func() (*xcore.Config, error) { return xcore.LoadConfig("json", cmdarg.Arg{path}) },
			} {
				t.Run(source, func(t *testing.T) {
					cfg, err := load()
					if err != nil {
						t.Fatal(err)
					}
					p := nativeForwardedPing(t, cfg)
					if p.SamplingCount != tc.want || p.Destination != "http://localhost/probe" || p.Connectivity != "http://localhost/connect" || p.Interval != int64(17*time.Second) || p.Timeout != int64(1750*time.Millisecond) || p.HttpMethod != "GET" {
						t.Fatalf("native ping=%v", p)
					}
				})
			}
		})
	}
}

func TestNativeBurstAliasFileOverride(t *testing.T) {
	alias := `{"burstObservatory":{"subjectSelector":["direct"],"pingConfig":{"samplingCount":2}}}`
	native := `{"burstObservatory":{"subjectSelector":["direct"],"pingConfig":{"sampling":4}}}`
	for _, tc := range []struct {
		name string
		docs []string
		want int32
	}{
		{"alias_then_native", []string{alias, native}, 4},
		{"native_then_alias", []string{native, alias}, 2},
		{"unrelated_later_file", []string{alias, `{}`}, 2},
		{"empty_multi_later_file", []string{alias, `{"multiObservatory":{}}`}, 2},
		{"null_burst_later_file", []string{alias, `{"burstObservatory":null}`}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var files []*xcore.ConfigSource
			var args cmdarg.Arg
			for i, doc := range tc.docs {
				path := filepath.Join(t.TempDir(), fmt.Sprintf("%d.json", i))
				if err := os.WriteFile(path, []byte(doc), 0600); err != nil {
					t.Fatal(err)
				}
				files = append(files, &xcore.ConfigSource{Name: path, Format: "json"})
				args = append(args, path)
			}
			for mode, load := range map[string]func() (*xcore.Config, error){
				"files": func() (*xcore.Config, error) { return buildConfigFromFiles(files) },
				"args":  func() (*xcore.Config, error) { return xcore.LoadConfig("json", args) },
			} {
				t.Run(mode, func(t *testing.T) {
					cfg, err := load()
					if err != nil {
						t.Fatal(err)
					}
					if p := nativeForwardedPing(t, cfg); p.SamplingCount != tc.want {
						t.Fatalf("sampling=%d want %d", p.SamplingCount, tc.want)
					}
				})
			}
		})
	}
}

func TestNativeBurstAliasConflict(t *testing.T) {
	raw := `{"burstObservatory":{"subjectSelector":["direct"],"pingConfig":{"sampling":2,"samplingCount":3}}}`
	if _, err := xcore.LoadConfig("json", strings.NewReader(raw)); err == nil || !strings.Contains(err.Error(), "sampling and samplingCount disagree") {
		t.Fatalf("conflict error=%v", err)
	}
}
