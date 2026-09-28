package conf

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	multiobs "github.com/v2rayA/v2raya-core/hint/app/observatory/multiobservatory"
	"github.com/xtls/xray-core/app/observatory"
	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	xcore "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/extension"
	_ "github.com/xtls/xray-core/transport/internet/tagged/taggedimpl"
)

func TestObservatoryJSONRuntime(t *testing.T) {
	for _, tc := range []struct {
		name, method string
		delay        time.Duration
		alive        bool
	}{
		{"get", "GET", 2 * time.Millisecond, true},
		{"head", "HEAD", 2 * time.Millisecond, true},
		{"timeout", "GET", 300 * time.Millisecond, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var mu sync.Mutex
			methods := []string{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				methods = append(methods, r.Method)
				mu.Unlock()
				time.Sleep(tc.delay)
				w.WriteHeader(204)
			}))
			defer server.Close()
			raw := fmt.Sprintf(`{"outbounds":[{"tag":"direct","protocol":"freedom"}],"multiObservatory":{"observers":[{"tag":"group","settings":{"subjectSelector":["direct"],"pingConfig":{"destination":%q,"interval":"10s","timeout":"60ms","samplingCount":2,"httpMethod":%q}}}]}}`, server.URL, tc.method)
			config, err := xcore.LoadConfig("json", bytes.NewBufferString(raw))
			if err != nil {
				t.Fatal(err)
			}
			instance, err := xcore.New(config)
			if err != nil {
				t.Fatal(err)
			}
			defer instance.Close()
			if err := instance.Start(); err != nil {
				t.Fatal(err)
			}
			observer := instance.GetFeature(extension.ObservatoryType()).(*multiobs.MultiObservatory)
			deadline := time.Now().Add(2 * time.Second)
			for {
				result, err := observer.GetObservationByTag("group", context.Background())
				if err != nil {
					t.Fatal(err)
				}
				status := result.(*observatory.ObservationResult).Status
				if len(status) > 0 {
					t.Logf("JSON runtime status=%v", status)
					if status[0].Alive != tc.alive {
						t.Errorf("alive=%v want %v", status[0].Alive, tc.alive)
					}
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("no observation in 2s")
				}
				time.Sleep(5 * time.Millisecond)
			}
			mu.Lock()
			defer mu.Unlock()
			if len(methods) == 0 {
				t.Fatal("server got no request")
			}
			for _, method := range methods {
				if method != tc.method {
					t.Errorf("method=%s want %s", method, tc.method)
				}
			}
			t.Logf("endpoint methods=%v", methods)
		})
	}
}
