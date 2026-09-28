package conf_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	compat "github.com/v2rayA/v2raya-core/hint/app/observatory/command"
	multi "github.com/v2rayA/v2raya-core/hint/app/observatory/multiobservatory"
	obs "github.com/xtls/xray-core/app/observatory"
	"github.com/xtls/xray-core/app/observatory/burst"
	obscmd "github.com/xtls/xray-core/app/observatory/command"
	"github.com/xtls/xray-core/common/cmdarg"
	core "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/extension"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestBurstSingleConfig(t *testing.T) {
	raw := `{"burstObservatory":{"subjectSelector":["direct"],"pingConfig":{"destination":"http://localhost/probe","connectivity":"http://localhost/connect","interval":"17s","timeout":"3s","sampling":4,"httpMethod":"GET"}}}`
	path := filepath.Join(t.TempDir(), "burst.json")
	if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"reader", "files", "args"} {
		t.Run(mode, func(t *testing.T) {
			var cfg *core.Config
			var err error
			switch mode {
			case "reader":
				cfg, err = core.LoadConfig("json", strings.NewReader(raw))
			case "files":
				cfg, err = core.ConfigBuilderForFiles([]*core.ConfigSource{{Name: path, Format: "json"}})
			case "args":
				cfg, err = core.LoadConfig("json", cmdarg.Arg{path})
			}
			if err != nil {
				t.Fatal(err)
			}
			native, multis := 0, 0
			for _, app := range cfg.App {
				v, err := app.GetInstance()
				if err != nil {
					t.Fatal(err)
				}
				switch v := v.(type) {
				case *burst.Config:
					native++
					p := v.PingConfig
					if p.Destination != "http://localhost/probe" || p.Connectivity != "http://localhost/connect" || p.Interval != int64(17*time.Second) || p.Timeout != int64(3*time.Second) || p.SamplingCount != 4 || p.HttpMethod != "GET" {
						t.Errorf("native settings changed: %v", p)
					}
				case *multi.Config:
					multis++
				}
			}
			t.Logf("native=%d multi=%d", native, multis)
			if native != 1 || multis != 0 {
				t.Errorf("want exactly one native burst")
			}
		})
	}
}

func TestBurstCompatibleAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(5 * time.Millisecond); w.WriteHeader(204) }))
	defer server.Close()
	cases := []struct {
		name, extension string
		tags            map[string]int
	}{
		{"burst", fmt.Sprintf(`"burstObservatory":{"subjectSelector":["direct-a"],"pingConfig":{"destination":%q,"interval":"1h","sampling":3,"timeout":"2s","httpMethod":"GET"}}`, server.URL), map[string]int{"": 1, "_burst_global": 1, "unknown": 1}},
		{"ordinary", fmt.Sprintf(`"observatory":{"subjectSelector":["direct-a"],"probeURL":%q,"probeInterval":"1h"}`, server.URL), map[string]int{"": 1, "unknown": 1}},
		{"nested", fmt.Sprintf(`"multiObservatory":{"observers":[{"tag":"a","burstObservatory":{"subjectSelector":["direct-a"],"pingConfig":{"destination":%q,"interval":"1h"}}},{"tag":"b","burstObservatory":{"subjectSelector":["direct-b"],"pingConfig":{"destination":%q,"interval":"1h"}}}]}`, server.URL, server.URL), map[string]int{"": 2, "a": 1, "b": 1, "unknown": 2}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			addr := listener.Addr().String()
			listener.Close()
			raw := fmt.Sprintf(`{"log":{"loglevel":"none"},"api":{"tag":"api","listen":%q,"services":[]},"outbounds":[{"protocol":"freedom","tag":"direct-a"},{"protocol":"freedom","tag":"direct-b"}],%s}`, addr, tc.extension)
			cfg, err := core.LoadConfig("json", strings.NewReader(raw))
			if err != nil {
				t.Fatal(err)
			}
			instance, err := core.New(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer instance.Close()
			if err = instance.Start(); err != nil {
				t.Fatal(err)
			}
			client, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			for tag, want := range tc.tags {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				response := new(obscmd.GetOutboundStatusResponse)
				for {
					err = client.Invoke(ctx, "/v2ray.core.app.observatory.command.ObservatoryService/GetOutboundStatus", &compat.GetOutboundStatusRequest{Tag: tag}, response)
					if err == nil && len(response.GetStatus().GetStatus()) == want {
						break
					}
					if ctx.Err() != nil {
						t.Fatalf("tag=%q err=%v response=%v", tag, err, response)
					}
					time.Sleep(10 * time.Millisecond)
				}
				for _, status := range response.Status.Status {
					if !status.Alive {
						t.Errorf("not alive: %v", status)
					}
					if (tag == "a" || tag == "b") && status.OutboundTag != "direct-"+tag {
						t.Errorf("wrong group: %v", status)
					}
				}
				t.Logf("tag=%q statuses=%d lookup=%T", tag, len(response.Status.Status), instance.GetFeature(extension.ObservatoryType()))
			}
		})
	}
}

func TestBurstLifecycleCycles(t *testing.T) {
	before := runtime.NumGoroutine()
	for i := 0; i < 10; i++ {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(5 * time.Millisecond); w.WriteHeader(204) }))
		raw := fmt.Sprintf(`{"log":{"loglevel":"none"},"outbounds":[{"protocol":"freedom","tag":"direct"}],"burstObservatory":{"subjectSelector":["direct"],"pingConfig":{"destination":%q,"interval":"1h"}}}`, server.URL)
		cfg, err := core.LoadConfig("json", strings.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		instance, err := core.New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { instance.Close(); server.Close() })
		if err = instance.Start(); err != nil {
			t.Fatal(err)
		}
		deadline := time.Now().Add(3 * time.Second)
		for {
			result, err := instance.GetFeature(extension.ObservatoryType()).(extension.Observatory).GetObservation(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if len(result.(*obs.ObservationResult).Status) > 0 {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("no initial result")
			}
			time.Sleep(10 * time.Millisecond)
		}
		if err = instance.Close(); err != nil {
			t.Fatal(err)
		}
		server.Close()
	}
	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		runtime.GC()
		time.Sleep(20 * time.Millisecond)
	}
	after := runtime.NumGoroutine()
	t.Logf("10 fresh-instance restart cycles goroutines=%d->%d", before, after)
	if after > before {
		t.Errorf("goroutines remain after close")
	}
}
