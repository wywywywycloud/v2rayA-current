package conf

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	multiobs "github.com/v2rayA/v2raya-core/hint/app/observatory/multiobservatory"
	xcore "github.com/xtls/xray-core/core"
)

func TestObservatoryLegacyAndPrecedence(t *testing.T) {
	for _, tc := range []struct {
		name, raw, url       string
		interval             time.Duration
		samples              int32
		timeout              time.Duration
		method, connectivity string
	}{
		{name: "flat", raw: `{"tag":"g","subjectSelector":["flat"],"probeURL":"http://flat","probeInterval":"1.5s"}`, url: "http://flat", interval: 1500 * time.Millisecond},
		{name: "legacySettings", raw: `{"tag":"g","settings":{"subjectSelector":["flat"],"probeURL":"http://settings","probeInterval":"12s"}}`, url: "http://settings", interval: 12 * time.Second},
		{name: "emptyPing", raw: `{"tag":"g","subjectSelector":["flat"],"settings":{"pingConfig":{}}}`},
		{name: "nullPing", raw: `{"tag":"g","subjectSelector":["flat"],"settings":{"pingConfig":null}}`},
		{name: "samplingAlias", raw: `{"tag":"g","subjectSelector":["flat"],"settings":{"pingConfig":{"sampling":3}}}`, samples: 3},
		{name: "equalAliases", raw: `{"tag":"g","subjectSelector":["flat"],"settings":{"pingConfig":{"sampling":3,"samplingCount":3}}}`, samples: 3},
		{name: "precedence", raw: `{"tag":"g","subjectSelector":["flat"],"probeURL":"http://flat","probeInterval":"15s","settings":{"subjectSelector":["settings"],"probeURL":"http://settings","probeInterval":"16s","pingConfig":{"destination":"http://ping","interval":"17s","timeout":"75ms","samplingCount":3,"httpMethod":"GET"}},"burstObservatory":{"subjectSelector":["burst"],"pingConfig":{"destination":"http://burst","interval":"18s","timeout":"90ms","samplingCount":4,"httpMethod":"HEAD","connectivity":"http://connectivity"}}}`, url: "http://flat", interval: 15 * time.Second, samples: 3, timeout: 75 * time.Millisecond, method: "GET", connectivity: "http://connectivity"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var entry multiObsEntryJSON
			if err := json.Unmarshal([]byte(tc.raw), &entry); err != nil {
				t.Fatal(err)
			}
			got := normalizeObserverConfig(entry)
			if got == nil || got.Tag != "g" || len(got.SubjectSelector) != 1 || got.SubjectSelector[0] != "flat" {
				t.Fatalf("observer=%v", got)
			}
			if got.ProbeUrl != tc.url || got.ProbeInterval != int64(tc.interval) || got.Timeout != int64(tc.timeout) || got.SamplingCount != tc.samples || got.HttpMethod != tc.method || got.Connectivity != tc.connectivity {
				t.Fatalf("observer=%v", got)
			}
		})
	}
}

func TestObservatoryEmptySemantics(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want int
	}{
		{`{}`, 0},
		{`{"multiObservatory":null}`, 0},
		{`{"multiObservatory":{}}`, 1},
		{`{"multiObservatory":{"observers":[]}}`, 1},
		{`{"multiObservatory":{"observers":[{"tag":"empty"}]}}`, 0},
	} {
		cfg, err := xcore.LoadConfig("json", bytes.NewBufferString(tc.raw))
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, app := range cfg.App {
			value, err := app.GetInstance()
			if err != nil {
				t.Fatal(err)
			}
			if multi, ok := value.(*multiobs.Config); ok {
				count++
				if len(multi.Observers) != 0 {
					t.Errorf("unexpected child observers for %s: %v", tc.raw, multi.Observers)
				}
			}
		}
		if count != tc.want {
			t.Errorf("multi owners=%d want %d for %s", count, tc.want, tc.raw)
		}
	}
}

func TestObservatoryErrorContext(t *testing.T) {
	raw := `{"multiObservatory":{"observers":[{"tag":"group-one","settings":{"pingConfig":{"timeout":"invalid"}}}]}}`
	_, err := xcore.LoadConfig("json", strings.NewReader(raw))
	if err == nil || !strings.Contains(err.Error(), `multiObservatory.observers[0] tag "group-one"`) {
		t.Fatalf("missing observer error context: %v", err)
	}
}
