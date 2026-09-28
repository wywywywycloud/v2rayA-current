package conf

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	multiobs "github.com/v2rayA/v2raya-core/hint/app/observatory/multiobservatory"
	_ "github.com/xtls/xray-core/app/dispatcher"
	_ "github.com/xtls/xray-core/app/proxyman/inbound"
	_ "github.com/xtls/xray-core/app/proxyman/outbound"
	"github.com/xtls/xray-core/common/session"
	xray_core "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/outbound"
	"github.com/xtls/xray-core/features/routing"
	routing_session "github.com/xtls/xray-core/features/routing/session"
	_ "github.com/xtls/xray-core/proxy/blackhole"
	_ "github.com/xtls/xray-core/proxy/freedom"
)

func TestExplicitEmptyMultiObservatoryPreservesFallbackDependency(t *testing.T) {
	for _, raw := range []string{`{"multiObservatory":{}}`, `{"multiObservatory":{"observers":[]}}`} {
		path := filepath.Join(t.TempDir(), "config.json")
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		for _, load := range []func() (*xray_core.Config, error){
			func() (*xray_core.Config, error) { return xray_core.LoadConfig("json", bytes.NewBufferString(raw)) },
			func() (*xray_core.Config, error) {
				return buildConfigFromFiles([]*xray_core.ConfigSource{{Name: path, Format: "json"}})
			},
		} {
			cfg, err := load()
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, app := range cfg.App {
				msg, err := app.GetInstance()
				if err != nil {
					t.Fatal(err)
				}
				if obs, ok := msg.(*multiobs.Config); ok {
					found = true
					if len(obs.Observers) != 0 {
						t.Fatalf("unexpected probe workers: %+v", obs.Observers)
					}
				}
			}
			if !found {
				t.Fatal("explicit empty observatory was discarded")
			}
		}
	}
}

func TestEmptyMultiObservatoryRoundRobinAndRemovedMembers(t *testing.T) {
	cfg, err := xray_core.LoadConfig("json", bytes.NewBufferString(`{
	  "outbounds": [
	    {"tag":"direct","protocol":"freedom"},
	    {"tag":"member-a","protocol":"freedom"},
	    {"tag":"member-b","protocol":"freedom"},
	    {"tag":"block","protocol":"blackhole"}
	  ],
	  "routing": {
	    "balancers":[{"tag":"group","selector":["member-"],"fallbackTag":"block","strategy":{"type":"roundrobin"}}],
	    "rules":[{"type":"field","inboundTag":["client"],"balancerTag":"group"}]
	  },
	  "multiObservatory":{"observers":[]}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	instance, err := xray_core.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer instance.Close()
	if err := instance.Start(); err != nil {
		t.Fatal(err)
	}
	router := instance.GetFeature(routing.RouterType()).(routing.Router)
	pick := func() string {
		t.Helper()
		route, err := router.PickRoute(&routing_session.Context{Inbound: &session.Inbound{Tag: "client"}})
		if err != nil {
			t.Fatal(err)
		}
		return route.GetOutboundTag()
	}
	first, second := pick(), pick()
	if first == second || (first != "member-a" && first != "member-b") || (second != "member-a" && second != "member-b") {
		t.Fatalf("healthy members did not rotate: %s, %s", first, second)
	}
	if pick() != first || pick() != second {
		t.Fatal("roundrobin order changed")
	}
	manager := instance.GetFeature(outbound.ManagerType()).(outbound.Manager)
	if err := manager.RemoveHandler(context.Background(), "member-a"); err != nil {
		t.Fatal(err)
	}
	if got := pick(); got != "member-b" {
		t.Fatalf("one remaining member: got %s", got)
	}
	if err := manager.RemoveHandler(context.Background(), "member-b"); err != nil {
		t.Fatal(err)
	}
	if got := pick(); got != "block" {
		t.Fatalf("empty selector leaked to %s", got)
	}
}

func TestInvalidObserversAreNotAnExplicitEmptyObservatory(t *testing.T) {
	cfg, err := xray_core.LoadConfig("json", bytes.NewBufferString(`{"multiObservatory":{"observers":[{"tag":"invalid"}]}}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, app := range cfg.App {
		msg, err := app.GetInstance()
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := msg.(*multiobs.Config); ok {
			t.Fatal("invalid observers became an empty observatory feature")
		}
	}
}
