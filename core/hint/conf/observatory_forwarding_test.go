package conf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	multiobs "github.com/v2rayA/v2raya-core/hint/app/observatory/multiobservatory"
	"github.com/xtls/xray-core/app/observatory/burst"
	xcore "github.com/xtls/xray-core/core"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestObservatoryPingForwarding(t *testing.T) {
	ping := `"pingConfig":{"destination":"http://127.0.0.1/probe","interval":"13s","timeout":"1750ms","httpMethod":"GET","samplingCount":3,"connectivity":"http://127.0.0.1/connectivity"}`
	for name, document := range map[string]string{
		"settings":    `{"multiObservatory":{"observers":[{"tag":"g","settings":{"subjectSelector":["direct"],` + ping + `}}]}}`,
		"nestedBurst": `{"multiObservatory":{"observers":[{"tag":"g","burstObservatory":{"subjectSelector":["direct"],` + ping + `}}]}}`,
		"globalBurst": `{"burstObservatory":{"subjectSelector":["direct"],` + ping + `}}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(document), 0600); err != nil {
				t.Fatal(err)
			}
			loaders := map[string]func() (*xcore.Config, error){
				"reader": func() (*xcore.Config, error) { return xcore.LoadConfig("json", bytes.NewBufferString(document)) },
				"file": func() (*xcore.Config, error) {
					return buildConfigFromFiles([]*xcore.ConfigSource{{Name: path, Format: "json"}})
				},
			}
			for source, load := range loaders {
				t.Run(source, func(t *testing.T) {
					cfg, err := load()
					if err != nil {
						t.Fatal(err)
					}
					var obs *multiobs.ObserverConfig
					for _, app := range cfg.App {
						v, err := app.GetInstance()
						if err != nil {
							t.Fatal(err)
						}
						t.Logf("app=%T", v)
						if mo, ok := v.(*multiobs.Config); ok {
							obs = mo.Observers[0]
						}
					}
					if obs == nil {
						t.Fatal("missing multi observer")
					}
					if obs.ProbeInterval != int64(13*time.Second) {
						t.Fatalf("interval=%d", obs.ProbeInterval)
					}
					for field, want := range map[string]string{"timeout": "1750000000", "http_method": "GET", "sampling_count": "3", "connectivity": "http://127.0.0.1/connectivity"} {
						d := obs.ProtoReflect().Descriptor().Fields().ByName(protoreflect.Name(field))
						if d == nil {
							t.Errorf("missing forwarded field %s", field)
							continue
						}
						if got := fmt.Sprint(obs.ProtoReflect().Get(d).Interface()); got != want {
							t.Errorf("%s=%s want %s", field, got, want)
						}
					}
				})
			}
		})
	}
}

func TestTopLevelBurstAppInventory(t *testing.T) {
	cfg, err := xcore.LoadConfig("json", bytes.NewBufferString(`{"burstObservatory":{"subjectSelector":["direct"],"pingConfig":{"destination":"http://127.0.0.1/probe","interval":"13s"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	native, multi := 0, 0
	for _, app := range cfg.App {
		v, err := app.GetInstance()
		if err != nil {
			t.Fatal(err)
		}
		switch v.(type) {
		case *burst.Config:
			native++
		case *multiobs.Config:
			multi++
		}
	}
	t.Logf("nativeBurst=%d multiObservatory=%d", native, multi)
}

func TestObservatoryInvalidPingConfig(t *testing.T) {
	for _, ping := range []string{
		`{"timeout":"nonsense"}`, `{"timeout":1000}`, `{"interval":10}`,
		`{"samplingCount":2147483648}`, `{"samplingCount":-1}`,
		`{"timeout":"-1s"}`, `{"httpMethod":"bad method"}`,
		`{"samplingCount":2,"sampling":3}`, `{"unknownSetting":true}`,
	} {
		document := `{"multiObservatory":{"observers":[{"tag":"g","settings":{"subjectSelector":["direct"],"pingConfig":` + ping + `}}]}}`
		if _, err := xcore.LoadConfig("json", bytes.NewBufferString(document)); err == nil {
			t.Errorf("accepted unsupported pingConfig %s", ping)
		} else {
			t.Logf("rejected %s: %v", ping, err)
		}
	}
}

func TestObservatoryMalformedDurationsAllPaths(t *testing.T) {
	for name, raw := range map[string]string{
		"settings":    `{"multiObservatory":{"observers":[{"tag":"group","settings":{"subjectSelector":["direct"],"pingConfig":{"timeout":"oops"}}}]}}`,
		"nestedBurst": `{"multiObservatory":{"observers":[{"tag":"group","burstObservatory":{"subjectSelector":["direct"],"pingConfig":{"timeout":"oops"}}}]}}`,
		"globalBurst": `{"burstObservatory":{"subjectSelector":["direct"],"pingConfig":{"timeout":"oops"}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
				t.Fatal(err)
			}
			for source, load := range map[string]func() error{
				"reader": func() error { _, err := xcore.LoadConfig("json", bytes.NewBufferString(raw)); return err },
				"files": func() error {
					_, err := buildConfigFromFiles([]*xcore.ConfigSource{{Name: path, Format: "json"}})
					return err
				},
				"extend": func() error { _, _, _, err := loadAndExtend(path); return err },
			} {
				t.Run(source, func(t *testing.T) {
					if err := load(); err == nil {
						t.Fatal("malformed timeout accepted")
					} else {
						t.Log(err)
					}
				})
			}
		})
	}
}
