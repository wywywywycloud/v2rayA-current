package multiobservatory

import (
	"encoding/hex"
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestObserverProtoCompatibility(t *testing.T) {
	legacy, err := hex.DecodeString("0a016712016118092206646972656374")
	if err != nil {
		t.Fatal(err)
	}
	var config ObserverConfig
	if err := proto.Unmarshal(legacy, &config); err != nil {
		t.Fatal(err)
	}
	if config.Tag != "g" || config.ProbeUrl != "a" || config.ProbeInterval != 9 || len(config.SubjectSelector) != 1 || config.SubjectSelector[0] != "direct" {
		t.Fatalf("legacy=%v", &config)
	}
	if config.Timeout != 0 || config.HttpMethod != "" || config.SamplingCount != 0 || config.Connectivity != "" {
		t.Fatalf("legacy new fields not zero: %v", &config)
	}
	encoded, err := proto.Marshal(&config)
	if err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(encoded) != hex.EncodeToString(legacy) {
		t.Fatalf("legacy encoding changed: %x", encoded)
	}
	d := config.ProtoReflect().Descriptor()
	if string(d.FullName()) != "xray.core.app.observatory.multiobservatory.ObserverConfig" || d.ParentFile().Path() != "app/observatory/multiobservatory/config.proto" {
		t.Fatalf("descriptor=%s path=%s", d.FullName(), d.ParentFile().Path())
	}
	config.Timeout = 1250000
	config.HttpMethod = "GET"
	config.SamplingCount = 3
	config.Connectivity = "http://localhost/"
	encoded, err = proto.Marshal(&config)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ObserverConfig
	if err := proto.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(&config, &decoded) {
		t.Fatalf("roundtrip=%v", &decoded)
	}
}
