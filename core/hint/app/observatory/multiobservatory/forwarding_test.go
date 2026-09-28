package multiobservatory

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/xtls/xray-core/app/dispatcher"
	"github.com/xtls/xray-core/app/observatory"
	"github.com/xtls/xray-core/app/observatory/burst"
	"github.com/xtls/xray-core/app/proxyman"
	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	"github.com/xtls/xray-core/common/serial"
	xcore "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/extension"
	"github.com/xtls/xray-core/proxy/freedom"
	_ "github.com/xtls/xray-core/transport/internet/tagged/taggedimpl"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Reflection keeps this same regression source compilable against the old schema.
func setPingField(t *testing.T, c *ObserverConfig, name string, value any) {
	t.Helper()
	field := c.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name(name))
	if field == nil {
		t.Logf("baseline lacks %s", name)
		return
	}
	c.ProtoReflect().Set(field, protoreflect.ValueOf(value))
}

func TestObservatoryRuntimeHTTP(t *testing.T) {
	for _, method := range []string{"GET", "HEAD"} {
		t.Run(method, func(t *testing.T) {
			var mu sync.Mutex
			methods := []string{}
			paths := map[string]int{}
			connectivity := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				methods = append(methods, r.Method)
				paths[r.URL.Path]++
				if r.URL.Path == "/connectivity" {
					connectivity++
				}
				mu.Unlock()
				switch r.URL.Path {
				case "/ok":
					// Avoid zero RTT on hosts with coarse timer resolution.
					time.Sleep(2 * time.Millisecond)
				case "/slow":
					<-r.Context().Done()
					return
				case "/broken":
					conn, _, err := w.(http.Hijacker).Hijack()
					if err == nil {
						conn.Close()
					}
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()
			configs := []*ObserverConfig{}
			for _, path := range []string{"/ok", "/slow", "/broken"} {
				c := &ObserverConfig{Tag: path, ProbeUrl: server.URL + path, ProbeInterval: int64(13 * time.Second)}
				setPingField(t, c, "http_method", method)
				setPingField(t, c, "sampling_count", int32(2))
				timeout := 2 * time.Second
				if path == "/slow" {
					timeout = 60 * time.Millisecond
				}
				setPingField(t, c, "timeout", int64(timeout))
				setPingField(t, c, "connectivity", server.URL+"/connectivity")
				configs = append(configs, c)
			}
			instance, err := xcore.New(&xcore.Config{
				App:      []*serial.TypedMessage{serial.ToTypedMessage(&dispatcher.Config{}), serial.ToTypedMessage(&proxyman.InboundConfig{}), serial.ToTypedMessage(&proxyman.OutboundConfig{}), serial.ToTypedMessage(&Config{Observers: configs})},
				Outbound: []*xcore.OutboundHandlerConfig{{Tag: "direct", ProxySettings: serial.ToTypedMessage(&freedom.Config{})}},
			})
			if err != nil {
				t.Fatal(err)
			}
			defer instance.Close()
			// Empty selectors prevent background probes; exercise the production child directly.
			if err := instance.Start(); err != nil {
				t.Fatal(err)
			}
			mo := instance.GetFeature(extension.ObservatoryType()).(*MultiObservatory)
			for i := 0; i < 4; i++ {
				mo.children["/ok"].(*burst.Observer).Check([]string{"direct"})
			}
			result, err := mo.GetObservationByTag("/ok", context.Background())
			if err != nil {
				t.Fatal(err)
			}
			status := result.(*observatory.ObservationResult).Status
			if len(status) != 1 {
				t.Fatalf("statuses=%v", status)
			}
			t.Logf("method=%s sample-window=%d", method, status[0].HealthPing.All)
			if status[0].HealthPing.All != 2 {
				t.Errorf("sample window=%d want 2", status[0].HealthPing.All)
			}
			if !status[0].Alive || status[0].HealthPing.Fail != 0 {
				t.Fatalf("successful probes failed before timeout/EOF controls: %v", status)
			}
			started := time.Now()
			mo.children["/slow"].(*burst.Observer).Check([]string{"direct"})
			elapsed := time.Since(started)
			result, _ = mo.GetObservationByTag("/slow", context.Background())
			status = result.(*observatory.ObservationResult).Status
			t.Logf("slow elapsed=%s status=%v", elapsed, status)
			if len(status) != 1 || status[0].Alive {
				t.Errorf("timeout not applied: %v", status)
			}
			if elapsed > 250*time.Millisecond {
				t.Errorf("timeout elapsed=%s expected under 250ms", elapsed)
			}
			mo.children["/broken"].(*burst.Observer).Check([]string{"direct"})
			mu.Lock()
			defer mu.Unlock()
			t.Logf("endpoint methods=%v paths=%v connectivity requests=%d", methods, paths, connectivity)
			if connectivity != 2 {
				t.Errorf("connectivity calls=%d want 2", connectivity)
			}
			for _, got := range methods {
				if got != method {
					t.Error(fmt.Sprintf("endpoint method=%s want %s", got, method))
				}
			}
		})
	}
}
