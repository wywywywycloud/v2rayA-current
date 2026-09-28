package serverObj

import (
	"encoding/json"
	"testing"

	jsoniter "github.com/json-iterator/go"
)

func TestXHTTPImportedRangesUseCoreJSON(t *testing.T) {
	const link = "vless://6df044dd-dae5-4133-8c6a-1cfd6e3a1402@127.0.0.1:443?type=xhttp&security=none&mode=packet-up&scMaxEachPostBytesFrom=17000&scMaxEachPostBytesTo=19000&scMinPostsIntervalFrom=10&scMinPostsIntervalTo=10&scStreamUpServerFrom=20&scStreamUpServerTo=30&xPaddingBytesFrom=100&xPaddingBytesTo=200&xmuxMaxConcurFrom=2&xmuxMaxConcurTo=4"
	server, err := ParseVlessURL(link)
	if err != nil {
		t.Fatal(err)
	}
	config, err := server.Configuration(PriorInfo{Tag: "proxy"})
	if err != nil {
		t.Fatal(err)
	}
	// Exercise the serializer used when writing the actual generated config.
	data, err := jsoniter.Marshal(config.CoreOutbound.StreamSettings.XHTTPSettings)
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]json.RawMessage
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"scMaxEachPostBytes":   `"17000-19000"`,
		"scMinPostsIntervalMs": `10`,
		"scStreamUpServerSecs": `"20-30"`,
		"xPaddingBytes":        `"100-200"`,
		"xmux":                 `{"maxConcurrency":"2-4"}`,
	} {
		if got := string(settings[key]); got != want {
			t.Errorf("%s = %s, want core-compatible %s", key, got, want)
		}
	}
	exported, err := ParseVlessURL(server.ExportToURL())
	if err != nil {
		t.Fatal(err)
	}
	if exported.ScMaxEachPostBytesFrom != 17000 || exported.ScMaxEachPostBytesTo != 19000 {
		t.Fatal("core serialization changed the link range")
	}
}
