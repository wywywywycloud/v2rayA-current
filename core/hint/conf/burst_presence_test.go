package conf

import (
	"encoding/json"
	"testing"
)

func TestMultiPresencePreserved(t *testing.T) {
	for _, tc := range []struct {
		raw     string
		present bool
	}{
		{`{}`, false}, {`{"multiObservatory":null}`, false},
		{`{"multiObservatory":{}}`, true}, {`{"multiObservatory":{"observers":[]}}`, true},
	} {
		var ext extendedJSON
		if err := json.Unmarshal([]byte(tc.raw), &ext); err != nil {
			t.Fatal(err)
		}
		if (ext.MultiObservatory != nil) != tc.present {
			t.Errorf("presence changed: %s", tc.raw)
		}
	}
}
