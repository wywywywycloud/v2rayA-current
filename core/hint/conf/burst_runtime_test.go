package conf_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	multi "github.com/v2rayA/v2raya-core/hint/app/observatory/multiobservatory"
	_ "github.com/v2rayA/v2raya-core/hint/conf"
	_ "github.com/xtls/xray-core/app/dispatcher"
	"github.com/xtls/xray-core/app/observatory/burst"
	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	"github.com/xtls/xray-core/common/serial"
	core "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/extension"
	_ "github.com/xtls/xray-core/proxy/freedom"
	_ "github.com/xtls/xray-core/transport/internet/tagged/taggedimpl"
)

func TestBurstRuntimeInventory(t *testing.T) {
	for _, mode := range []string{"loaded", "native-only", "multi-only"} {
		t.Run(mode, func(t *testing.T) {
			var nativeRequests, multiRequests, active atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				active.Add(1)
				defer active.Add(-1)
				if r.URL.Path == "/native" {
					nativeRequests.Add(1)
				} else {
					multiRequests.Add(1)
				}
				<-r.Context().Done()
			}))
			defer server.Close()
			raw := fmt.Sprintf(`{"log":{"loglevel":"none"},"outbounds":[{"protocol":"freedom","tag":"direct"}],"burstObservatory":{"subjectSelector":["direct"],"pingConfig":{"destination":%q,"interval":"1h","timeout":"30s"}}}`, server.URL)
			cfg, err := core.LoadConfig("json", strings.NewReader(raw))
			if err != nil {
				t.Fatal(err)
			}
			apps := cfg.App[:0]
			native, multis := 0, 0
			for _, app := range cfg.App {
				v, err := app.GetInstance()
				if err != nil {
					t.Fatal(err)
				}
				switch v := v.(type) {
				case *burst.Config:
					if mode == "multi-only" {
						continue
					}
					native++
					v.PingConfig.Destination = server.URL + "/native"
					app = serial.ToTypedMessage(v)
				case *multi.Config:
					if mode == "native-only" {
						continue
					}
					multis++
					for _, o := range v.Observers {
						o.ProbeUrl = server.URL + "/multi"
					}
					app = serial.ToTypedMessage(v)
				}
				apps = append(apps, app)
			}
			cfg.App = apps
			instance, err := core.New(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer instance.Close()
			if err = instance.Start(); err != nil {
				t.Fatal(err)
			}
			deadline := time.Now().Add(3 * time.Second)
			for time.Now().Before(deadline) && (nativeRequests.Load() < int32(native) || multiRequests.Load() < int32(multis)) {
				time.Sleep(10 * time.Millisecond)
			}
			stack := make([]byte, 1<<20)
			n := runtime.Stack(stack, true)
			initialChecks := bytes.Count(stack[:n], []byte("burst.(*HealthPing).Check("))
			t.Logf("mode=%s nativeApps=%d multiApps=%d nativeHTTP=%d multiHTTP=%d active=%d initialCheckGoroutines=%d lookup=%T", mode, native, multis, nativeRequests.Load(), multiRequests.Load(), active.Load(), initialChecks, instance.GetFeature(extension.ObservatoryType()))
			if nativeRequests.Load() < int32(native) || multiRequests.Load() < int32(multis) || initialChecks != native+multis {
				t.Errorf("runtime counts do not match configured observers")
			}
			if err := instance.Close(); err != nil {
				t.Fatal(err)
			}
			deadline = time.Now().Add(2 * time.Second)
			for time.Now().Before(deadline) && active.Load() != 0 {
				time.Sleep(10 * time.Millisecond)
			}
			before := nativeRequests.Load() + multiRequests.Load()
			time.Sleep(100 * time.Millisecond)
			if active.Load() != 0 || nativeRequests.Load()+multiRequests.Load() != before {
				t.Error("activity remains after close")
			}
			t.Logf("closed active=%d requests=%d", active.Load(), before)
		})
	}
}
